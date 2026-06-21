/**
 * 应用配置文件
 * 集中管理敏感配置项和环境相关参数
 */

// API 配置
const API_CONFIG = {
  // API 域名地址
  BASE_URL: 'https://www.jiaju.cn/api',
}

// APP 下载配置
const APP_CONFIG = {
  // 默认 APP 下载链接（当后端未返回有效链接时使用）
  DEFAULT_DOWNLOAD_URL: 'https://www.jiaju.cn/app/download/jiaju_mall.apk',
}

export {
  API_CONFIG,
  APP_CONFIG,
}
