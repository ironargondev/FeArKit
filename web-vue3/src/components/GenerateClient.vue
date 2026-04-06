<template>
  <el-dialog
    :model-value="visible"
    title="Generate Client"
    width="520px"
    draggable
    destroy-on-close
    @update:model-value="$emit('update:visible', $event)"
  >
    <el-form :model="form" :rules="rules" ref="formRef" label-width="80px">
      <el-form-item label="Host" prop="host">
        <el-input v-model="form.host" />
      </el-form-item>
      <el-form-item label="Port" prop="port">
        <el-input-number v-model="form.port" :min="1" :max="65535" style="width:100%" />
      </el-form-item>
      <el-form-item label="Path" prop="path">
        <el-input v-model="form.path" />
      </el-form-item>
      <el-form-item label="OS / Arch" prop="osArch">
        <el-cascader v-model="form.osArch" :options="prebuilt" style="width:100%" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="$emit('update:visible', false)">Cancel</el-button>
      <el-button type="primary" :loading="submitting" @click="onSubmit">Generate</el-button>
    </template>
  </el-dialog>
</template>

<script>
import { ref, reactive } from 'vue';
import { ElMessage } from 'element-plus';
import { request, post } from '../api/request.js';
import prebuilt from '../config/prebuilt.js';

export default {
  name: 'GenerateClient',
  props: { visible: Boolean },
  emits: ['update:visible'],
  setup(props, { emit }) {
    const formRef    = ref(null);
    const submitting = ref(false);

    const initPort = () => {
      if (location.port) return parseInt(location.port);
      return location.protocol === 'https:' ? 443 : 80;
    };

    const form = reactive({
      host:   location.hostname,
      port:   initPort(),
      path:   location.pathname,
      osArch: ['windows', 'amd64'],
    });

    const rules = {
      host:   [{ required: true, message: 'Required' }],
      port:   [{ required: true, type: 'number' }],
      path:   [{ required: true, message: 'Required' }],
      osArch: [{ required: true, message: 'Required' }],
    };

    async function onSubmit() {
      const valid = await formRef.value?.validate().catch(() => false);
      if (!valid) return;
      submitting.value = true;
      try {
        const payload = {
          host:   form.host,
          port:   String(form.port),
          path:   form.path,
          os:     form.osArch[0],
          arch:   form.osArch[1],
          secure: location.protocol === 'https:' ? 'true' : 'false',
        };
        const basePath = location.origin + location.pathname + 'api/client/';
        const res = await request(basePath + 'check', payload);
        if (res.data.code === 0) {
          post(basePath + 'generate', payload);
          emit('update:visible', false);
        } else {
          ElMessage.warning(res.data.msg || 'Failed to generate client config');
        }
      } finally {
        submitting.value = false;
      }
    }

    return { form, rules, formRef, submitting, prebuilt, onSubmit };
  },
};
</script>
