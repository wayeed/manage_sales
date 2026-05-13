<template>
  <el-container class="layout-container">
    <el-header class="layout-header" height="60px">
      <AppHeader />
    </el-header>
    <el-container class="layout-body">
      <AppSidebar />
      <el-main class="layout-main">
        <router-view v-slot="{ Component }">
          <transition name="fade-transform" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup>
import AppHeader from './Header.vue'
import AppSidebar from './Sidebar.vue'
</script>

<style lang="scss" scoped>
.layout-container {
  width: 100%;
  height: 100vh;
  overflow: hidden;
}

.layout-header {
  padding: 0;
  background-color: #fff;
  box-shadow: 0 1px 4px rgba(0, 21, 41, 0.08);
  z-index: 100;
}

.layout-body {
  height: calc(100vh - 60px);
  overflow: hidden;
}

.layout-main {
  background-color: #f0f2f5;
  padding: 16px;
  overflow-y: auto;
}

// 路由切换动画
.fade-transform-enter-active,
.fade-transform-leave-active {
  transition: all 0.3s;
}

.fade-transform-enter-from {
  opacity: 0;
  transform: translateX(-10px);
}

.fade-transform-leave-to {
  opacity: 0;
  transform: translateX(10px);
}
</style>
