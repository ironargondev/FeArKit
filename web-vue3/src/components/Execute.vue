<template>
  <el-dialog
    :model-value="open"
    :title="'Run - [' + (device.conn ? device.conn.slice(0,8) : '') + '] ' + device.hostname"
    width="480px"
    draggable
    destroy-on-close
    @close="$emit('cancel')"
  >
    <el-form :model="form" ref="formRef" label-width="90px">
      <el-form-item label="Command" prop="cmd">
        <el-input v-model="form.cmd" placeholder="Command" />
      </el-form-item>
      <el-form-item label="Arguments" prop="args">
        <el-input v-model="form.args" placeholder="Arguments (separated by space)" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="$emit('cancel')">Cancel</el-button>
      <el-button type="primary" :loading="loading" @click="onExecute">Run</el-button>
    </template>
  </el-dialog>
</template>

<script>
import { ref, reactive } from 'vue';
import { ElMessage } from 'element-plus';
import { request } from '../api/request.js';

export default {
  name: 'Execute',
  props: {
    device: { type: Object, required: true },
    open:   { type: Boolean, default: false },
  },
  emits: ['cancel'],
  setup(props, { emit }) {
    const loading = ref(false);
    const formRef = ref(null);
    const form    = reactive({ cmd: '', args: '' });

    async function onExecute() {
      if (!form.cmd.trim()) return;
      loading.value = true;
      try {
        const res = await request('/api/device/exec', {
          uuid: props.device.conn,
          cmd:    form.cmd,
          args:   form.args,
        });
        if (res.data.code === 0) {
          ElMessage.success('Execution success');
          emit('cancel');
        }
      } finally {
        loading.value = false;
      }
    }

    return { form, formRef, loading, onExecute };
  },
};
</script>
