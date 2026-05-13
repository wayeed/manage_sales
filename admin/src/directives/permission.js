import { useUserStore } from '@/store/user'

/**
 * v-permission 按钮权限指令
 * 用法: v-permission="'admin'" 或 v-permission="['admin', 'manager']"
 */
export const permissionDirective = {
  mounted(el, binding) {
    const { value } = binding
    const userStore = useUserStore()
    const currentRoles = userStore.roles

    if (value) {
      const requiredRoles = Array.isArray(value) ? value : [value]
      const hasPermission = requiredRoles.some((role) => currentRoles.includes(role))

      if (!hasPermission) {
        // 没有权限则移除元素
        el.parentNode && el.parentNode.removeChild(el)
      }
    }
  },
}
