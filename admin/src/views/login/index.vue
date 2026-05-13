<template>
  <div class="login-container">
    <div class="login-card">
      <div class="login-header">
        <el-icon :size="40" color="#1890ff"><Shop /></el-icon>
        <h2 class="login-title">{{ title }}</h2>
        <p class="login-subtitle">家具销售提成管理系统</p>
      </div>

      <el-form
        ref="loginFormRef"
        :model="loginForm"
        :rules="loginRules"
        class="login-form"
        size="large"
      >
        <el-form-item prop="username">
          <el-input
            v-model="loginForm.username"
            placeholder="请输入用户名"
            prefix-icon="User"
            clearable
          />
        </el-form-item>

        <el-form-item prop="password">
          <el-input
            v-model="loginForm.password"
            type="password"
            placeholder="请输入密码"
            prefix-icon="Lock"
            show-password
            @keyup.enter="handleLogin"
          />
        </el-form-item>

        <el-form-item>
          <el-button
            type="primary"
            :loading="loading"
            class="login-btn"
            @click="handleLogin"
          >
            {{ loading ? '登录中...' : '登 录' }}
          </el-button>
        </el-form-item>
      </el-form>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useUserStore } from '@/store/user'
import { usePermissionStore } from '@/store/permission'
import { phoneRule, passwordRule } from '@/utils/validate'

const title = import.meta.env.VITE_APP_TITLE
const router = useRouter()
const route = useRoute()
const userStore = useUserStore()
const permissionStore = usePermissionStore()

const loginFormRef = ref(null)
const loading = ref(false)

const loginForm = reactive({
  username: 'admin',
  password: '123456',
})

const loginRules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: passwordRule,
}

const handleLogin = () => {
  loginFormRef.value?.validate(async (valid) => {
    if (!valid) return

    loading.value = true
    try {
      // 1. 调用登录接口，获取token
      await userStore.login(loginForm)

      // 2. 获取用户信息、角色、权限
      await userStore.getUserInfo()

      // 3. 根据角色生成动态路由
      const accessRoutes = permissionStore.generateRoutes(userStore.roles)
      accessRoutes.forEach((routeItem) => {
        router.addRoute(routeItem)
      })

      ElMessage.success('登录成功')

      // 4. 跳转到首页或重定向页面
      const redirect = route.query.redirect || '/'
      router.push(redirect)
    } catch (error) {
      ElMessage.error(error.message || '登录失败，请检查账号密码')
    } finally {
      loading.value = false
    }
  })
}
</script>

<style lang="scss" scoped>
.login-container {
  width: 100%;
  height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.login-card {
  width: 420px;
  padding: 40px;
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.15);
}

.login-header {
  text-align: center;
  margin-bottom: 32px;
}

.login-title {
  margin: 12px 0 4px;
  font-size: 24px;
  font-weight: 600;
  color: #303133;
}

.login-subtitle {
  font-size: 14px;
  color: #909399;
}

.login-form {
  .el-form-item {
    margin-bottom: 24px;
  }
}

.login-btn {
  width: 100%;
  background-color: #1890ff;
  border-color: #1890ff;
  font-size: 16px;
  height: 44px;

  &:hover,
  &:focus {
    background-color: #40a9ff;
    border-color: #40a9ff;
  }
}
</style>
