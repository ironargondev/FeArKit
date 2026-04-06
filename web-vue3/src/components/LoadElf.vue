<template>
  <el-dialog
    :model-value="open"
    :title="'Load ELF - [' + (device.conn ? device.conn.slice(0,8) : '') + '] ' + device.hostname"
    width="480px"
    draggable
    destroy-on-close
    @close="$emit('cancel')"
  >
    <el-upload
      drag
      :action="uploadUrl"
      :on-success="onSuccess"
      :on-error="onError"
      :show-file-list="false"
      :limit="1"
    >
      <div class="upload-area">
        <i class="fa-solid fa-cloud-arrow-up upload-icon"></i>
        <div>Drop ELF binary here or click to upload</div>
        <div class="upload-hint">ELF will be loaded into memory on the target</div>
      </div>
    </el-upload>
    <template #footer>
      <el-button @click="$emit('cancel')">Close</el-button>
    </template>
  </el-dialog>
</template>

<script>
import { computed } from 'vue';
import { ElMessage } from 'element-plus';

export default {
  name: 'LoadElf',
  props: {
    device: { type: Object, required: true },
    open:   { type: Boolean, default: false },
  },
  emits: ['cancel'],
  setup(props, { emit }) {
    const uploadUrl = computed(() => `./api/device/loadelf?uuid=${props.device.conn}`);

    function onSuccess(res) {
      if (res.code === 0) {
        ElMessage.success('ELF loaded');
        emit('cancel');
      } else {
        ElMessage.error(res.msg || 'Load failed');
      }
    }

    function onError() {
      ElMessage.error('Upload failed');
    }

    return { uploadUrl, onSuccess, onError };
  },
};
</script>

<style scoped>
.upload-area { padding: 20px; text-align: center; color: #666; }
.upload-icon { font-size: 40px; color: #409eff; margin-bottom: 12px; display: block; }
.upload-hint { font-size: 12px; color: #999; margin-top: 6px; }
</style>
