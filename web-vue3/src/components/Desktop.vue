<template>
  <el-dialog
    :model-value="open"
    :title="dialogTitle"
    width="960px"
    draggable
    destroy-on-close
    @opened="onOpened"
    @close="onClose"
  >
    <div style="position:relative;background:#000;line-height:0">
      <canvas ref="canvasEl" style="display:block;max-width:100%" />
    </div>
    <template #footer>
      <el-button :icon="FullScreen" @click="goFullscreen">Fullscreen</el-button>
      <el-button :icon="RefreshRight" @click="requestRefresh">Refresh</el-button>
      <el-button @click="onClose">Close</el-button>
    </template>
  </el-dialog>
</template>

<script>
import { ref, computed, onUnmounted } from 'vue';
import { ElMessage } from 'element-plus';
import { FullScreen, RefreshRight } from '@element-plus/icons-vue';
import { encrypt, decrypt, stripI18n, genRandHex, getBaseURL, hex2ua, ua2hex, str2ua } from '../api/request.js';

// Outgoing frame: [34, 22, 19, 17, 20, 3, len_hi, len_lo, ...encrypted_json]
// Incoming frame: skip first 5 bytes, then byte[0] = op: 0=frame blocks, 2=resize, 3=JSON control
const MAGIC_OUT = [34, 22, 19, 17, 20, 3];

function formatBytes(bytes) {
  if (!bytes) return '0 B';
  const k = 1024, units = ['B', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return (Math.round(bytes / Math.pow(k, i) * 100) / 100) + ' ' + (units[i] ?? 'B');
}

export default {
  name: 'Desktop',
  props: {
    device: { type: Object, required: true },
    open:   { type: Boolean, default: false },
  },
  emits: ['cancel'],
  setup(props, { emit }) {
    const canvasEl   = ref(null);
    const bandwidth  = ref(0);
    const fps        = ref(0);
    const resolution = ref('');

    const dialogTitle = computed(() => {
      const guid = props.device.conn ? props.device.conn.slice(0, 8) : '';
      const parts = ['Desktop', '-', `[${guid}]`, props.device.hostname];
      if (resolution.value) parts.push(resolution.value);
      if (bandwidth.value)  parts.push(formatBytes(bandwidth.value) + '/s');
      if (fps.value)        parts.push(`FPS: ${fps.value}`);
      return parts.join(' ');
    });

    let ws = null, ctx = null, canvas = null;
    let secret = null, connected = false;
    let statsTicker = null;
    let frameCount = 0, byteCount = 0, tickCount = 0;

    function sendJSON(obj) {
      if (!connected || !ws) return;
      const body = encrypt(str2ua(JSON.stringify(obj)), secret);
      const buf  = new Uint8Array(body.length + 8);
      buf.set(MAGIC_OUT, 0);
      buf[6] = (body.length >> 8) & 0xff;
      buf[7] =  body.length       & 0xff;
      buf.set(body, 8);
      ws.send(buf);
    }

    function parseBlocks(ab) {
      ab = ab.slice(5);
      const dv = new DataView(ab);
      const op = dv.getUint8(0);

      if (op === 3) {
        handleControlJSON(ab.slice(1));
        return;
      }
      if (op === 2) {
        const w = dv.getUint16(3, false);
        const h = dv.getUint16(5, false);
        if (w && h && canvas) {
          canvas.width  = w;
          canvas.height = h;
          resolution.value = `${w}×${h}`;
        }
        return;
      }
      if (op === 0) {
        frameCount++;
        byteCount += ab.byteLength;
        let offset = 1;
        while (offset < ab.byteLength) {
          const bl = dv.getUint16(offset,      false);
          const it = dv.getUint16(offset +  2, false);
          const dx = dv.getUint16(offset +  4, false);
          const dy = dv.getUint16(offset +  6, false);
          const bw = dv.getUint16(offset +  8, false);
          const bh = dv.getUint16(offset + 10, false);
          const il = bl - 10;
          offset += 12;
          drawBlock(ab.slice(offset, offset + il), it, dx, dy, bw, bh);
          offset += il;
        }
      }
    }

    function drawBlock(ab, type, dx, dy, bw, bh) {
      if (!ctx) return;
      if (type === 0) {
        ctx.putImageData(new ImageData(new Uint8ClampedArray(ab), bw, bh), dx, dy);
      } else {
        createImageBitmap(new Blob([ab]), 0, 0, bw, bh, {
          premultiplyAlpha: 'none', colorSpaceConversion: 'none',
        }).then((ib) => ctx?.drawImage(ib, 0, 0, bw, bh, dx, dy, bw, bh));
      }
    }

    function handleControlJSON(ab) {
      const text = decrypt(ab, secret);
      let pkt;
      try { pkt = JSON.parse(text); } catch { return; }
      if (pkt.act === 'WARN') {
        ElMessage.warning(stripI18n(pkt.msg) || 'Unknown error');
      } else if (pkt.act === 'QUIT') {
        ElMessage.warning(stripI18n(pkt.msg) || 'Unknown error');
        connected = false;
        ws?.close();
      }
    }

    function connect() {
      if (!canvas) return;
      secret = hex2ua(genRandHex(32));
      const url = getBaseURL(true, `api/device/desktop?uuid=${props.device.conn}&secret=${ua2hex(secret)}`);
      ws = new WebSocket(url);
      ws.binaryType = 'arraybuffer';

      ws.onopen = () => {
        connected = true;
        statsTicker = setInterval(() => {
          bandwidth.value = byteCount;
          fps.value = frameCount;
          byteCount = 0; frameCount = 0;
          tickCount++;
          if (tickCount > 10) { tickCount = 0; sendJSON({ act: 'DESKTOP_PING' }); }
        }, 1000);
      };

      ws.onmessage = (e) => parseBlocks(e.data);

      ws.onclose = () => {
        if (connected) { connected = false; ElMessage.warning('Session disconnected'); }
      };

      ws.onerror = () => {
        if (connected) { connected = false; ElMessage.warning('Session disconnected'); }
        else { ElMessage.warning('Connection failed'); }
      };
    }

    function requestRefresh() {
      if (!connected) {
        ctx = canvas?.getContext('2d', { alpha: false });
        if (ctx) ctx.imageSmoothingEnabled = false;
        connect();
      } else {
        sendJSON({ act: 'DESKTOP_SHOT' });
      }
    }

    function goFullscreen() {
      canvasEl.value?.requestFullscreen().catch(() => ElMessage.warning('Failed to enter fullscreen'));
    }

    function onOpened() {
      canvas = canvasEl.value;
      if (!canvas) return;
      ctx = canvas.getContext('2d', { alpha: false });
      ctx.imageSmoothingEnabled = false;
      connect();
    }

    function onClose() {
      clearInterval(statsTicker);
      statsTicker = null;
      if (connected) { connected = false; ws.onclose = null; ws?.close(); }
      ws = null; ctx = null; canvas = null;
      frameCount = byteCount = tickCount = 0;
      emit('cancel');
    }

    onUnmounted(() => {
      clearInterval(statsTicker);
      if (connected) ws?.close();
    });

    return {
      canvasEl, dialogTitle,
      FullScreen, RefreshRight,
      onOpened, onClose, goFullscreen, requestRefresh,
    };
  },
};
</script>
