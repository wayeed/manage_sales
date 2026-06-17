import { createApp } from 'vue'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import zhCn from 'element-plus/dist/locale/zh-cn.mjs'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'

import App from './App.vue'
import router from './router'
import { createPinia } from 'pinia'
import { permissionDirective } from './directives/permission'
import { dialogDragDirective } from './directives/dialog-drag'
import { initCsrfToken } from './api/request'
import './styles/index.scss'

const app = createApp(App)

// 注册所有 Element Plus 图标
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}

// 注册自定义指令
app.directive('permission', permissionDirective)
app.directive('dialog-drag', dialogDragDirective)

app.use(createPinia())
app.use(router)
app.use(ElementPlus, { locale: zhCn })

// 初始化CSRF token后再挂载应用
initCsrfToken().then(() => {
  app.mount('#app')
})
