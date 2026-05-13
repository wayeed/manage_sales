<template>
  <div class="profile-page">
    <el-row :gutter="20">
      <!-- 左侧：用户信息卡片 -->
      <el-col :span="8">
        <el-card shadow="never">
          <div class="user-card">
            <el-avatar :size="80" class="user-avatar">
              {{ userNameFirstChar }}
            </el-avatar>
            <h3 class="user-name">{{ userStore.userName }}</h3>
            <p class="user-role">{{ roleDisplay }}</p>
            <p class="user-phone">{{ userStore.userPhone }}</p>
          </div>
        </el-card>
      </el-col>

      <!-- 右侧：基本信息 -->
      <el-col :span="16">
        <el-card shadow="never">
          <template #header>
            <span class="card-title">基本信息</span>
          </template>
          <el-descriptions :column="2" border>
            <el-descriptions-item label="用户ID">{{ userInfo.id }}</el-descriptions-item>
            <el-descriptions-item label="用户名">{{ userInfo.username }}</el-descriptions-item>
            <el-descriptions-item label="真实姓名">{{ userInfo.real_name }}</el-descriptions-item>
            <el-descriptions-item label="手机号码">{{ userInfo.phone }}</el-descriptions-item>
            <el-descriptions-item label="所属门店">{{ userInfo.store_name || '-' }}</el-descriptions-item>
            <el-descriptions-item label="账号状态">
              <el-tag :type="userInfo.status === 1 ? 'success' : 'danger'" size="small">
                {{ userInfo.status === 1 ? '正常' : '禁用' }}
              </el-tag>
            </el-descriptions-item>
          </el-descriptions>
        </el-card>

        <!-- 角色权限信息 -->
        <el-card shadow="never" class="mt-4">
          <template #header>
            <span class="card-title">角色与权限</span>
          </template>
          <div class="role-section">
            <h4>我的角色</h4>
            <el-tag
              v-for="role in userInfo.roles"
              :key="role.id"
              class="role-tag"
              type="primary"
            >
              {{ role.role_name || role.role_code }}
            </el-tag>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useUserStore } from '@/store/user'

const userStore = useUserStore()

const userInfo = computed(() => userStore.userInfo || {})

const userNameFirstChar = computed(() => {
  const name = userStore.userName
  return name ? name.charAt(0).toUpperCase() : 'U'
})

const roleDisplay = computed(() => {
  const roles = userStore.userInfo?.roles
  if (!roles || roles.length === 0) return '暂无角色'
  return roles.map(r => r.role_name || r.role_code).join('、')
})
</script>

<style lang="scss" scoped>
.profile-page {
  padding: 20px;

  .user-card {
    text-align: center;
    padding: 20px;

    .user-avatar {
      background-color: #1890ff;
      color: #fff;
      font-size: 32px;
      margin-bottom: 16px;
    }

    .user-name {
      font-size: 20px;
      font-weight: 600;
      color: #303133;
      margin: 0 0 8px;
    }

    .user-role {
      font-size: 14px;
      color: #909399;
      margin: 0 0 8px;
    }

    .user-phone {
      font-size: 13px;
      color: #606266;
      margin: 0;
    }
  }

  .card-title {
    font-size: 16px;
    font-weight: 600;
    color: #303133;
  }

  .mt-4 {
    margin-top: 16px;
  }

  .role-section {
    h4 {
      font-size: 14px;
      color: #606266;
      margin: 0 0 12px;
    }

    .role-tag {
      margin-right: 8px;
      margin-bottom: 8px;
    }
  }
}
</style>
