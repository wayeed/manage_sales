<template>
  <div class="header-container">
    <div class="header-left">
      <el-icon class="collapse-btn" @click="toggleSidebar">
        <Fold v-if="!appStore.isSidebarCollapsed" />
        <Expand v-else />
      </el-icon>
      <div class="logo-area">
        <el-icon :size="24" color="#1890ff"><Shop /></el-icon>
        <span class="system-title">{{ title }}</span>
      </div>
    </div>
    <div class="header-right">
      <el-dropdown trigger="click" @command="handleCommand">
        <div class="user-info">
          <el-avatar :size="32" class="user-avatar">
            {{ userStore.userName?.charAt(0) || 'U' }}
          </el-avatar>
          <div class="user-detail">
            <span class="user-name">{{ userStore.userName || '用户' }}</span>
            <span v-if="roleDisplay" class="user-role">{{ roleDisplay }}</span>
          </div>
          <el-icon><ArrowDown /></el-icon>
        </div>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item command="profile">
              <el-icon><User /></el-icon>个人中心
            </el-dropdown-item>
            <el-dropdown-item divided command="logout">
              <el-icon><SwitchButton /></el-icon>退出登录
            </el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAppStore } from '@/store/app'
import { useUserStore } from '@/store/user'

const router = useRouter()

const title = import.meta.env.VITE_APP_TITLE
const appStore = useAppStore()
const userStore = useUserStore()

const toggleSidebar = () => {
  appStore.toggleSidebar()
}

/**
 * 显示用户角色名称（取第一个角色或显示"无角色"）
 */
const roleDisplay = computed(() => {
  const roles = userStore.userInfo?.roles
  if (!roles || roles.length === 0) return ''
  return roles.map((r) => r.name || r).join('、')
})

const handleCommand = (command) => {
  if (command === 'profile') {
    // 跳转到个人中心页面
    router.push('/profile')
  } else if (command === 'logout') {
    userStore.logout()
  }
}
</script>

<style lang="scss" scoped>
.header-container {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 100%;
  padding: 0 16px;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.collapse-btn {
  font-size: 20px;
  cursor: pointer;
  color: #606266;
  transition: color 0.3s;

  &:hover {
    color: #1890ff;
  }
}

.logo-area {
  display: flex;
  align-items: center;
  gap: 8px;
}

.system-title {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
  white-space: nowrap;
}

.header-right {
  display: flex;
  align-items: center;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 4px;
  transition: background-color 0.3s;

  &:hover {
    background-color: #f5f7fa;
  }
}

.user-detail {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  line-height: 1.3;
}

.user-avatar {
  background-color: #1890ff;
  color: #fff;
  font-size: 14px;
}

.user-name {
  font-size: 14px;
  color: #303133;
  max-width: 120px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.user-role {
  font-size: 11px;
  color: #909399;
  max-width: 120px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
