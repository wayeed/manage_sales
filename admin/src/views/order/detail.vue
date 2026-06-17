<template>
  <div class="order-detail">
    <el-card v-loading="loading" shadow="never">
      <template #header>
        <div class="card-header">
          <span class="title">订单详情</span>
          <el-button icon="ArrowLeft" @click="handleBack">返回列表</el-button>
        </div>
      </template>

      <template v-if="detail">
        <!-- 基本信息 -->
        <el-card class="info-card" shadow="never">
          <template #header>
            <span class="section-title">基本信息</span>
          </template>
          <el-descriptions :column="3" border>
            <el-descriptions-item label="订单号">{{ detail.order_no }}</el-descriptions-item>
            <el-descriptions-item label="订单状态">
              <el-tag :type="getStatusTag(detail.order_status)" size="small">
                {{ getStatusLabel(detail.order_status) }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="订单类型">
              <el-tag :type="getOrderTypeTag(detail.order_type)" size="small">
                {{ getOrderTypeLabel(detail.order_type) }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="创建时间">{{ formatDate(detail.created_at) }}</el-descriptions-item>
            <el-descriptions-item label="业务员">{{ detail.salesman?.real_name || '-' }}</el-descriptions-item>
            <el-descriptions-item label="客户名称">{{ detail.customer_name || '-' }}</el-descriptions-item>
            <el-descriptions-item label="客户手机">{{ detail.customer_phone || '-' }}</el-descriptions-item>
            <el-descriptions-item label="客户等级">{{ detail.customer?.level || '-' }}</el-descriptions-item>
            <el-descriptions-item label="备注" :span="3">{{ detail.remark || '-' }}</el-descriptions-item>
          </el-descriptions>
        </el-card>

        <!-- 金额信息 -->
        <el-card class="info-card" shadow="never">
          <template #header>
            <span class="section-title">金额信息</span>
          </template>
          <el-descriptions :column="4" border>
            <el-descriptions-item label="挂牌总价">
              <span class="price">{{ formatCurrency(detail.total_list_price) }}</span>
            </el-descriptions-item>
            <el-descriptions-item label="实际销售价">
              <span class="price">{{ formatCurrency(detail.total_sale_price) }}</span>
            </el-descriptions-item>
            <el-descriptions-item label="折扣金额">
              <span class="discount">-{{ formatCurrency(detail.discount_amount) }}</span>
            </el-descriptions-item>
            <el-descriptions-item label="最终成交价">
              <span class="price">{{ formatCurrency(detail.final_amount) }}</span>
            </el-descriptions-item>
            <el-descriptions-item label="已付金额">
              <span class="profit">{{ formatCurrency(detail.paid_amount) }}</span>
            </el-descriptions-item>
            <el-descriptions-item label="尾款金额">
              <span :class="remainingRate > 20 ? 'loss' : 'price'">
                {{ formatCurrency(detail.remaining_amount) }}
              </span>
            </el-descriptions-item>
            <el-descriptions-item label="尾款比例">
              <span :class="remainingRate > 20 ? 'loss' : 'price'">
                {{ remainingRate.toFixed(1) }}%
              </span>
            </el-descriptions-item>
            <el-descriptions-item label="商品成本">
              <span class="cost">{{ formatCurrency(detail.total_cost) }}</span>
            </el-descriptions-item>
            <el-descriptions-item label="礼品成本">
              <span class="cost">{{ formatCurrency(detail.gift_cost) }}</span>
            </el-descriptions-item>
            <el-descriptions-item label="实际利润">
              <span :class="detail.actual_profit >= 0 ? 'profit' : 'loss'">
                {{ formatCurrency(detail.actual_profit) }}
              </span>
            </el-descriptions-item>
            <el-descriptions-item label="利润率">
              <span :class="detail.profit_rate >= 0 ? 'profit' : 'loss'">
                {{ formatPercent(detail.profit_rate) }}
              </span>
            </el-descriptions-item>
            <el-descriptions-item v-if="detail.is_returned === 1" label="退货信息" :span="3">
              <el-tag type="danger" size="small" style="margin-right: 8px">已退货</el-tag>
              <span>退货金额：{{ formatCurrency(detail.return_amount) }}</span>
              <span style="margin-left: 16px" v-if="detail.return_time">退货时间：{{ formatDate(detail.return_time) }}</span>
            </el-descriptions-item>
          </el-descriptions>
        </el-card>

        <!-- 商品明细 -->
        <el-card class="info-card" shadow="never">
          <template #header>
            <div class="section-header">
              <span class="section-title">商品明细</span>
              <el-button
                v-if="Number(detail.order_status) === 1"
                type="success"
                size="small"
                icon="Document"
                @click="handlePrintContract"
              >
                打印销售合同
              </el-button>
            </div>
          </template>
          <el-table
            :data="detail.items || []"
            border
            stripe
            style="width: 100%"
            show-summary
            :summary-method="getItemSummary"
            :row-class-name="getItemRowClass"
          >
            <el-table-column label="商品名称" width="400">
              <template #default="{ row }">
                <span :class="{ 'item-removed-text': row.item_status === 2 }">{{ row.sku?.sku_name }}</span>
                <el-tag v-if="row.item_status === 1" type="success" size="small" style="margin-left: 8px">新增</el-tag>
                <el-tag v-if="row.item_status === 2" type="danger" size="small" style="margin-left: 8px">移除</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="品牌款式" min-width="150" align="center">
              <template #default="{ row }">
                {{ row.sku?.product?.brand || '' }} / {{ row.sku?.product?.style || '' }}
              </template>
            </el-table-column>
            <el-table-column label="SKU" width="120" align="center">
              <template #default="{ row }">
                {{ row.sku?.sku_code || '' }}
              </template>
            </el-table-column>
            <el-table-column label="单位" width="60" align="center">
              <template #default="{ row }">
                {{ row.sku?.product?.unit || '件' }}
              </template>
            </el-table-column>
            <el-table-column prop="quantity" label="数量" width="70" align="center" />
            <el-table-column label="挂牌价" width="120" align="right">
              <template #default="{ row }">
                <span>{{ formatCurrency(row.list_price) }}</span>
              </template>
            </el-table-column>
            <el-table-column label="销售价" width="120" align="right">
              <template #default="{ row }">
                <span class="price">{{ formatCurrency(row.sale_price) }}</span>
              </template>
            </el-table-column>
            <el-table-column label="折扣率" width="90" align="center">
              <template #default="{ row }">
                <span>{{ formatPercent(row.discount_rate) }}</span>
              </template>
            </el-table-column>
            <el-table-column label="单位成本" width="120" align="right">
              <template #default="{ row }">
                <span class="cost">{{ formatCurrency(row.unit_cost) }}</span>
              </template>
            </el-table-column>
            <el-table-column label="成本合计" width="120" align="right">
              <template #default="{ row }">
                <span class="cost">{{ formatCurrency((row.quantity || 0) * (row.unit_cost || 0)) }}</span>
              </template>
            </el-table-column>
          </el-table>
        </el-card>

        <!-- 礼品信息 -->
        <el-card v-if="detail.gifts && detail.gifts.length > 0" class="info-card" shadow="never">
          <template #header>
            <span class="section-title">礼品信息</span>
          </template>
          <el-table
            :data="detail.gifts"
            border
            stripe
            style="width: 100%"
            show-summary
            :summary-method="getGiftSummary"
          >
            <el-table-column prop="gift_name" label="礼品名称" min-width="160" />
            <el-table-column prop="quantity" label="数量" width="100" align="center" />
            <el-table-column label="成本价" width="120" align="right">
              <template #default="{ row }">
                <span class="cost">{{ formatCurrency(row.cost_price) }}</span>
              </template>
            </el-table-column>
            <el-table-column label="成本合计" width="120" align="right">
              <template #default="{ row }">
                <span class="cost">{{ formatCurrency((row.quantity || 0) * (row.cost_price || 0)) }}</span>
              </template>
            </el-table-column>
          </el-table>
        </el-card>

        <!-- 回款记录 -->
        <el-card class="info-card" shadow="never">
          <template #header>
            <div class="section-header">
              <span class="section-title">回款记录</span>
              <el-button
                v-if="Number(detail.order_status) === 1 && detail.payment_status !== 2"
                type="primary"
                size="small"
                icon="Plus"
                @click="handleAddPayment"
              >
                回款录入
              </el-button>
            </div>
          </template>
          <el-table
            :data="detail.payments || []"
            border
            stripe
            style="width: 100%"
          >
            <el-table-column prop="payment_no" label="回款流水号" width="260" />
            <el-table-column label="金额" width="120" align="right">
              <template #default="{ row }">
                <span class="price">{{ formatCurrency(row.amount) }}</span>
              </template>
            </el-table-column>
            <el-table-column label="日期" width="180">
              <template #default="{ row }">
                {{ formatDate(row.payment_date) }}
              </template>
            </el-table-column>
            <el-table-column label="方式" width="100" align="center">
              <template #default="{ row }">
                {{ getPaymentMethodLabel(row.payment_method) }}
              </template>
            </el-table-column>
            <el-table-column label="状态" width="100" align="center">
              <template #default="{ row }">
                <el-tag :type="getPaymentStatusTag(row.status)" size="small">
                  {{ getPaymentStatusLabel(row.status) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="remark" label="备注" min-width="140" />
            <el-table-column label="凭证" width="100" align="center">
              <template #default="{ row }">
                <el-image
                  v-if="row.voucher_url"
                  :src="row.voucher_url"
                  :preview-src-list="[row.voucher_url]"
                  fit="cover"
                  style="width: 60px; height: 60px; border-radius: 4px; cursor: pointer"
                  preview-teleported
                />
                <span v-else style="color: #ccc">无</span>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="140" align="center">
              <template #default="{ row }">
                <el-button
                  v-if="row.status === 0"
                  type="success" link size="small" @click="handleApprovePayment(row)"
                >
                  确认
                </el-button>
                <el-button
                  v-if="row.status === 0"
                  type="danger" link size="small" @click="handleRejectPayment(row)"
                >
                  驳回
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>

        <!-- 送货状态 -->
        <el-card class="info-card" shadow="never">
          <template #header>
            <div class="section-header">
              <span class="section-title">送货信息</span>
              <el-button
                v-if="Number(detail.order_status) === 1 && Number(detail.delivery_status) === 0 && (!detail.outbound_request || detail.outbound_request.status === 3)"
                type="warning"
                size="small"
                icon="Promotion"
                @click="outboundRequestDialogVisible = true"
              >
                申请出库
              </el-button>
              <el-button
                v-if="Number(detail.order_status) === 1 && Number(detail.delivery_status) === 0 && detail.outbound_request?.status === 4 && !detail.outbound_confirmed"
                type="primary"
                size="small"
                icon="Box"
                @click="handleOutbound"
              >
                出库确认
              </el-button>
              <el-button
                v-if="Number(detail.order_status) === 1 && Number(detail.delivery_status) === 0 && detail.outbound_request?.status === 4 && detail.outbound_confirmed"
                type="success"
                size="small"
                icon="Printer"
                @click="handlePrintDelivery"
              >
                打印送货单
              </el-button>
              <el-button
                v-if="Number(detail.order_status) === 1 && Number(detail.delivery_status) === 1"
                type="warning"
                size="small"
                icon="Check"
                @click="handleConfirmDelivery"
              >
                确认送达
              </el-button>
            </div>
          </template>
          <el-descriptions :column="3" border>
            <el-descriptions-item label="送货状态">
              <el-tag :type="getDeliveryStatusTag(detail.delivery_status, detail.outbound_request)" size="small">
                {{ getDeliveryStatusLabel(detail.delivery_status, detail.outbound_request) }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="出库申请">
              <template v-if="detail.outbound_request">
                <el-tag :type="getOutboundStatusTag(detail.outbound_request.status)" size="small">
                  {{ getOutboundStatusLabel(detail.outbound_request.status) }}
                </el-tag>
                <span v-if="detail.outbound_request.status === 3" style="margin-left: 8px; color: #909399; font-size: 12px">
                  {{ detail.outbound_request.supervisor_remark || detail.outbound_request.finance_remark || '可重新申请' }}
                </span>
              </template>
              <span v-else style="color: #909399">未申请</span>
            </el-descriptions-item>
            <el-descriptions-item label="送货地址">{{ detail.customer_address || '-' }}</el-descriptions-item>
          </el-descriptions>

          <!-- 送货记录 -->
          <el-table
            v-if="detail.delivery_records && detail.delivery_records.length > 0"
            :data="detail.delivery_records"
            border
            stripe
            style="width: 100%; margin-top: 16px"
          >
            <el-table-column prop="delivery_no" label="送货单号" width="200" />
            <el-table-column label="送货类型" width="100" align="center">
              <template #default="{ row }">
                {{ getDeliveryTypeLabel(row.delivery_type) }}
              </template>
            </el-table-column>
            <el-table-column prop="total_quantity" label="数量" width="80" align="center" />
            <el-table-column prop="receiver_name" label="收货人" width="100" />
            <el-table-column prop="receiver_phone" label="收货电话" width="140" />
            <el-table-column prop="receiver_address" label="收货地址" min-width="200" show-overflow-tooltip />
            <el-table-column prop="delivery_time" label="送货时间" width="170" />
            <el-table-column label="状态" width="80" align="center">
              <template #default="{ row }">
                <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">
                  {{ row.status === 1 ? '正常' : '作废' }}
                </el-tag>
              </template>
            </el-table-column>
          </el-table>
          <el-empty
            v-else
            description="暂无送货记录"
            :image-size="60"
            style="margin-top: 16px"
          />
        </el-card>

        <!-- 操作按钮 -->
        <div class="action-bar">
          <el-button
            v-if="Number(detail.order_status) === 0"
            type="success"
            @click="handleApprove"
          >
            审核订单
          </el-button>
          <el-button
            v-if="Number(detail.order_status) === 1 && Number(detail.delivery_status) === 0"
            type="primary"
            @click="handleGeneratePurchase"
          >
            一键生成采购单
          </el-button>
          <el-button
            v-if="Number(detail.order_status) === 1"
            type="warning"
            @click="handleReturn"
          >
            退货
          </el-button>
        </div>
      </template>
    </el-card>

    <!-- 选择商品生成采购单弹窗 -->
    <el-dialog v-dialog-drag v-model="purchaseSelectDialogVisible" title="选择商品生成采购单" width="900px" destroy-on-close>
      <div class="purchase-select-tip">
        <el-alert type="info" :closable="false">
          <template #title>
            请选择需要采购的商品（库存不足的商品已自动勾选）
          </template>
        </el-alert>
      </div>
      <el-table ref="purchaseSelectTable" :data="purchaseSelectItems" border stripe max-height="400" @selection-change="handlePurchaseSelectionChange">
        <el-table-column type="selection" width="55" />
        <el-table-column label="商品名称" min-width="160">
          <template #default="{ row }">
            {{ row.sku_name || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="品牌款式" width="120">
          <template #default="{ row }">
            {{ row.brand || '-' }}<span v-if="row.brand && row.style"> <br/> </span>{{ row.style || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="SKU编码" width="120">
          <template #default="{ row }">
            {{ row.sku_code || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="订单数量" width="90" align="center" prop="quantity" />
        <el-table-column label="可用库存" width="90" align="center">
          <template #default="{ row }">
            <span :class="{ 'low-stock': row.available_qty < row.quantity }">
              {{ row.available_qty || 0 }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="需采购数量" width="100" align="center">
          <template #default="{ row }">
            <span class="need-purchase">{{ Math.max(0, row.quantity - (row.available_qty || 0)) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="成本价" width="100" align="right">
          <template #default="{ row }">
            ¥{{ row.sku?.product?.cost_price || 0 }}
          </template>
        </el-table-column>
      </el-table>
      <div class="selected-summary" v-if="selectedPurchaseItems.length > 0">
        已选择 <strong>{{ selectedPurchaseItems.length }}</strong> 件商品，需采购合计 <strong>{{ selectedPurchaseTotalQty }}</strong> 件
      </div>

      <el-divider content-position="left">采购方式</el-divider>
      <el-radio-group v-model="mergeTarget" class="merge-target-group">
        <el-radio value="new">新建采购单</el-radio>
        <el-radio value="existing">加入已有采购单</el-radio>
      </el-radio-group>

      <div v-if="mergeTarget === 'existing'" class="merge-select-area">
        <el-select
          v-model="selectedPurchaseOrder"
          placeholder="请选择要加入的采购单"
          filterable
          :loading="mergeLoading"
          style="width: 100%"
        >
          <el-option
            v-for="po in mergeablePurchaseList"
            :key="po.id"
            :label="`${po.purchase_no} - ${po.supplier_name || '无供应商'} (¥${po.total_amount})`"
            :value="po.id"
          />
        </el-select>
        <div v-if="selectedPurchaseOrder" class="merge-hint">
          将把选中的商品追加到该采购单中
        </div>
      </div>

      <template #footer>
        <el-button @click="purchaseSelectDialogVisible = false">取消</el-button>
        <el-button
          type="primary"
          :disabled="selectedPurchaseItems.length === 0 || (mergeTarget === 'existing' && !selectedPurchaseOrder)"
          :loading="mergeLoading"
          @click="handleConfirmPurchaseSelect"
        >
          {{ mergeTarget === 'new' ? '确认并新建采购单' : '确认并加入采购单' }}
        </el-button>
      </template>
    </el-dialog>

    <!-- 审核弹窗 -->
    <ApproveDialog
      v-model:visible="approveDialogVisible"
      :order="detail"
      @success="fetchDetail"
    />

    <!-- 退货弹窗 -->
    <ReturnDialog
      v-model:visible="returnDialogVisible"
      :order="detail"
      @success="fetchDetail"
    />

    <!-- 回款录入弹窗 -->
    <PaymentDialog
      v-model:visible="paymentDialogVisible"
      :order-id="detail?.id"
      @success="fetchDetail"
    />

    <!-- 出库申请弹窗 -->
    <OutboundRequestDialog
      v-model="outboundRequestDialogVisible"
      :order="detail"
      @success="fetchDetail"
    />

    <!-- 送货单打印弹窗 -->
    <OutboundDialog
      v-model="outboundDialogVisible"
      :order="detail"
      @success="fetchDetail"
      @print="handlePrintDelivery"
    />
    <PrintDeliveryDialog
      v-model:visible="printDialogVisible"
      :order="detail"
      @success="handlePrintSuccess"
    />

    <!-- 销售合同打印弹窗 -->
    <PrintContractDialog
      v-model="contractDialogVisible"
      :order="detail || {}"
      :items="detail?.items || []"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getOrderDetail } from '@/api/order'
import { confirmDelivery } from '@/api/delivery'
import { approvePayment } from '@/api/payment'
import { getWarehouseList } from '@/api/warehouse'
import { getStockList } from '@/api/inventory'
import { getMergeablePurchases, appendPurchaseItems } from '@/api/purchase'
import { formatCurrency, formatPercent, formatDate } from '@/utils/format'
import ApproveDialog from './components/ApproveDialog.vue'
import ReturnDialog from './components/ReturnDialog.vue'
import PaymentDialog from './components/PaymentDialog.vue'
import PrintDeliveryDialog from './components/PrintDeliveryDialog.vue'
import OutboundDialog from './components/OutboundDialog.vue'
import OutboundRequestDialog from './components/OutboundRequestDialog.vue'
import { getOutboundRequestByOrder } from '@/api/outbound-request'
import PrintContractDialog from './components/PrintContractDialog.vue'

const router = useRouter()
const route = useRoute()

const loading = ref(false)
const detail = ref(null)
const outboundRequestDialogVisible = ref(false)

// 商品选择生成采购单弹窗
const purchaseSelectDialogVisible = ref(false)
const purchaseSelectTable = ref(null)
const purchaseSelectItems = ref([])
const selectedPurchaseItems = ref([])
const purchaseSelectLoading = ref(false)

// 合并模式相关
const mergeTarget = ref('new')
const selectedPurchaseOrder = ref(null)
const mergeablePurchaseList = ref([])
const mergeLoading = ref(false)

// 尾款比例
const remainingRate = computed(() => {
  if (!detail.value) return 0
  const finalAmount = parseFloat(detail.value.final_amount) || 0
  const remainingAmount = parseFloat(detail.value.remaining_amount) || 0
  if (finalAmount <= 0) return 0
  return (remainingAmount / finalAmount) * 100
})

const fetchDetail = async () => {
  loading.value = true
  try {
    const res = await getOrderDetail(route.params.id)
    const data = res.data || {}
    const order = data.order || {}

    let profitRate = 0
    const finalAmount = parseFloat(order.final_amount) || 0
    const actualProfit = parseFloat(order.actual_profit) || 0
    if (finalAmount > 0) {
      profitRate = actualProfit / finalAmount
    }

    detail.value = {
      ...order,
      profit_rate: profitRate,
      items: data.items || [],
      gifts: data.gifts || [],
      payments: data.payments || [],
      delivery_records: data.delivery_records || [],
    }
  } catch (error) {
    console.error('获取订单详情失败:', error)
    ElMessage.error('获取订单详情失败')
  } finally {
    loading.value = false
  }
}

// ==================== 状态映射 ====================
const getStatusLabel = (status) => {
  const map = { 0: '待审批', 1: '已生效', 2: '已驳回', 3: '已取消', 4: '已退货' }
  return map[status] ?? status ?? '未知'
}

const getStatusTag = (status) => {
  const map = { 0: 'warning', 1: 'success', 2: 'danger', 3: 'info', 4: '' }
  return map[status] ?? 'info'
}

const getOrderTypeLabel = (type) => {
  const map = { 1: '单品', 2: '多品', 3: '特殊审批', 4: '同行单品', 5: '同行多品', 6: '同行特批' }
  return map[type] ?? type ?? '-'
}

const getOrderTypeTag = (type) => {
  const map = { 1: '', 2: 'success', 3: 'warning', 4: 'info', 5: '', 6: 'danger' }
  return map[type] ?? 'info'
}

const getPaymentMethodLabel = (method) => {
  const map = { 0: '银行转账', 1: '现金', 2: '微信', 3: '支付宝' }
  return map[method] || method || '-'
}

const getPaymentStatusLabel = (status) => {
  const map = { 0: '待确认', 1: '已确认', 2: '已驳回' }
  return map[status] || status || '-'
}

const getPaymentStatusTag = (status) => {
  const map = { 0: 'warning', 1: 'success', 2: 'danger' }
  return map[status] || 'info'
}

const getDeliveryStatusLabel = (status, outboundRequest) => {
  // 如果有已通过的出库申请但未打印，显示"已出库"
  if (status === 0 && outboundRequest?.status === 4) return '已出库'
  const map = { 0: '待出库', 1: '配送中', 2: '已送达' }
  return map[status] ?? '未知'
}

const getDeliveryStatusTag = (status, outboundRequest) => {
  // 如果有已通过的出库申请但未打印，显示成功色
  if (status === 0 && outboundRequest?.status === 4) return 'success'
  const map = { 0: 'info', 1: 'warning', 2: 'success' }
  return map[status] ?? 'info'
}

const getDeliveryTypeLabel = (type) => {
  const map = { 1: '自送', 2: '物流', 3: '快递' }
  return map[type] ?? '-'
}

const getOutboundStatusLabel = (status) => {
  const map = { 1: '主管待审批', 2: '财务待审批', 3: '已拒绝', 4: '已通过' }
  return map[status] || '未申请'
}
const getOutboundStatusTag = (status) => {
  const map = { 1: 'warning', 2: '', 3: 'danger', 4: 'success' }
  return map[status] || 'info'
}

// ==================== 合计行 ====================
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

const getItemRowClass = ({ row }) => {
  if (row.item_status === 2) return 'item-removed-row'
  return ''
}

const getItemSummary = (param) => {
  const { columns, data } = param
  // 过滤掉移除的商品
  const validData = data.filter(row => row.item_status !== 2)
  const sums = []
  columns.forEach((column, index) => {
    if (index === 0) { sums[index] = '合计'; return }
    // index 6 = 数量
    if (index === 6) {
      sums[index] = validData.reduce((sum, row) => sum + (row.quantity || 0), 0)
      return
    }
    // index 7 = 挂牌价合计
    if (index === 7) {
      sums[index] = formatCurrency(validData.reduce((sum, row) => sum + (row.list_price || 0) * (row.quantity || 0), 0))
      return
    }
    // index 8 = 销售价合计
    if (index === 8) {
      sums[index] = formatCurrency(validData.reduce((sum, row) => sum + (row.sale_price || 0) * (row.quantity || 0), 0))
      return
    }
    // index 11 = 成本合计
    if (index === 11) {
      sums[index] = formatCurrency(validData.reduce((sum, row) => sum + (row.quantity || 0) * (row.unit_cost || 0), 0))
      return
    }
    sums[index] = ''
  })
  return sums
}

const getGiftSummary = (param) => {
  const { columns, data } = param
  const sums = []
  columns.forEach((column, index) => {
    if (index === 0) { sums[index] = '合计'; return }
    if (index === 1) {
      sums[index] = data.reduce((sum, row) => sum + (row.quantity || 0), 0)
      return
    }
    if (index === 3) {
      sums[index] = formatCurrency(data.reduce((sum, row) => sum + (row.quantity || 0) * (row.cost_price || 0), 0))
      return
    }
    sums[index] = ''
  })
  return sums
}

// ==================== 操作 ====================
const handleBack = () => {
  router.push('/order/index')
}

const approveDialogVisible = ref(false)
const returnDialogVisible = ref(false)
const contractDialogVisible = ref(false)
const paymentDialogVisible = ref(false)
const printDialogVisible = ref(false)
const outboundDialogVisible = ref(false)

const handleApprove = () => { approveDialogVisible.value = true }
const handleReturn = () => { returnDialogVisible.value = true }
const handleAddPayment = () => { paymentDialogVisible.value = true }

// 一键生成采购单 - 打开商品选择弹窗
const handleGeneratePurchase = async () => {
  if (!detail.value?.items?.length) {
    ElMessage.warning('订单没有商品明细')
    return
  }

  purchaseSelectLoading.value = true
  purchaseSelectDialogVisible.value = true
  purchaseSelectItems.value = []
  selectedPurchaseItems.value = []
  mergeTarget.value = 'new'
  selectedPurchaseOrder.value = null

  try {
    // 获取默认仓库
    const whRes = await getWarehouseList({ page: 1, page_size: 100 })
    const warehouses = Array.isArray(whRes.data) ? whRes.data : (whRes.data?.list || [])
    const defaultWarehouse = warehouses[0]

    // 查询每个商品的库存
    const itemsWithStock = await Promise.all(
      detail.value.items.map(async (item) => {
        let availableQty = 0
        try {
          const stockRes = await getStockList({
            warehouse_id: defaultWarehouse?.id,
            sku_id: item.sku_id,
            page_size: 1
          })
          const stockList = stockRes.data?.list || stockRes.data || []
          if (stockList.length > 0) {
            availableQty = stockList[0].available_quantity || stockList[0].stock_quantity || 0
          }
        } catch (e) {
          console.warn('查询库存失败:', item.sku_id, e)
        }
        return {
          ...item,
          available_qty: availableQty,
          need_purchase_qty: Math.max(0, item.quantity - availableQty),
          brand: item.sku?.product?.brand || '',
          style: item.sku?.product?.style || '',
        }
      })
    )

    purchaseSelectItems.value = itemsWithStock

    // 自动勾选库存不足的商品
    setTimeout(() => {
      if (purchaseSelectTable.value) {
        itemsWithStock.forEach((item, index) => {
          if (item.need_purchase_qty > 0) {
            purchaseSelectTable.value.toggleRowSelection(item, true)
          }
        })
      }
    }, 100)
  } catch (error) {
    console.error('加载库存信息失败:', error)
    ElMessage.error('加载库存信息失败')
  } finally {
    purchaseSelectLoading.value = false
  }

  // 加载可合并采购单列表
  loadMergeablePurchases()
}

const loadMergeablePurchases = async () => {
  mergeLoading.value = true
  try {
    const res = await getMergeablePurchases()
    mergeablePurchaseList.value = res.data || []
  } catch (error) {
    console.error('获取可合并采购单失败:', error)
  } finally {
    mergeLoading.value = false
  }
}

const handlePurchaseSelectionChange = (selection) => {
  selectedPurchaseItems.value = selection
}

const selectedPurchaseTotalQty = computed(() => {
  return selectedPurchaseItems.value.reduce((sum, item) => {
    return sum + Math.max(0, item.quantity - (item.available_qty || 0))
  }, 0)
})

// 确认选择并跳转采购单页面
const handleConfirmPurchaseSelect = async () => {
  if (selectedPurchaseItems.value.length === 0) {
    ElMessage.warning('请选择需要采购的商品')
    return
  }

  // 构建采购商品数据
  const purchaseItems = selectedPurchaseItems.value.map(item => ({
    sku_id: Number(item.sku_id) || 0,
    product_name: item.product_name || item.sku?.product?.product_name || '',
    sku_name: item.sku_name || '',
    sku_code: item.sku_code || item.sku?.sku_code || '',
    style: item.style || item.sku?.product?.style || '',
    quantity: Math.max(1, item.quantity - (item.available_qty || 0)),
    purchase_price: Number(item.sku?.product?.cost_price) || Number(item.cost_price) || Number(item.unit_cost) || (Number(item.sale_price) * 0.7) || 0,
  }))

  if (mergeTarget.value === 'existing' && selectedPurchaseOrder.value) {
    // 加入已有采购单
    try {
      mergeLoading.value = true
      await appendPurchaseItems(selectedPurchaseOrder.value, { items: purchaseItems })
      ElMessage.success('已成功追加到采购单')
      purchaseSelectDialogVisible.value = false
    } catch (error) {
      console.error('追加采购商品失败:', error)
    } finally {
      mergeLoading.value = false
    }
  } else {
    // 新建采购单，跳转到采购单页面
    router.push({
      path: '/inventory/purchase',
      query: {
        mode: 'create',
        order_id: detail.value.id,
        order_no: detail.value.order_no,
      },
      state: {
        prefillItems: purchaseItems,
      }
    })
  }
}

const handleOutbound = () => {
  outboundDialogVisible.value = true
}

const handlePrintDelivery = () => {
  printDialogVisible.value = true
}

const handleConfirmDelivery = async () => {
  try {
    await ElMessageBox.confirm('确认该订单已送达？', '确认送达', {
      confirmButtonText: '确认',
      cancelButtonText: '取消',
      type: 'warning',
    })
    await confirmDelivery(detail.value.id)
    ElMessage.success('已确认送达')
    fetchDetail()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '确认送达失败')
    }
  }
}

const handlePrintContract = () => {
  contractDialogVisible.value = true
}

const handlePrintSuccess = () => {
  // 打印审批状态更新后刷新
  fetchDetail()
}

const handleApprovePayment = async (row) => {
  try {
    await approvePayment(row.id, { approved: true })
    ElMessage.success('回款确认成功')
    fetchDetail()
  } catch (error) {
    console.error('确认回款失败:', error)
  }
}

const handleRejectPayment = async (row) => {
  try {
    const { value } = await ElMessageBox.prompt('请输入驳回原因', '驳回回款', {
      confirmButtonText: '确认驳回',
      cancelButtonText: '取消',
      inputType: 'textarea',
      inputValidator: (val) => {
        if (!val || val.trim() === '') {
          return '请输入驳回原因'
        }
        return true
      }
    })
    await approvePayment(row.id, { approved: false, reject_reason: value })
    ElMessage.success('已驳回')
    fetchDetail()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '驳回失败')
    }
  }
}

onMounted(() => {
  fetchDetail()
})
</script>

<style lang="scss" scoped>
.order-detail {
  .card-header {
    display: flex;
    align-items: center;
    justify-content: space-between;

    .title {
      font-size: 16px;
      font-weight: 600;
      color: #303133;
    }
  }

  .info-card {
    margin-bottom: 16px;

    .section-header {
      display: flex;
      align-items: center;
      justify-content: space-between;
    }

    .section-title {
      font-size: 15px;
      font-weight: 600;
      color: #303133;
    }

    .section-header {
      display: flex;
      align-items: center;
      justify-content: space-between;
    }
  }

  .price { color: #f56c6c; font-weight: 500; }
  .cost { color: #909399; }
  .discount { color: #e6a23c; }
  .profit { color: #67c23a; font-weight: 500; }
  .loss { color: #f56c6c; font-weight: 500; }

  .action-bar {
    margin-top: 24px;
    text-align: center;
  }
}

.item-removed-row {
  opacity: 0.55;
  .item-removed-text {
    text-decoration: line-through;
    color: #999;
  }
}

// 商品选择生成采购单弹窗样式
.purchase-select-tip {
  margin-bottom: 16px;
}

.selected-summary {
  margin-top: 12px;
  padding: 8px 12px;
  background: #f0f9eb;
  border-radius: 4px;
  color: #67c23a;
  font-size: 14px;
  strong {
    color: #409eff;
  }
}

.low-stock {
  color: #f56c6c;
  font-weight: 500;
}

.need-purchase {
  color: #e6a23c;
  font-weight: 500;
}

.merge-target-group {
  margin-bottom: 12px;
}

.merge-select-area {
  margin-top: 8px;
}

.merge-hint {
  margin-top: 8px;
  color: #909399;
  font-size: 13px;
}
</style>
