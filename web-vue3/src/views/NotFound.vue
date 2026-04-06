<template>
  <div class="not-found">
    <h1>404</h1>
    <p>Page Not Found</p>
    <el-button type="primary" @click="$router.push('/')">Go Home</el-button>
  </div>
</template>

<script>
export default { name: 'NotFound' };
</script>

<style scoped>
.not-found {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 60vh;
  gap: 16px;
  color: #666;
}
.not-found h1 { font-size: 80px; margin: 0; color: #ddd; }
</style>
