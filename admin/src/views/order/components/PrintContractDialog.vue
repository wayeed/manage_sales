<template>
  <el-dialog
    v-model="visible"
    title="打印销售合同"
    width="900px"
    :close-on-click-modal="false"
    class="contract-dialog"
  >
    <div class="contract-container" ref="contractRef">
      <!-- 页眉：标题+合同编号 -->
      <div class="contract-header">
        <h1>悦邦家居城销售合同</h1>
        <div class="contract-no">合同编号：{{ contractNo }}</div>
      </div>

      <!-- 客户信息栏 -->
      <div class="customer-info">
        <div class="info-row">
          <div class="info-item">客户名称：{{ order.customer_name || '' }}</div>
          <div class="info-item">联系电话：{{ order.customer_phone || '' }}</div>
          <div class="info-item">订货日期：{{ formatDate(order.created_at) }}</div>
        </div>
        <div class="info-row">
          <div class="info-item wide">送货地址：{{ order.customer_address || '' }}</div>
          <div class="info-item">交货日期：____年____月____日</div>
          <div class="info-item">来源：{{ getSourceLabel(order.source) }}</div>
        </div>
      </div>

      <!-- 商品明细表格 -->
      <table class="goods-table">
        <thead>
          <tr>
            <th style="width: 40px">序号</th>
            <th style="width: 120px">商品名称</th>
            <th style="width: 60px">类别</th>
            <th style="width: 60px">款式</th>
            <th style="width: 120px">规格属性</th>
            <th style="width: 50px">数量</th>
            <th style="width: 40px">单位</th>
            <th style="width: 70px">单价</th>
            <th style="width: 80px">金额</th>
            <th>备注</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(item, index) in displayItems" :key="index">
            <td>{{ index + 1 }}</td>
            <td>{{ item.product_name || '' }}</td>
            <td>{{ item.sku?.product?.category?.category_name || '' }}</td>
            <td>{{ item.sku?.product?.style || '' }}</td>
            <td>{{ formatAttributes(item.sku?.attributes) }}</td>
            <td>{{ item.quantity || 0 }}</td>
            <td>{{ item.sku?.product?.unit || '件' }}</td>
            <td>{{ formatCurrency(item.sale_price) }}</td>
            <td>{{ formatCurrency((item.quantity || 0) * (item.sale_price || 0)) }}</td>
            <td>{{ item.remark || '' }}</td>
          </tr>
          <!-- 备注行 -->
          <tr class="remark-row">
            <td>备注</td>
            <td colspan="9">{{ order.remark || '' }}</td>
          </tr>
        </tbody>
      </table>

      <!-- 金额结算区域 -->
      <div class="amount-section">
        <div class="amount-row">
          <span class="amount-label">合计金额（大写）：</span>
          <span class="amount-chinese">{{ amountInChinese }}</span>
          <span class="amount-small">小写：{{ formatCurrency(order.final_amount) }}元</span>
        </div>
        <div class="amount-row">
          <span>已交合同定金：{{ formatCurrency(order.paid_amount) }}元</span>
          <span style="margin-left: 40px">尚欠合同尾款：{{ formatCurrency(order.remaining_amount) }}元</span>
        </div>
      </div>

      <!-- 支付方式 -->
      <div class="payment-method">
        <span class="method-label">支付方式：</span>
        <span class="method-item">□ 对公转账</span>
        <span class="method-item">□ 微信</span>
        <span class="method-item">□ 支付宝</span>
        <span class="method-item">□ 现金</span>
        <span class="method-item">□ 刷卡</span>
        <span class="method-item">□ 扫码</span>
        <span class="method-item">□ 其他（    ）</span>
      </div>

      <!-- 合同条款 -->
      <div class="contract-terms">
        <div class="terms-title">买方权利及义务</div>
        <div class="terms-content">
          <p>1. 本合同请您仔细核对并审阅，定金金额不低于合同金额的80%，且合同金额尾款需要在送货前结清。</p>
          <p>2. 本合同预付款用于下单采购，付款后因买方原因取消，恕不退还，定制产品不支持退换货。</p>
          <p>3. 大件商品市区免费送货，市区外/步梯房另收运费和搬运费。</p>
          <p>4. 商品保修以说明书为准，人为损坏维修收成本费。</p>
          <p>5. 本店商品一经售出，非质量问题不退不换。</p>
          <p>6. 打包产品、特价品，样品售出不退不换。</p>
          <p>7. 本合同签字/盖章生效，传真件有效。</p>
          <p>8. 本店地址：来宾市桂中大道摩尔城负一楼悦邦家居城1001-1038号停车场旁。</p>
          <p>9. 联系电话：13367827032  19978382727</p>
        </div>
      </div>

      <!-- 客户确认区域 -->
      <div class="customer-confirm">
        <div class="confirm-notice">本合同条款，商家已全部告知客户，客户已知晓！</div>
        <div class="confirm-sign-row">
          <span>客户确认后签字：________________</span>
          <span style="margin-left: 40px">______年____月____日</span>
        </div>
      </div>

      <!-- 签署栏 -->
      <div class="signature-section">
        <div class="sign-item">总经理：________</div>
        <div class="sign-item">销售主管/店长：________</div>
        <div class="sign-item">销售代表（章）：{{ order.salesman?.real_name || '' }}</div>
      </div>
    </div>

    <!-- 操作按钮 -->
    <template #footer>
      <div class="no-print">
        <el-button @click="visible = false">取消</el-button>
        <el-button type="primary" @click="handlePrint" :loading="printLoading">
          打印
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { formatCurrency, formatDate, amountToChinese } from '@/utils/format'

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  },
  order: {
    type: Object,
    default: () => ({})
  },
  items: {
    type: Array,
    default: () => []
  }
})

