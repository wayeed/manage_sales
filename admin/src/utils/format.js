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
 * @param {number} value - 支持小数形式（如 0.15）或百分比形式（如 15）
 * @param {number} decimals - 小数位数，默认1位
 * @returns {string}
 */
export function formatPercent(value, decimals = 0) {
  if (value === null || value === undefined || isNaN(value)) return '0%'
  const num = Number(value)
  // 如果值小于1（表示小数形式如0.15），乘以100转换为百分比
  if (num <= 1) {
    return `${(num * 100).toFixed(decimals)}%`
  }
  return `${num.toFixed(decimals)}%`
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

// 别名导出，兼容不同命名习惯
export const formatMoney = formatAmount
export const formatDateTime = formatDate

/**
 * 金额转大写
 * @param {number} amount - 金额
 * @returns {string} 大写金额
 */
export function amountToChinese(amount) {
  if (amount === null || amount === undefined || isNaN(amount)) return '零元整'
  
  const num = Math.abs(Number(amount))
  if (num === 0) return '零元整'
  
  const digits = ['零', '壹', '贰', '叁', '肆', '伍', '陆', '柒', '捌', '玖']
  const units = ['', '拾', '佰', '仟']
  const bigUnits = ['', '万', '亿']
  
  let result = ''
  let integerPart = Math.floor(num)
  let decimalPart = Math.round((num - integerPart) * 100)
  
  // 处理整数部分
  if (integerPart > 0) {
    let unitIndex = 0
    let zeroFlag = false
    
    while (integerPart > 0) {
      let section = integerPart % 10000
      let sectionStr = ''
      let sectionZeroFlag = false
      
      for (let i = 0; i < 4 && section > 0; i++) {
        let digit = section % 10
        if (digit === 0) {
          sectionZeroFlag = true
        } else {
          if (sectionZeroFlag) {
            sectionStr = '零' + sectionStr
            sectionZeroFlag = false
          }
          sectionStr = digits[digit] + units[i] + sectionStr
        }
        section = Math.floor(section / 10)
      }
      
      if (sectionStr) {
        sectionStr += bigUnits[unitIndex]
        if (zeroFlag && !sectionStr.startsWith('零')) {
          sectionStr = '零' + sectionStr
        }
        result = sectionStr + result
        zeroFlag = false
      } else {
        zeroFlag = true
      }
      
      integerPart = Math.floor(integerPart / 10000)
      unitIndex++
    }
    
    result += '元'
  }
  
  // 处理小数部分
  if (decimalPart > 0) {
    const jiao = Math.floor(decimalPart / 10)
    const fen = decimalPart % 10
    
    if (jiao > 0) {
      result += digits[jiao] + '角'
    }
    if (fen > 0) {
      result += digits[fen] + '分'
    }
  } else {
    result += '整'
  }
  
  return result
}
