<template>
  <el-container class="app-layout">
    <el-header class="app-header">
      <span class="app-title">FeArKit</span>
    </el-header>
    <el-main class="app-main">
      <router-view />
    </el-main>
  </el-container>
</template>

<script>
export default { name: 'App' };
</script>

<style>
html, body { margin: 0; padding: 0; height: 100%; }
#app { height: 100%; }
.app-layout { height: 100vh; }
.app-header {
  background: #001529;
  color: #fff;
  display: flex;
  align-items: center;
  padding: 0 24px;
  height: 56px;
  flex-shrink: 0;
}
.app-title { font-size: 20px; font-weight: 600; letter-spacing: 1px; }
.app-main { padding: 16px; overflow: auto; background: #f0f2f5; }
</style>
