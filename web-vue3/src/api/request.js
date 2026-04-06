// API utilities — ported from web/src/utils/utils.js
// Mirrors the same interface so server API calls stay identical.

// ─── HTTP helpers ────────────────────────────────────────────────────────────

/** POST with application/x-www-form-urlencoded body.
 *  Returns { data, status } — compatible with previous axios shape. */
export async function request(url, data, headers, ext) {
  const _headers = { 'Content-Type': 'application/x-www-form-urlencoded', ...(headers ?? {}) };
  const body = data ? new URLSearchParams(flattenParams(data)).toString() : undefined;
  const res = await fetch(url, { method: 'POST', headers: _headers, body });
  let responseData;
  if ((ext ?? {}).responseType === 'arraybuffer') {
    responseData = await res.arrayBuffer();
  } else {
    try { responseData = await res.json(); }
    catch { responseData = await res.text(); }
  }
  return { data: responseData, status: res.status };
}

/** Submit a hidden form (triggers browser file download) */
export function post(url, data, ext) {
  const form = document.createElement('form');
  form.action = url;
  form.method = 'POST';
  form.target = '_self';
  Object.assign(form, ext ?? {});
  for (const [key, val] of Object.entries(data ?? {})) {
    if (Array.isArray(val)) {
      for (const v of val) {
        const input = document.createElement('input');
        input.name = key; input.value = v;
        form.appendChild(input);
      }
    } else {
      const input = document.createElement('input');
      input.name = key; input.value = val;
      form.appendChild(input);
    }
  }
  document.body.appendChild(form);
  form.submit();
  form.remove();
}

/** Flatten nested object one level (cpu.usage → cpu_usage) */
export function flattenParams(obj) {
  const out = {};
  for (const [k, v] of Object.entries(obj ?? {})) {
    if (v !== null && v !== undefined) {
      out[k] = String(v);
    }
  }
  return out;
}

// ─── URL helpers ─────────────────────────────────────────────────────────────

export function getBaseURL(ws, suffix) {
  const scheme = location.protocol === 'https:' ? (ws ? 'wss' : 'https') : (ws ? 'ws' : 'http');
  return `${scheme}://${location.host}${location.pathname}${suffix}`;
}

// ─── Misc utilities ───────────────────────────────────────────────────────────

export function waitTime(ms = 100) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

export function formatSize(size) {
  size = isNaN(size) ? 0 : Math.max(Number(size) || 0, 0);
  if (size === 0) return '0 B';
  const k = 1024;
  const i = Math.floor(Math.log(size) / Math.log(k));
  const units = ['B', 'KB', 'MB', 'GB', 'TB', 'PB'];
  return (Math.round((size / Math.pow(k, i)) * 100) / 100) + ' ' + (units[i] ?? 'B');
}

export function tsToTime(ts) {
  if (isNaN(ts)) return 'Unknown';
  const hours = Math.floor(ts / 3600);
  const minutes = Math.floor((ts % 3600) / 60);
  return `${hours}h ${minutes}m`;
}

export function renderUnixEpochToHumanReadable(epoch) {
  if (epoch === undefined) epoch = Date.now();
  if (typeof epoch !== 'number' || isNaN(epoch)) return 'Invalid';
  if (epoch < 1e12) epoch *= 1000;
  const now = Date.now();
  const absDiff = Math.abs(epoch - now);
  const hours = Math.floor(absDiff / 3600000);
  const minutes = Math.floor((absDiff % 3600000) / 60000);
  return `${hours}h ${minutes}m`;
}

export function genRandHex(len) {
  return [...Array(len)].map(() => Math.floor(Math.random() * 16).toString(16)).join('');
}

// ─── Type conversions ─────────────────────────────────────────────────────────

export function hex2ua(hex) {
  if (typeof hex !== 'string') return new Uint8Array([]);
  const list = hex.match(/.{1,2}/g);
  if (!list) return new Uint8Array([]);
  return new Uint8Array(list.map((b) => parseInt(b, 16)));
}

export function ua2hex(buf) {
  return Array.prototype.map.call(buf, (b) => ('00' + b.toString(16)).slice(-2)).join('');
}

export function str2ua(str) {
  return new TextEncoder().encode(str);
}

export function ua2str(buf) {
  return new TextDecoder().decode(buf);
}

export function hex2str(hex) {
  return ua2str(hex2ua(hex));
}

export function str2hex(str) {
  return ua2hex(new TextEncoder().encode(str));
}

// ─── XOR encryption (used for terminal/desktop WebSocket payloads) ────────────

export function encrypt(data, secret) {
  const buf = data instanceof Uint8Array ? new Uint8Array(data) : data;
  for (let i = 0; i < buf.length; i++) buf[i] ^= secret[i % secret.length];
  return buf;
}

export function decrypt(data, secret) {
  const buf = new Uint8Array(data);
  for (let i = 0; i < buf.length; i++) buf[i] ^= secret[i % secret.length];
  return ua2str(buf);
}

/** Strip Go-backend i18n template placeholders: ${i18n|SOME.KEY} → "Some Key" */
export function stripI18n(msg) {
  if (!msg) return msg;
  return msg.replace(/\$\{i18n\|([^}]+)\}/g, (_, key) => {
    const label = key.split('.').pop();
    return label.split('_').map(w => w[0] + w.slice(1).toLowerCase()).join(' ');
  });
}

// ─── Collator-based sort ──────────────────────────────────────────────────────

let _collator;
try {
  _collator = new Intl.Collator(navigator.language, { numeric: true, sensitivity: 'base' });
} catch (_) {}

export const orderCompare = _collator
  ? _collator.compare.bind(_collator)
  : (a, b) => String(a).localeCompare(String(b));
