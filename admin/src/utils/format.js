/**
 * 数据格式化工具函数
 */

/**
 * 金额格式化
 * @param {number} amount - 金额
 * @param {number} decimals - 小数位数，默认2位
 * @returns {string} 格式化后的金额字符串
 */
export function formatAmount(amount, decimals = 2) {
  if (amount === null || amount === undefined || isNaN(amount)) return '0.00'
  return Number(amount).toLocaleString('zh-CN', {
    minimumFractionDigits: decimals,
    maximumFractionDigits: decimals,
  })
}

/**
 * 金额带人民币符号
 * @param {number} amount
 * @returns {string}
 */
export function formatCurrency(amount) {
  return `¥ ${formatAmount(amount)}`
}

/**
 * 日期格式化
 * @param {string|Date} date - 日期
 * @param {string} fmt - 格式模板，默认 'YYYY-MM-DD HH:mm:ss'
 * @returns {string}
 */
export function formatDate(date, fmt = 'YYYY-MM-DD HH:mm:ss') {
  if (!date) return ''
  const d = new Date(date)
  if (isNaN(d.getTime())) return ''

  const map = {
    'YYYY': d.getFullYear(),
    'MM': String(d.getMonth() + 1).padStart(2, '0'),
    'DD': String(d.getDate()).padStart(2, '0'),
    'HH': String(d.getHours()).padStart(2, '0'),
    'mm': String(d.getMinutes()).padStart(2, '0'),
    'ss': String(d.getSeconds()).padStart(2, '0'),
  }

  let result = fmt
  for (const [key, value] of Object.entries(map)) {
    result = result.replace(key, value)
  }
  return result
}

/**
 * 短日期格式化
 * @param {string|Date} date
 * @returns {string}
 */
export function formatShortDate(date) {
  return formatDate(date, 'YYYY-MM-DD')
}

/**
 * 订单状态映射
 * @param {number|string} status
 * @returns {Object} { label, type }
 */
export function getOrderStatus(status) {
  const statusMap = {
    0: { label: '待确认', type: 'info' },
    1: { label: '已确认', type: 'primary' },
    2: { label: '生产中', type: 'warning' },
    3: { label: '已发货', type: '' },
    4: { label: '已完成', type: 'success' },
    5: { label: '已取消', type: 'danger' },
    6: { label: '已退款', type: 'danger' },
  }
  return statusMap[status] || { label: '未知状态', type: 'info' }
}

/**
 * 提成状态映射
 * @param {number|string} status
 * @returns {Object} { label, type }
 */
export function getCommissionStatus(status) {
  const statusMap = {
    0: { label: '待结算', type: 'info' },
    1: { label: '已结算', type: 'success' },
    2: { label: '已发放', type: '' },
    3: { label: '已退回', type: 'danger' },
  }
  return statusMap[status] || { label: '未知状态', type: 'info' }
}

/**
 * 百分比格式化
 * @param {number} value
 * @param {number} decimals
 * @returns {string}
 */
export function formatPercent(value, decimals = 1) {
  if (value === null || value === undefined || isNaN(value)) return '0%'
  return `${Number(value).toFixed(decimals)}%`
}

/**
 * 文件大小格式化
 * @param {number} bytes
 * @returns {string}
 */
export function formatFileSize(bytes) {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return `${(bytes / Math.pow(k, i)).toFixed(2)} ${sizes[i]}`
}
