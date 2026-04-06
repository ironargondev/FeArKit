<template>
  <el-dialog
    :model-value="open"
    :title="'File Explorer - [' + (device.conn ? device.conn.slice(0,8) : '') + '] ' + device.hostname"
    width="1050px"
    draggable
    destroy-on-close
    @open="onOpen"
    @close="$emit('cancel')"
  >
    <!-- Breadcrumb -->
    <el-breadcrumb separator="/" style="margin-bottom:8px;font-size:13px">
      <el-breadcrumb-item v-for="(seg, i) in breadcrumbs" :key="i" @click="navigateTo(seg.path)" style="cursor:pointer">
        <i v-if="i===0" class="fa-solid fa-house" style="margin-right:4px"></i>
        {{ seg.label }}
      </el-breadcrumb-item>
    </el-breadcrumb>

    <!-- Toolbar -->
    <div style="display:flex;gap:8px;margin-bottom:8px;flex-wrap:wrap">
      <el-button size="small" :icon="RefreshRight" @click="loadFiles">Refresh</el-button>
      <el-upload
        :action="uploadUrl"
        :show-file-list="false"
        :on-success="onUploadSuccess"
        :on-error="onUploadError"
        :data="{ uuid: device.conn, path: currentPath }"
        style="display:inline-block"
      >
        <el-button size="small" :icon="Upload">Upload</el-button>
      </el-upload>
      <el-button
        v-if="selectedRows.length > 0"
        size="small"
        type="danger"
        @click="deleteSelected"
      >
        Delete ({{ selectedRows.length }})
      </el-button>
    </div>

    <!-- Text editor overlay -->
    <div v-if="editingFile" style="margin-bottom:8px">
      <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:4px">
        <span style="font-size:13px;color:#666">Editing: {{ editingFile }}</span>
        <div style="display:flex;gap:8px">
          <el-button size="small" type="primary" @click="saveFile">Save</el-button>
          <el-button size="small" @click="closeEditor">Cancel</el-button>
        </div>
      </div>
      <div ref="editorEl" class="cm-editor-wrap"></div>
    </div>

    <!-- Image preview -->
    <div v-if="previewUrl" style="text-align:center;margin-bottom:8px">
      <img :src="previewUrl" style="max-width:100%;max-height:400px" alt="preview" />
      <div><el-button size="small" @click="previewUrl=''">Close Preview</el-button></div>
    </div>

    <!-- File table -->
    <el-table
      v-show="!editingFile && !previewUrl"
      v-loading="loading"
      :data="displayedFiles"
      height="360"
      size="small"
      row-key="name"
      @selection-change="selectedRows = $event"
      @row-dblclick="onRowDblClick"
    >
      <el-table-column type="selection" width="40" />
      <el-table-column label="Name" min-width="200" show-overflow-tooltip>
        <template #default="{ row }">
          <i :class="fileIcon(row)" style="margin-right:6px"></i>
          <span>{{ row.name }}</span>
        </template>
      </el-table-column>
      <el-table-column label="Size" width="90" prop="size"
        :formatter="(row) => row.type === 0 || row.type === 2 ? formatSize(row.size) : '-'" />
      <el-table-column label="Date Modified" width="160" prop="time"
        :formatter="(row) => row.type === 0 ? formatTs(row.time) : '-'" />
      <el-table-column width="240">
        <template #default="{ row }">
          <div style="display:flex;gap:4px">
            <el-button v-if="row.type === 0" size="small" @click="downloadFile(row)">Download</el-button>
            <el-button v-if="row.type === 0" size="small" @click="editFile(row)">Edit</el-button>
            <el-button size="small" type="danger" @click="confirmDelete(row)">Delete</el-button>
          </div>
        </template>
      </el-table-column>
    </el-table>

    <template #footer>
      <el-button @click="$emit('cancel')">Close</el-button>
    </template>
  </el-dialog>
</template>

<script>
import { ref, computed, watch, nextTick, onMounted } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { RefreshRight, Upload } from '@element-plus/icons-vue';
import { EditorView, basicSetup } from 'codemirror';
import { javascript } from '@codemirror/lang-javascript';
import { python } from '@codemirror/lang-python';
import { html } from '@codemirror/lang-html';
import { css } from '@codemirror/lang-css';
import { json } from '@codemirror/lang-json';
import { markdown } from '@codemirror/lang-markdown';
import { request, formatSize, stripI18n } from '../api/request.js';

