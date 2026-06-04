import { get } from './request'

/**
 * 获取最新版本信息
 * @param {string} platform - 平台类型：ios/android
 * @returns {Promise}
 */
export const getLatestVersion = (platform = 'android') => {
  return get('/app-versions/latest', { platform })
}