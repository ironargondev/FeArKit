<template>
  <el-dialog
    :model-value="open"
    :title="'Terminal - [' + (device.conn ? device.conn.slice(0,8) : '') + '] ' + device.hostname"
    width="940px"
    draggable
    destroy-on-close
    @opened="onOpened"
    @close="onClose"
  >
    <!-- Special keys bar (Unix only) -->
    <div v-if="device.os !== 'windows'" class="ext-keys">
      <span class="ext-label">Function Keys:</span>
      <el-button
        v-for="k in funcKeys" :key="k.label"
        size="small" text class="key-btn"
        @click="sendInput(k.key)"
      >{{ k.label }}</el-button>
      <span class="ext-label" style="margin-left:12px">Special Keys:</span>
      <el-button
        v-for="k in specialKeys" :key="k.label"
        size="small" text class="key-btn"
        @click="sendInput(k.key)"
      >{{ k.label }}</el-button>
      <el-button
        size="small"
        :type="ctrlActive ? 'primary' : ''"
        class="key-btn"
        style="margin-left:8px"
        @click="ctrlActive = !ctrlActive"
      >Ctrl</el-button>
      <el-button
        v-if="fileSelectActive"
        size="small"
        class="key-btn"
        style="margin-left:8px"
        @click="fileInputRef?.click()"
      >Select File</el-button>
    </div>

    <!-- Terminal container -->
    <div ref="termContainer" class="term-wrap"></div>

    <!-- Hidden file input for Zmodem uploads -->
    <input ref="fileInputRef" type="file" style="display:none" @change="onFileSelected" />

    <template #footer>
      <el-button @click="onClose">Close</el-button>
    </template>
  </el-dialog>
</template>

<script>
import { ref, onUnmounted } from 'vue';
import { Terminal } from '@xterm/xterm';
import { FitAddon } from '@xterm/addon-fit';
import { WebLinksAddon } from '@xterm/addon-web-links';
import { ElMessage } from 'element-plus';
import Zmodem from 'zmodem.js';
import { encrypt, decrypt, genRandHex, getBaseURL, hex2ua, ua2hex, str2hex, hex2str, str2ua, ua2str } from '../api/request.js';

// Frame header: [34, 22, 19, 17, 21, type, len_hi, len_lo, ...body]
// type 0 = raw bytes, type 1 = XOR-encrypted JSON
const MAGIC = [34, 22, 19, 17, 21];

function buildFrame(body, raw) {
  const type = raw ? 0 : 1;
  const buf  = new Uint8Array(body.length + 8);
  buf.set(MAGIC, 0);
  buf[5] = type;
  buf[6] = (body.length >> 8) & 0xff;
  buf[7] =  body.length       & 0xff;
  buf.set(body, 8);
  return buf;
}

function isRawFrame(data) {
  return data[0] === 34 && data[1] === 22 && data[2] === 19 &&
         data[3] === 17 && data[4] === 21 && data[5] === 0;
}