const EXT_LANG = {
  js: () => javascript(), mjs: () => javascript(),
  ts: () => javascript({ typescript: true }),
  jsx: () => javascript({ jsx: true }),
  tsx: () => javascript({ typescript: true, jsx: true }),
  py: () => python(),
  html: () => html(), htm: () => html(),
  css: () => css(),
  json: () => json(),
  md: () => markdown(), markdown: () => markdown(),
};

export default {
  name: 'FileExplorer',
  props: {
    device: { type: Object, required: true },
    open:   { type: Boolean, default: false },
  },
  emits: ['cancel'],
  setup(props) {
    const loading      = ref(false);
    const files        = ref([]);
    const currentPath  = ref('/');
    const selectedRows = ref([]);
    const editingFile  = ref('');
    const previewUrl   = ref('');
    const editorEl     = ref(null);
    const isWindows    = computed(() => props.device.os === 'windows');

    const uploadUrl = './api/device/file/upload';

    let cmEditor = null;
    let pendingContent = '';

    // Create/destroy CodeMirror when editingFile changes (flush:post = after DOM update)
    watch(editingFile, async (val) => {
      if (!val) {
        cmEditor?.destroy();
        cmEditor = null;
        return;
      }
      await nextTick();
      if (!editorEl.value) return;
      const ext = val.split('.').pop().toLowerCase();
      const langFn = EXT_LANG[ext];
      const extensions = [basicSetup];
      if (langFn) extensions.push(langFn());
      cmEditor = new EditorView({ doc: pendingContent, extensions, parent: editorEl.value });
    }, { flush: 'post' });

    const breadcrumbs = computed(() => {
      const sep  = isWindows.value ? '\\' : '/';
      const path = currentPath.value;
      if (!path || path === sep) return [{ label: isWindows.value ? 'This PC' : '/', path: sep }];
      const parts = path.split(sep).filter(Boolean);
      const crumbs = [{ label: isWindows.value ? 'This PC' : '/', path: sep }];
      let acc = isWindows.value ? '' : '/';
      for (const p of parts) {
        acc = isWindows.value ? (acc ? acc + '\\' + p : p) : acc + p + '/';
        // Ensure bare Windows drive letters (e.g. "C:") resolve to the drive root ("C:\")
        const crumbPath = (isWindows.value && /^[A-Za-z]:$/.test(acc)) ? acc + '\\' : acc;
        crumbs.push({ label: p, path: crumbPath });
      }
      return crumbs;
    });

    async function loadFiles() {
      loading.value = true;
      files.value   = [];
      try {
        const res  = await request('/api/device/file/list', { uuid: props.device.conn, path: currentPath.value });
        const data = res.data;
        if (data.code === 0) {
          files.value = data.data?.files ?? [];
        } else {
          ElMessage.warning(stripI18n(data.msg) || 'Request failed');
        }
      } finally {
        loading.value = false;
      }
    }

    function onOpen() {
      currentPath.value = isWindows.value ? '' : '/';
      editingFile.value = '';
      previewUrl.value  = '';
      loadFiles();
    }

    function navigateTo(path) {
      currentPath.value = path;
      closeEditor();
      previewUrl.value  = '';
      loadFiles();
    }

    function onRowDblClick(row) {
      if (row.type === 1 || row.type === 2) {
        if (row.type === 2) {
          // Windows volumes: navigate directly to the mountpoint (e.g. "C:\")
          navigateTo(row.name);
          return;
        }
        const sep  = isWindows.value ? '\\' : '/';
        const base = currentPath.value.endsWith(sep) ? currentPath.value : currentPath.value + sep;
        navigateTo(base + row.name);
      }
    }

    function filePath(name) {
      const sep  = isWindows.value ? '\\' : '/';
      const base = currentPath.value.endsWith(sep) ? currentPath.value : currentPath.value + sep;
      return base + name;
    }

    async function downloadFile(row) {
      const fpath = filePath(row.name);
      let res;
      try {
        const params = new URLSearchParams();
        params.append('uuid', props.device.conn);
        params.append('files', fpath);
        res = await fetch('./api/device/file/get', {
          method: 'POST',
          headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
          body: params.toString(),
        });
      } catch {
        ElMessage.error('Download failed: network error');
        return;
      }
      if (!res.ok) {
        let msg = 'Download failed';
        try { const j = await res.json(); msg = stripI18n(j.msg) || msg; } catch {}
        ElMessage.error(msg);
        return;
      }
      const cd = res.headers.get('Content-Disposition');
      let filename = row.name;
      if (cd) {
        const m = cd.match(/filename\*=UTF-8''([^;]+)/i) || cd.match(/filename="([^"]+)"/i);
        if (m) filename = decodeURIComponent(m[1]);
      }
      const blob = await res.blob();
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = filename;
      document.body.appendChild(a);
      a.click();
      a.remove();
      URL.revokeObjectURL(url);
    }

    async function editFile(row) {
      try {
        const res = await request('/api/device/file/text', { uuid: props.device.conn, file: filePath(row.name) }, {}, { responseType: 'arraybuffer' });
        pendingContent = new TextDecoder().decode(new Uint8Array(res.data));
        editingFile.value = filePath(row.name);
      } catch {
        ElMessage.error('Request failed');
      }
    }

    function closeEditor() {
      cmEditor?.destroy();
      cmEditor = null;
      editingFile.value = '';
      pendingContent = '';
    }

    async function saveFile() {
      const content = cmEditor ? cmEditor.state.doc.toString() : pendingContent;
      const blob = new Blob([new TextEncoder().encode(content)]);
      const sep   = isWindows.value ? '\\' : '/';
      const parts = editingFile.value.replace(/\\/g, '/').split('/');
      const filename = parts.pop();
      const dir = parts.join(sep) || sep;
      const fd   = new FormData();
      fd.append('uuid', props.device.conn);
      fd.append('path',   dir);
      fd.append('file',   blob, filename);
      try {
        const res  = await fetch('./api/device/file/upload', { method: 'POST', body: fd });
        const json = await res.json();
        if (json.code === 0) {
          ElMessage.success('File saved successfully');
          closeEditor();
        } else {
          ElMessage.error(json.msg || 'Failed to save file');
        }
      } catch {
        ElMessage.error('Failed to save file');
      }
    }

    async function confirmDelete(row) {
      const kind = row.type === 0 ? 'file' : 'folder';
      await ElMessageBox.confirm(`Are you sure to delete this ${kind}?`, {
        confirmButtonText: 'Yes', cancelButtonText: 'No', type: 'warning',
      });
      const res = await request('/api/device/file/remove', { uuid: props.device.conn, files: [filePath(row.name)] });
      if (res.data.code === 0) {
        ElMessage.success('File or folder deleted');
        await loadFiles();
      }
    }

    async function deleteSelected() {
      await ElMessageBox.confirm('Are you sure to delete these items?', {
        confirmButtonText: 'Yes', cancelButtonText: 'No', type: 'warning',
      });
      for (const row of selectedRows.value) {
        await request('/api/device/file/remove', { uuid: props.device.conn, files: [filePath(row.name)] });
      }
      ElMessage.success('File or folder deleted');
      await loadFiles();
    }

    function onUploadSuccess(res) {
      if (res.code === 0) {
        ElMessage.success('Upload success');
        loadFiles();
      } else {
        ElMessage.error(res.msg || 'Upload failed');
      }
    }

    function onUploadError() {
      ElMessage.error('Upload failed');
    }

    function fileIcon(row) {
      if (row.type === 1) return 'fa-regular fa-folder';
      if (row.type === 2) return 'fa-solid fa-hard-drive';
      const ext = (row.name ?? '').split('.').pop().toLowerCase();
      if (['png','jpg','jpeg','gif','bmp','webp','ico','svg'].includes(ext)) return 'fa-regular fa-image';
      return 'fa-regular fa-file';
    }

    function formatTs(epoch) {
      if (!epoch) return '-';
      return new Date(epoch < 1e12 ? epoch * 1000 : epoch).toLocaleString();
    }

    const isDriveRoot = computed(() => {
      if (!isWindows.value) return false;
      return /^[A-Za-z]:[\\\/]?$/.test(currentPath.value);
    });

    const displayedFiles = computed(() =>
      isDriveRoot.value ? files.value.filter(f => f.type === 1) : files.value
    );

    onMounted(onOpen);
    return {
      loading, files, displayedFiles, currentPath, breadcrumbs, selectedRows,
      editingFile, previewUrl, editorEl, uploadUrl,
      RefreshRight, Upload, formatSize,
      onOpen, navigateTo, onRowDblClick, loadFiles,
      downloadFile, editFile, saveFile, closeEditor,
      confirmDelete, deleteSelected,
      onUploadSuccess, onUploadError, fileIcon, formatTs,
    };
  },
};
</script>

<style scoped>
.cm-editor-wrap {
  border: 1px solid #ddd;
  border-radius: 4px;
  height: 360px;
  overflow: auto;
  font-size: 13px;
}
.cm-editor-wrap :deep(.cm-editor) {
  height: 100%;
}
</style>
