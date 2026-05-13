<template>
  <el-aside :width="appStore.isSidebarCollapsed ? '64px' : '220px'" class="sidebar-container">
    <el-scrollbar>
      <el-menu
        :default-active="activeMenu"
        :collapse="appStore.isSidebarCollapsed"
        :collapse-transition="false"
        router
        background-color="#001529"
        text-color="#ffffffb3"
        active-text-color="#ffffff"
      >
        <template v-for="route in menuList" :key="route.path">
          <!-- 有子菜单（过滤隐藏项后大于1个） -->
          <el-sub-menu
            v-if="route.children && route.children.length > 1"
            :index="route.path"
          >
            <template #title>
              <el-icon v-if="route.meta?.icon">
                <component :is="route.meta.icon" />
              </el-icon>
              <span>{{ route.meta?.title }}</span>
            </template>
            <el-menu-item
              v-for="child in route.children"
              :key="child.path"
              :index="route.path === '/' ? `/${child.path}` : `${route.path}/${child.path}`"
            >
              <el-icon v-if="child.meta?.icon">
                <component :is="child.meta.icon" />
              </el-icon>
              <template #title>{{ child.meta?.title }}</template>
            </el-menu-item>
          </el-sub-menu>

          <!-- 只有一个子菜单或无子菜单 -->
          <el-menu-item
            v-else
            :index="route.children?.[0] 
              ? (route.path === '/' ? `/${route.children[0].path}` : `${route.path}/${route.children[0].path}`)
              : route.path"
          >
            <el-icon v-if="route.meta?.icon || route.children?.[0]?.meta?.icon">
              <component :is="route.meta?.icon || route.children?.[0]?.meta?.icon" />
            </el-icon>
            <template #title>
              {{ route.meta?.title || route.children?.[0]?.meta?.title }}
            </template>
          </el-menu-item>
        </template>
      </el-menu>
    </el-scrollbar>
  </el-aside>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useAppStore } from '@/store/app'
import { usePermissionStore } from '@/store/permission'

const route = useRoute()
const appStore = useAppStore()
const permissionStore = usePermissionStore()

const activeMenu = computed(() => route.path)

/**
 * 从 permission store 获取菜单列表
 * menuList 已经在 generateRoutes 中过滤了 hidden 项
 */
const menuList = computed(() => permissionStore.menuList)
</script>

<style lang="scss" scoped>
.sidebar-container {
  background-color: #001529;
  transition: width 0.3s;
  overflow: hidden;
}

:deep(.el-menu) {
  border-right: none;
}

:deep(.el-menu-item.is-active) {
  background-color: #1890ff !important;
}

:deep(.el-sub-menu .el-menu-item.is-active) {
  background-color: #1890ff !important;
}

:deep(.el-sub-menu__title:hover),
:deep(.el-menu-item:hover) {
  background-color: #000c17 !important;
}
</style>
