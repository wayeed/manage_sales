/**
 * 表单验证规则
 */

// 手机号验证
export const phoneRule = [
  { required: true, message: '请输入手机号', trigger: 'blur' },
  {
    pattern: /^1[3-9]\d{9}$/,
    message: '请输入正确的手机号格式',
    trigger: 'blur',
  },
]

// 邮箱验证
export const emailRule = [
  { required: true, message: '请输入邮箱', trigger: 'blur' },
  {
    type: 'email',
    message: '请输入正确的邮箱格式',
    trigger: 'blur',
  },
]

// 金额验证（非负数，最多两位小数）
export const amountRule = [
  { required: true, message: '请输入金额', trigger: 'blur' },
  {
    pattern: /^(?!0\d)\d+(\.\d{1,2})?$/,
    message: '请输入正确的金额格式（正数，最多两位小数）',
    trigger: 'blur',
  },
]

// 必填验证
export const requiredRule = (message = '此项为必填项') => [
  { required: true, message, trigger: 'blur' },
]

// 必选验证（下拉框）
export const requiredSelectRule = (message = '请选择') => [
  { required: true, message, trigger: 'change' },
]

// 密码验证（至少6位）
export const passwordRule = [
  { required: true, message: '请输入密码', trigger: 'blur' },
  { min: 6, message: '密码长度不能少于6位', trigger: 'blur' },
]

// 确认密码验证
export function confirmPasswordRule(formRef, fieldName = 'password') {
  return [
    { required: true, message: '请再次输入密码', trigger: 'blur' },
    {
      validator: (rule, value, callback) => {
        if (value !== formRef[fieldName]) {
          callback(new Error('两次输入的密码不一致'))
        } else {
          callback()
        }
      },
      trigger: 'blur',
    },
  ]
}

// 最大长度验证
export function maxLengthRule(max, message) {
  return [
    { max, message: message || `最多输入${max}个字符`, trigger: 'blur' },
  ]
}

// 最小长度验证
export function minLengthRule(min, message) {
  return [
    { min, message: message || `最少输入${min}个字符`, trigger: 'blur' },
  ]
}