const emit = defineEmits(['update:modelValue'])

const visible = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val)
})

const contractRef = ref(null)
const printLoading = ref(false)

// 合同编号（使用订单号）
const contractNo = computed(() => {
  return props.order.order_no || ''
})

// 显示的商品列表（最多16行）
const displayItems = computed(() => {
  const items = props.items || []
  const result = [...items]
  // 填充空行至16行
  while (result.length < 16) {
    result.push({
      product_name: '',
      category_name: '',
      style: '',
      spec: '',
      color: '',
      quantity: '',
      unit: '',
      sale_price: '',
      remark: ''
    })
  }
  return result
})

// 金额大写
const amountInChinese = computed(() => {
  const amount = props.order.final_amount || 0
  return amountToChinese(amount)
})

// 来源标签
// 格式化规格属性（JSON -> 字符串）
const formatAttributes = (attrs) => {
  if (!attrs) return ''
  if (typeof attrs === 'string') {
    try {
      attrs = JSON.parse(attrs)
    } catch {
      return attrs
    }
  }
  if (typeof attrs === 'object') {
    return Object.entries(attrs).map(([k, v]) => `${k}:${v}`).join(' ')
  }
  return String(attrs)
}

const getSourceLabel = (source) => {
  const sourceMap = {
    1: '门店',
    2: '老客户介绍',
    3: '网络推广',
    4: '其他'
  }
  return sourceMap[source] || ''
}

// 打印
const handlePrint = () => {
  printLoading.value = true
  setTimeout(() => {
    window.print()
    printLoading.value = false
  }, 300)
}
</script>

<style lang="scss" scoped>
.contract-dialog {
  :deep(.el-dialog__body) {
    padding: 10px 20px;
    max-height: 70vh;
    overflow-y: auto;
  }
}

.contract-container {
  font-family: SimSun, '宋体', serif;
  font-size: 14px;
  line-height: 1.6;
  color: #333;
  background: #fff;
  padding: 10px 0;
}

// 页眉
.contract-header {
  text-align: center;
  margin-bottom: 15px;
  position: relative;

  h1 {
    font-size: 22px;
    font-weight: bold;
    margin: 0;
    letter-spacing: 4px;
  }

  .contract-no {
    position: absolute;
    right: 0;
    top: 50%;
    transform: translateY(-50%);
    color: #c00;
    font-size: 14px;
  }
}

// 客户信息
.customer-info {
  border: 1px solid #333;
  border-bottom: none;
  margin-bottom: 0;

  .info-row {
    display: flex;
    border-bottom: 1px solid #333;

    .info-item {
      flex: 1;
      padding: 6px 10px;
      border-right: 1px solid #333;

      &.wide {
        flex: 2;
      }

      &:last-child {
        border-right: none;
      }
    }
  }
}

// 商品表格
.goods-table {
  width: 100%;
  border-collapse: collapse;
  margin-bottom: 15px;

  th, td {
    border: 1px solid #333;
    padding: 4px 6px;
    text-align: center;
    font-size: 12px;
  }

  th {
    background: #f5f5f5;
    font-weight: bold;
  }

  .remark-row {
    td {
      text-align: left;
      padding: 6px 10px;
    }
  }
}

// 金额区域
.amount-section {
  border: 1px solid #333;
  padding: 10px;
  margin-bottom: 10px;

  .amount-row {
    margin-bottom: 8px;

    &:last-child {
      margin-bottom: 0;
    }
  }

  .amount-label {
    font-weight: bold;
  }

  .amount-chinese {
    margin: 0 20px;
    font-weight: bold;
  }

  .amount-small {
    color: #666;
  }
}

// 支付方式
.payment-method {
  border: 1px solid #333;
  padding: 8px 10px;
  margin-bottom: 10px;

  .method-label {
    font-weight: bold;
    margin-right: 10px;
  }

  .method-item {
    margin-right: 15px;
  }
}

// 合同条款
.contract-terms {
  display: flex;
  border: 1px solid #333;
  margin-bottom: 10px;

  .terms-title {
    width: 30px;
    background: #333;
    color: #fff;
    text-align: center;
    writing-mode: vertical-rl;
    letter-spacing: 4px;
    padding: 10px 0;
    font-weight: bold;
  }

  .terms-content {
    flex: 1;
    padding: 8px 12px;
    font-size: 12px;
    line-height: 1.8;

    p {
      margin: 0;
      margin-bottom: 2px;
    }
  }
}

// 客户确认
.customer-confirm {
  border: 1px solid #333;
  padding: 10px;
  margin-bottom: 10px;

  .confirm-notice {
    background: #333;
    color: #fff;
    padding: 4px 10px;
    font-weight: bold;
    margin-bottom: 10px;
  }

  .confirm-sign-row {
    display: flex;
    justify-content: flex-start;
  }
}

// 签署栏
.signature-section {
  display: flex;
  justify-content: space-between;
  padding: 10px 0;

  .sign-item {
    flex: 1;
  }
}

// 打印样式
@media print {
  .no-print {
    display: none !important;
  }

  :deep(.el-dialog) {
    margin: 0 !important;
    width: 100% !important;
    max-width: 100% !important;
    box-shadow: none !important;
  }

  :deep(.el-dialog__header) {
    display: none !important;
  }

  :deep(.el-dialog__body) {
    padding: 0 !important;
    max-height: none !important;
    overflow: visible !important;
  }

  :deep(.el-dialog__footer) {
    display: none !important;
  }

  :deep(.el-overlay) {
    position: static !important;
    background: none !important;
  }

  .contract-container {
    padding: 0 !important;
  }
}
</style>