export default {
  name: 'Terminal',
  props: {
    device: { type: Object, required: true },
    open:   { type: Boolean, default: false },
  },
  emits: ['cancel'],
  setup(props, { emit }) {
    const termContainer  = ref(null);
    const ctrlActive     = ref(false);
    const fileInputRef   = ref(null);
    const fileSelectActive = ref(false);

    let term = null, fit = null, ws = null;
    let secret = null, connected = false, ticker = null;
    let winBuffer = { cmd: '', cursor: 0, index: 0, history: [], temp: '', tempCursor: 0 };
    let outputBuffer = '';
    let zsentry = null, zsession = null;

    const funcKeys = [
      { key:'\x1B\x4F\x50', label:'F1'  }, { key:'\x1B\x4F\x51', label:'F2'  },
      { key:'\x1B\x4F\x52', label:'F3'  }, { key:'\x1B\x4F\x53', label:'F4'  },
      { key:'\x1B\x5B\x31\x35\x7E', label:'F5' },
      { key:'\x1B\x5B\x31\x37\x7E', label:'F6' },
      { key:'\x1B\x5B\x31\x38\x7E', label:'F7' },
      { key:'\x1B\x5B\x31\x39\x7E', label:'F8' },
      { key:'\x1B\x5B\x32\x30\x7E', label:'F9' },
      { key:'\x1B\x5B\x32\x31\x7E', label:'F10'},
      { key:'\x1B\x5B\x32\x33\x7E', label:'F11'},
      { key:'\x1B\x5B\x32\x34\x7E', label:'F12'},
    ];

    const specialKeys = [
      { key:'\x1B\x5B\x48', label:'Home' },
      { key:'\x1B\x5B\x46', label:'End'  },
      { key:'\x1B\x5B\x32\x7E', label:'Ins'  },
      { key:'\x1B\x5B\x33\x7E', label:'Del'  },
      { key:'\x1B\x5B\x35\x7E', label:'PgUp' },
      { key:'\x1B\x5B\x36\x7E', label:'PgDn' },
      { key:'\t', label:'Tab' },
      { key:'\x1B', label:'ESC' },
    ];

    function sendRaw(data) {
      if (!connected || !ws) return;
      const chunk = data instanceof Uint8Array ? data : str2ua(data);
      // Split large payloads to respect frame size limits
      if (chunk.length > 65536) {
        for (let off = 0; off < chunk.length; off += 65536) {
          ws.send(buildFrame(chunk.slice(off, off + 65536), true));
        }
      } else {
        ws.send(buildFrame(chunk, true));
      }
    }

    function sendJSON(obj) {
      if (!connected || !ws) return;
      const encoded = encrypt(str2ua(JSON.stringify(obj)), secret);
      ws.send(buildFrame(encoded, false));
    }

    function sendInput(input) {
      if (!connected) { term?.focus?.(); return; }
      if (ctrlActive.value && props.device.os !== 'windows') {
        let code = input.charCodeAt(0);
        if (code >= 0x61 && code <= 0x7A) code -= 0x60;
        else if (code >= 0x40 && code <= 0x5F) code -= 0x40;
        input = String.fromCharCode(code);
        ctrlActive.value = false;
      }
      sendJSON({ act: 'TERMINAL_INPUT', data: { input: str2hex(input) } });
      term?.focus?.();
    }

    // ── Zmodem ───────────────────────────────────────────────────────────────

    function clearZsession() {
      if (zsession) {
        try { zsession._last_header_name = 'ZRINIT'; zsession.close(); } catch {}
        zsession = null;
      }
      fileSelectActive.value = false;
    }

    function initZmodem() {
      if (!Zmodem?.Sentry) return; // library unavailable
      zsentry = new Zmodem.Sentry({
        on_retract: () => {},
        on_detect: (detection) => {
          clearZsession();
          zsession = detection.confirm();
          if (zsession.type === 'send') {
            startZmodemUpload();
          } else {
            startZmodemDownload();
          }
        },
        to_terminal: (data) => {
          term?.write(ua2str(new Uint8Array(data)));
        },
        sender: (data) => {
          sendRaw(new Uint8Array(data));
        },
      });
    }

    function startZmodemUpload() {
      fileSelectActive.value = true;
      term?.write('\r\nZmodem: click "Select File" to upload, or wait will timeout.\r\n');
      // Auto-open file dialog
      setTimeout(() => fileInputRef.value?.click(), 100);
      // Cancel if no file chosen within 10s
      setTimeout(() => {
        if (fileSelectActive.value) {
          term?.write('\r\nZmodem: upload timed out.\r\n');
          clearZsession();
        }
      }, 10000);
    }

    function onFileSelected(event) {
      fileSelectActive.value = false;
      if (!zsession) { event.target.value = null; return; }
      const file = event.target.files[0];
      event.target.value = null;
      if (!file) {
        term?.write('\r\nZmodem: no file selected.\r\n');
        clearZsession();
        return;
      }
      term?.write('\r\n' + file.name + '\tTransferring...\r\n');
      Zmodem.Browser.send_files(zsession, [file], {
        on_offer_response: (f, xfer) => {
          if (!xfer) term?.write(f.name + '\tRejected by receiver.\r\n');
        },
        on_file_complete: (f) => {
          term?.write(f.name + '\tComplete.\r\n');
        },
      }).catch(() => {
        term?.write(file.name + '\tTransfer failed.\r\n');
      }).finally(() => clearZsession());
    }

    function startZmodemDownload() {
      zsession.on('offer', (xfer) => {
        const detail = xfer.get_details();
        if (detail.size > 16 * 1024 * 1024) {
          xfer.skip();
          term?.write('\r\n' + detail.name + '\tFile too large (max 16 MB).\r\n');
          return;
        }
        const chunks = [];
        xfer.on('input', (data) => chunks.push(new Uint8Array(data)));
        xfer.accept().then(() => {
          Zmodem.Browser.save_to_disk(chunks, detail.name);
          term?.write('\r\n' + detail.name + '\tSaved.\r\n');
        }).catch(() => {
          term?.write('\r\n' + detail.name + '\tTransfer failed.\r\n');
        });
      });
      zsession.on('session_end', () => { zsession = null; });
      zsession.start();
    }

    // ── WebSocket & message handling ─────────────────────────────────────────

    function connect() {
      secret = hex2ua(genRandHex(32));
      const url = getBaseURL(true, `api/device/terminal?uuid=${props.device.conn}&secret=${ua2hex(secret)}`);
      ws = new WebSocket(url);
      ws.binaryType = 'arraybuffer';

      ws.onopen = () => {
        connected = true;
        ticker = setInterval(() => { if (connected) sendJSON({ act: 'PING' }); }, 10000);
        sendResize();
      };

      ws.onmessage = (e) => onMessage(e.data);

      ws.onclose = () => {
        if (connected) {
          connected = false;
          secret = hex2ua(genRandHex(32));
          clearZsession();
          term?.write('\r\nSession disconnected\r\n');
        }
      };

      ws.onerror = () => {
        if (connected) {
          connected = false;
          secret = hex2ua(genRandHex(32));
          clearZsession();
          term?.write('\r\nSession disconnected\r\n');
        } else {
          term?.write('\r\nConnection failed\r\n');
        }
      };
    }

    function consumeZsentry(bytes) {
      // zsentry.consume expects a plain Array
      try { zsentry.consume(Array.from(bytes)); } catch (e) { console.error(e); }
    }

    function onMessage(ab) {
      const data = new Uint8Array(ab);
      if (isRawFrame(data)) {
        const bytes = data.slice(8);
        if (zsentry) {
          consumeZsentry(bytes);
        } else {
          writeOutput(ua2str(bytes));
        }
        return;
      }
      const text = decrypt(data, secret);
      let pkt;
      try { pkt = JSON.parse(text); } catch { return; }
      if (!connected) return;
      if (pkt.act === 'TERMINAL_OUTPUT') {
        const bytes = hex2ua(pkt.data?.output ?? '');
        if (zsentry) {
          consumeZsentry(bytes);
        } else {
          writeOutput(ua2str(bytes));
        }
      } else if (pkt.act === 'WARN') {
        ElMessage.warning(pkt.msg || 'Unknown error');
      } else if (pkt.act === 'QUIT') {
        ElMessage.warning(pkt.msg || 'Unknown error');
        ws?.close();
      }
    }

    function writeOutput(text) {
      if (outputBuffer) { text = outputBuffer + text; outputBuffer = ''; }
      term?.write(text);
    }

    function sendResize() {
      if (!connected || !term) return;
      fit?.fit?.();
      sendJSON({ act: 'TERMINAL_RESIZE', data: { cols: term.cols, rows: term.rows } });
    }

    const onWindowResize = () => sendResize();

    // ── Input handlers ────────────────────────────────────────────────────────

    function setupWindowsInput() {
      const buf = winBuffer;
      term.onData((e) => {
        if (!connected) {
          if (e === '\r' || e === '\n' || e === ' ') {
            term.write('\r\nReconnecting...\r\n');
            connect();
          }
          return;
        }
        switch (e) {
          case '\x1B\x5B\x41': // up
            if (buf.index > 0 && buf.index <= buf.history.length) {
              if (buf.index === buf.history.length) { buf.temp = buf.cmd; buf.tempCursor = buf.cursor; }
              buf.index--;
              clearWinTerm(buf);
              buf.cmd = buf.history[buf.index]; buf.cursor = buf.cmd.length;
              term.write(buf.cmd);
            }
            break;
          case '\x1B\x5B\x42': // down
            if (buf.index + 1 < buf.history.length) {
              buf.index++;
              clearWinTerm(buf);
              buf.cmd = buf.history[buf.index]; buf.cursor = buf.cmd.length;
              term.write(buf.cmd);
            } else if (buf.index + 1 <= buf.history.length) {
              clearWinTerm(buf);
              buf.index++;
              buf.cmd = buf.temp; buf.cursor = buf.tempCursor;
              buf.temp = ''; buf.tempCursor = 0;
              term.write(buf.cmd);
            }
            break;
          case '\x1B\x5B\x43': // right
            if (buf.cursor < buf.cmd.length) { term.write('\x1B\x5B\x43'); buf.cursor++; }
            break;
          case '\x1B\x5B\x44': // left
            if (buf.cursor > 0) { term.write('\x1B\x5B\x44'); buf.cursor--; }
            break;
          case '\r': case '\n':
            term.write('\r\n');
            sendJSON({ act: 'TERMINAL_INPUT', data: { input: str2hex(buf.cmd + '\n') } });
            if (buf.cmd.length > 0) buf.history.push(buf.cmd);
            if (buf.history.length > 128) buf.history = buf.history.slice(-128);
            buf.cmd = ''; buf.cursor = 0; buf.index = buf.history.length;
            break;
          case '\x7F': // backspace
            if (buf.cmd.length > 0 && buf.cursor > 0) {
              buf.cursor--;
              buf.cmd = buf.cmd.slice(0, buf.cursor) + buf.cmd.slice(buf.cursor + 1);
              term.write('\b \b');
            }
            break;
          default:
            if (e >= ' ' || e >= '\xA0') {
              buf.cmd = buf.cmd.slice(0, buf.cursor) + e + buf.cmd.slice(buf.cursor);
              buf.cursor += e.length;
              term.write(e);
            }
        }
      });
    }

    function clearWinTerm(buf) {
      term.write('\b'.repeat(buf.cursor));
      term.write(' '.repeat(buf.cmd.length));
      term.write('\b'.repeat(buf.cmd.length));
    }

    function setupUnixInput() {
      initZmodem();
      term.onData((e) => {
        if (!connected) {
          if (e === '\r' || e === ' ') {
            term.write('\r\nReconnecting...\r\n');
            connect();
          }
          return;
        }
        sendInput(e);
      });
    }

    function onOpened() {
      if (!termContainer.value) return;
      fit  = new FitAddon();
      const links = new WebLinksAddon();
      term = new Terminal({
        convertEol: true,
        allowProposedApi: true,
        cursorBlink: true,
        cursorStyle: 'block',
        fontFamily: 'Hack, Menlo, Consolas, monospace',
        fontSize: 15,
        logLevel: 'off',
      });
      term.loadAddon(fit);
      term.loadAddon(links);
      term.open(termContainer.value);
      fit.fit();
      if (props.device.os === 'windows') {
        setupWindowsInput();
      } else {
        setupUnixInput();
      }
      window.addEventListener('resize', onWindowResize);
      connect();
      term.focus();
    }

    function onClose() {
      clearInterval(ticker);
      ticker = null;
      window.removeEventListener('resize', onWindowResize);
      clearZsession();
      zsentry = null;
      if (connected) {
        sendJSON({ act: 'TERMINAL_KILL' });
        ws.onclose = null;
        ws?.close();
      }
      connected = false;
      term?.dispose();
      fit?.dispose();
      term = fit = ws = null;
      ctrlActive.value = false;
      fileSelectActive.value = false;
      emit('cancel');
    }

    onUnmounted(() => {
      clearInterval(ticker);
      window.removeEventListener('resize', onWindowResize);
      clearZsession();
      if (connected) ws?.close();
      term?.dispose();
      fit?.dispose();
    });

    return {
      termContainer, ctrlActive, fileInputRef, fileSelectActive,
      funcKeys, specialKeys,
      sendInput, onOpened, onClose, onFileSelected,
    };
  },
};
</script>

<style scoped>
.term-wrap {
  background: #000;
  padding: 4px 8px;
  min-height: 320px;
  border-radius: 4px;
}
.ext-keys {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 2px;
  margin-bottom: 8px;
  background: #f5f5f5;
  padding: 4px 8px;
  border-radius: 4px;
}
.ext-label { font-size: 12px; color: #888; margin-right: 4px; }
.key-btn { font-size: 12px; padding: 2px 6px; min-width: 36px; }
</style>
