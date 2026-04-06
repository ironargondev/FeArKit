<template>
  <el-dialog
    :model-value="open"
    :title="'Device Info - [' + (device.conn ? device.conn.slice(0,8) : '') + '] ' + device.hostname"
    width="780px"
    draggable
    destroy-on-close
    @close="$emit('cancel')"
  >
    <div v-loading="loading" style="min-height:200px">
      <el-alert v-if="error" :title="error" type="error" :closable="false" style="margin-bottom:12px" />

      <template v-if="meta">

        <!-- ── Identity ─────────────────────────────────────────────── -->
        <section-title>Identity</section-title>
        <info-grid :rows="[
          ['Device ID',    meta.id],
          ['Session UUID', device.conn],
          ['Hostname',     meta.hostname],
          ['Username',     meta.username],
          ['PID',          meta.pid],
        ]" />

        <!-- ── OS / Platform ────────────────────────────────────────── -->
        <section-title>Operating System</section-title>
        <info-grid :rows="[
          ['OS',               meta.os],
          ['Architecture',     meta.arch],
          ['Platform',         meta.platform + (meta.platform_version ? ' ' + meta.platform_version : '')],
          ['Platform Family',  meta.platform_family],
          ['Kernel',           meta.kernel_version],
          ['Timezone',         meta.timezone],
          ['Virtualization',   meta.virtualization || '—'],
          ['Boot Time',        bootTimeStr],
        ]" />

        <!-- ── Hardware ─────────────────────────────────────────────── -->
        <section-title>Hardware</section-title>
        <info-grid :rows="[
          ['CPU Model',     meta.cpu?.model],
          ['CPU Cores',     (meta.cpu?.cores?.physical ?? '?') + ' physical / ' + (meta.cpu?.cores?.logical ?? '?') + ' logical'],
          ['RAM Total',     formatSize(meta.ram?.total ?? 0)],
          ['Disk Total',    formatSize(meta.disk?.total ?? 0)],
          ['Primary MAC',   meta.mac || '—'],
        ]" />

        <!-- ── Network ──────────────────────────────────────────────── -->
        <section-title>Network</section-title>
        <info-grid :rows="[
          ['WAN IP', meta.wan || '—'],
          ['LAN IP', meta.lan || '—'],
        ]" />
        <el-table :data="meta.interfaces" size="small" border style="width:100%;margin-top:6px;margin-bottom:14px">
          <el-table-column prop="name" label="Interface" min-width="120" />
          <el-table-column prop="mac"  label="MAC"       min-width="140" />
          <el-table-column label="Addresses" min-width="200">
            <template #default="{ row }">
              <div v-for="a in (row.addrs || [])" :key="a" style="font-family:monospace;font-size:12px">{{ a }}</div>
              <span v-if="!row.addrs?.length" style="color:#aaa">—</span>
            </template>
          </el-table-column>
          <el-table-column label="Flags" width="140">
            <template #default="{ row }">
              <el-tag v-for="f in (row.flags || [])" :key="f" size="small" style="margin:1px">{{ f }}</el-tag>
            </template>
          </el-table-column>
        </el-table>

        <!-- ── Logged-in Users ──────────────────────────────────────── -->
        <section-title>Logged-in Users</section-title>
        <div style="margin-bottom:14px">
          <el-tag v-for="u in (meta.users || [])" :key="u" style="margin:2px">{{ u }}</el-tag>
          <span v-if="!meta.users?.length" style="color:#aaa;font-size:13px">None detected</span>
        </div>

        <!-- ── Client ───────────────────────────────────────────────── -->
        <section-title>Client</section-title>
        <info-grid :rows="[
          ['Commit',        meta.commit || '—'],
          ['Client Uptime', meta.client_uptime ? epochToAge(meta.client_uptime) : '—'],
        ]" />

        <!-- ── Environment ──────────────────────────────────────────── -->
        <section-title>Environment Variables</section-title>
        <el-table :data="envRows" size="small" border style="width:100%;margin-top:6px">
          <el-table-column prop="k" label="Variable" min-width="180" />
          <el-table-column prop="v" label="Value"    min-width="300" show-overflow-tooltip>
            <template #default="{ row }">
              <span style="font-family:monospace;font-size:12px">{{ row.v }}</span>
            </template>
          </el-table-column>
        </el-table>

      </template>
    </div>

    <template #footer>
      <el-button :icon="RefreshRight" @click="fetchInfo">Refresh</el-button>
      <el-button @click="$emit('cancel')">Close</el-button>
    </template>
  </el-dialog>
</template>

<script>
import { ref, computed, onMounted, defineComponent, h } from 'vue';
import { RefreshRight } from '@element-plus/icons-vue';
import { request, formatSize, renderUnixEpochToHumanReadable } from '../api/request.js';

// ── tiny local helper components ──────────────────────────────────────────────

const SectionTitle = defineComponent({
  props: { default: String },
  setup(_, { slots }) {
    return () => h('div', {
      style: 'font-size:13px;font-weight:600;color:var(--el-color-primary);margin:14px 0 4px;border-bottom:1px solid var(--el-border-color-light);padding-bottom:3px'
    }, slots.default?.());
  }
});

const InfoGrid = defineComponent({
  props: { rows: Array },
  setup(props) {
    return () => h('div', { style: 'display:grid;grid-template-columns:180px 1fr;gap:2px 8px;margin-bottom:4px;font-size:13px' },
      props.rows.flatMap(([label, value]) => [
        h('div', { style: 'color:#888;padding:2px 0' }, label),
        h('div', { style: 'font-family:monospace;padding:2px 0;word-break:break-all' }, value ?? '—'),
      ])
    );
  }
});

export default {
  name: 'DeviceInfo',
  components: { SectionTitle, InfoGrid },
  props: {
    device: { type: Object, required: true },
    open:   { type: Boolean, default: false },
  },
  emits: ['cancel'],
  setup(props) {
    const loading = ref(false);
    const error   = ref(null);
    const meta    = ref(null);

    async function fetchInfo() {
      loading.value = true;
      error.value   = null;
      meta.value    = null;
      try {
        const fd = new FormData();
        fd.append('uuid', props.device.conn);
        const res  = await fetch('./api/device/info', { method: 'POST', body: fd });
        const json = await res.json();
        if (json.code !== 0) throw new Error(json.msg || `Request failed`);
        meta.value = json.data?.meta ?? null;
      } catch (e) {
        error.value = e.message;
      } finally {
        loading.value = false;
      }
    }

    const bootTimeStr = computed(() => {
      if (!meta.value?.boot_time) return '—';
      return new Date(meta.value.boot_time * 1000).toLocaleString();
    });

    const envRows = computed(() => {
      if (!meta.value?.env) return [];
      return Object.entries(meta.value.env)
        .sort((a, b) => a[0].localeCompare(b[0]))
        .map(([k, v]) => ({ k, v }));
    });

    function epochToAge(epoch) {
      return renderUnixEpochToHumanReadable(epoch);
    }

    onMounted(fetchInfo);

    return {
      loading, error, meta,
      bootTimeStr, envRows,
      RefreshRight, formatSize, epochToAge,
      fetchInfo,
    };
  },
};
</script>
