import { getLatestVersion } from '../api/app-version'

/**
 * 版本号比较函数
 * @param {string} v1 - 版本号1
 * @param {string} v2 - 版本号2
 * @returns {number} - 1:v1>v2, -1:v1<v2, 0:相等
 */
const compareVersion = (v1, v2) => {
  const arr1 = v1.split('.').map(Number)
  const arr2 = v2.split('.').map(Number)
  const len = Math.max(arr1.length, arr2.length)

  for (let i = 0; i < len; i++) {
    const n1 = arr1[i] || 0
    const n2 = arr2[i] || 0
    if (n1 > n2) return 1
    if (n1 < n2) return -1
  }
  return 0
}

/**
 * 获取当前应用版本
 * 注意：WGT热更新后，需要读取已更新的WGT版本号
 * @returns {string}
 */
const getCurrentVersion = () => {
  // #ifdef APP-PLUS
  // 优先读取已安装的WGT版本号（热更新后的版本）
  const wgtVersion = uni.getStorageSync('wgt_version')
  if (wgtVersion) {
    return wgtVersion
  }
  return plus.runtime.version || '1.0.0'
  // #endif
  // #ifndef APP-PLUS
  return '1.0.0' // H5/小程序默认版本
  // #endif
}

/**
 * 保存WGT版本号（热更新成功后调用）
 * @param {string} version - 版本号
 */
const saveWgtVersion = (version) => {
  uni.setStorageSync('wgt_version', version)
}

/**
 * 获取当前平台
 * @returns {string} - ios/android
 */
const getPlatform = () => {
  // #ifdef APP-PLUS
  const systemInfo = uni.getSystemInfoSync()
  return systemInfo.platform === 'ios' ? 'ios' : 'android'
  // #endif
  return 'android'
}

/**
 * 下载并安装WGT热更新包
 * @param {string} url - 下载地址
 */
const downloadAndInstallWGT = (url) => {
  uni.showLoading({ title: '下载更新包...', mask: true })

  const downloadTask = uni.downloadFile({
    url: url,
    success: (res) => {
      uni.hideLoading()
      if (res.statusCode === 200) {
        // #ifdef APP-PLUS
        plus.runtime.install(res.tempFilePath, {
          force: true
        }, () => {
          // 保存WGT版本号到本地，避免重启前重复提示更新
          const latestVersion = uni.getStorageSync('latest_version_check')
          if (latestVersion) {
            saveWgtVersion(latestVersion)
          }
          uni.showModal({
            title: '更新完成',
            content: '应用已更新，是否立即重启？',
            success: (modalRes) => {
              if (modalRes.confirm) {
                plus.runtime.restart()
              }
            }
          })
        }, (e) => {
          uni.showToast({ title: '安装失败：' + e.message, icon: 'none' })
        })
        // #endif
      } else {
        uni.showToast({ title: '下载失败', icon: 'none' })
      }
    },
    fail: () => {
      uni.hideLoading()
      uni.showToast({ title: '下载失败', icon: 'none' })
    }
  })

  // 显示下载进度
  downloadTask.onProgressUpdate((res) => {
    uni.showLoading({ title: `下载中 ${res.progress}%`, mask: true })
  })
}

/**
 * 整包更新（跳转浏览器下载）
 * @param {string} url - 下载地址
 */
const downloadAndInstallAPP = (url) => {
  // #ifdef APP-PLUS
  plus.runtime.openURL(url)
  // #endif
  // #ifndef APP-PLUS
  uni.showToast({ title: '请在APP中使用此功能', icon: 'none' })
  // #endif
}

/**
 * 执行更新操作
 * @param {object} versionInfo - 版本信息
 */
const doUpdate = (versionInfo) => {
  const { update_type, download_url, version_code, update_content, is_force_update } = versionInfo

  // 构建更新提示内容
  let content = update_content || `发现新版本 ${version_code}，是否立即更新？`

  uni.showModal({
    title: '发现新版本',
    content: content,
    showCancel: !is_force_update, // 强制更新时不显示取消按钮
    confirmText: '立即更新',
    success: (res) => {
      if (res.confirm) {
        if (update_type === 'wgt') {
          // 热更新
          downloadAndInstallWGT(download_url)
        } else {
          // 整包更新
          downloadAndInstallAPP(download_url)
        }
      } else if (is_force_update) {
        // 强制更新时，用户取消则退出应用
        // #ifdef APP-PLUS
        plus.runtime.quit()
        // #endif
      }
    }
  })
}

/**
 * 检查更新
 * @param {object} options - 配置选项
 * @param {boolean} options.silent - 是否静默模式（仅在有新版本时提示）
 * @param {function} options.onUpdate - 有新版本时的回调
 * @param {function} options.onNoUpdate - 无新版本时的回调
 * @param {function} options.onError - 错误回调
 */
export const checkUpdate = async (options = {}) => {
  const { silent = false, onUpdate, onNoUpdate, onError } = options

  try {
    const platform = getPlatform()
    const currentVersion = getCurrentVersion()

    const res = await getLatestVersion(platform)
    const latestVersion = res.data

    if (!latestVersion) {
      if (!silent) {
        uni.showToast({ title: '暂无版本信息', icon: 'none' })
      }
      onNoUpdate && onNoUpdate()
      return
    }

    // 比较版本
    if (compareVersion(latestVersion.version_code, currentVersion) > 0) {
      // 有新版本，先保存版本号（供WGT安装后使用）
      uni.setStorageSync('latest_version_check', latestVersion.version_code)
      onUpdate && onUpdate(latestVersion)
      doUpdate(latestVersion)
    } else {
      // 已是最新
      if (!silent) {
        uni.showToast({ title: '已是最新版本', icon: 'none' })
      }
      onNoUpdate && onNoUpdate()
    }
  } catch (error) {
    console.error('检查更新失败:', error)
    if (!silent) {
      uni.showToast({ title: '检查更新失败', icon: 'none' })
    }
    onError && onError(error)
  }
}

export default {
  checkUpdate,
  compareVersion,
  getCurrentVersion,
  getPlatform,
  saveWgtVersion
}