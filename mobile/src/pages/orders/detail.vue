<template>
  <view class="detail-page">
    <view v-if="order" class="detail-content">
      <!-- 订单状态卡片 -->
      <view class="status-banner" :class="'status-bg--' + getStatusClassName(order.order_status)">
        <view class="status-tag-large">
          <text>{{ getStatusText(order.order_status) }}</text>
          <text v-if="order.edit_count > 0" class="edit-count-tag">已修改{{ order.edit_count }}次</text>
        </view>
        <text class="status-order-no">订单号: {{ order.order_no || '--' }}</text>
        <!-- 回款、配送、出库状态 -->
        <view class="status-extra">
          <view class="status-extra-item" :class="'payment-' + order.payment_status">
            <text class="status-extra-icon">💰</text>
            <text class="status-extra-text">{{ getPaymentStatusText(order.payment_status) }}</text>
          </view>
          <view class="status-extra-item" :class="'delivery-' + order.delivery_status">
            <text class="status-extra-icon">🚚</text>
            <text class="status-extra-text">{{ getDeliveryStatusText(order.delivery_status) }}</text>
          </view>
          <view class="status-extra-item" :class="'outbound-' + getOutboundStatus(order)">
            <text class="status-extra-icon">📦</text>
            <text class="status-extra-text">{{ getOutboundStatusText(order) }}</text>
          </view>
        </view>
      </view>

      <!-- 基本信息 -->
      <view class="info-section card">
        <text class="section-title">基本信息</text>
        <view class="info-row">
          <text class="info-label">订单号</text>
          <text class="info-value">{{ order.order_no || '--' }}</text>
        </view>
        <view class="info-row">
          <text class="info-label">创建时间</text>
          <text class="info-value">{{ formatTime(order.created_at) }}</text>
        </view>
        <view class="info-row">
          <text class="info-label">业务员</text>
          <text class="info-value">{{ order.salesman?.real_name || '--' }}</text>
        </view>
      </view>

      <!-- 客户信息 -->
      <view class="info-section card">
        <text class="section-title">客户信息</text>
        <view class="info-row">
          <text class="info-label">客户姓名</text>
          <text class="info-value">{{ order.customer_name || '--' }}</text>
        </view>
        <view class="info-row">
          <text class="info-label">联系电话</text>
          <text class="info-value">{{ order.customer_phone || '--' }}</text>
        </view>
        <view class="info-row">
          <text class="info-label">收货地址</text>
          <text class="info-value">{{ order.customer_address || '--' }}</text>
        </view>
      </view>

      <!-- 金额信息 -->
      <view class="info-section card">
        <text class="section-title">金额信息</text>
        <view class="info-row">
          <text class="info-label">挂牌总价</text>
          <text class="info-value">{{ order.total_list_price || '0.00' }}元</text>
        </view>
        <view class="info-row">
          <text class="info-label">销售总价</text>
          <text class="info-value">{{ order.total_sale_price || '0.00' }}元</text>
        </view>
        <view class="info-row">
          <text class="info-label">折扣金额</text>
          <text class="info-value info-discount">{{ order.discount_amount || '0.00' }}元</text>
        </view>
        <view class="info-row info-row--highlight">
          <text class="info-label info-label--bold">成交价</text>
          <text class="info-value info-value--red">{{ order.final_amount || '0.00' }}元</text>
        </view>
        <!-- 尾款信息（仅已生效订单显示） -->
        <view class="info-row" v-if="order.order_status === 1">
          <text class="info-label">尾款金额</text>
          <text class="info-value" :class="{ 'text-danger': remainingRate > 0 }">
            {{ order.remaining_amount || '0.00' }}元
            <text v-if="remainingRate > 0" class="remaining-tip">({{ remainingRate.toFixed(1) }}%)</text>
          </text>
        </view>
        <!-- 库存状态（仅已生效订单显示） -->
        <view class="info-row" v-if="order.order_status === 1 && order.stock_status !== undefined">
          <text class="info-label">库存状态</text>
          <view class="info-value">
            <text class="tag" :class="getStockStatusClass(order.stock_status)">
              {{ getStockStatusText(order.stock_status) }}
            </text>
          </view>
        </view>
        <!-- 查看利润提成按钮 -->
        <view class="commission-btn-row">
          <button class="commission-btn" @tap="openCommissionDialog">
            <text class="commission-btn-text">查看利润提成</text>
          </button>
        </view>
      </view>

      <!-- 商品明细 -->
      <view class="info-section card">
        <text class="section-title">商品明细</text>
        <view v-if="items && items.length > 0">
          <view v-for="(item, index) in items" :key="index" class="product-detail" :class="{ 'item-removed': item.item_status === 2 }">
            <view class="product-detail-header">
              <text class="product-detail-name">{{ item.product_name || '--' }}</text>
              <text v-if="item.item_status === 1" class="item-status-tag tag-new">新增</text>
              <text v-if="item.item_status === 2" class="item-status-tag tag-removed">移除</text>
              <text class="product-detail-sku">规格: {{ item.sku_name || '--' }}</text>
              <text class="product-detail-sku" v-if="item.sku_code">编码: {{ item.sku_code }}</text>
            </view>
            <view class="product-detail-row">
              <text class="product-detail-label">数量</text>
              <text class="product-detail-value">{{ item.quantity || 0 }}</text>
            </view>
            <view class="product-detail-row">
              <text class="product-detail-label">单价</text>
              <text class="product-detail-value">{{ item.sale_price || '0.00' }}元</text>
            </view>
          </view>
        </view>
        <view v-else class="empty-hint">
          <text class="empty-hint-text">暂无商品明细</text>
        </view>
      </view>

      <!-- 礼品信息 -->
      <view v-if="gifts && gifts.length > 0" class="info-section card">
        <text class="section-title">礼品信息</text>
        <view v-for="(gift, index) in gifts" :key="index" class="info-row">
          <text class="info-label">{{ gift.name || '礼品' + (index + 1) }}</text>
          <text class="info-value">x{{ gift.quantity || 0 }}</text>
        </view>
      </view>

      <!-- 回款记录 -->
      <view class="info-section card">
        <text class="section-title">回款记录</text>
        <view v-if="payments && payments.length > 0">
          <view v-for="(payment, index) in payments" :key="index" class="payment-item">
            <view class="payment-left">
              <view class="payment-info-row">
                <text class="payment-amount">+{{ payment.amount || '0.00' }}元</text>
                <text class="payment-type-tag" :class="'type-' + (payment.payment_type || 0)">
                  {{ payment.payment_type === 1 ? '订金' : payment.payment_type === 2 ? '尾款' : '回款' }}
                </text>
              </view>
              <text class="payment-time">{{ formatTime(payment.created_at) }}</text>
              <view v-if="payment.voucher_url" class="payment-voucher" @tap="previewVoucher(payment.voucher_url)">
                <image :src="getFullImageUrl(payment.voucher_url)" mode="aspectFill" class="payment-voucher-img" />
              </view>
            </view>
            <view class="tag" :class="getPaymentRecordStatusClass(payment.status)">
              <text>{{ getPaymentRecordStatusText(payment.status) }}</text>
            </view>
          </view>
        </view>
        <view v-else class="empty-hint">
          <text class="empty-hint-text">暂无回款记录</text>
        </view>
      </view>

      <!-- 底部占位（为操作按钮留空间） -->
      <view v-if="showActions" style="height: 140rpx;"></view>
    </view>

    <!-- 加载中 -->
    <view v-else class="loading-state">
      <text class="loading-text">加载中...</text>
    </view>

    <!-- 底部操作按钮 -->
    <view v-if="showActions && order" class="bottom-actions safe-bottom">
      <!-- 待审批 -->
      <template v-if="order.order_status === 0">
        <button class="action-btn action-btn--reject" @tap="showRejectDialog">驳回</button>
        <button class="action-btn action-btn--approve" @tap="handleApprove">审核通过</button>
      </template>
      <!-- 已生效 -->
      <template v-if="order.order_status === 1">
        <button v-if="showEditBtn" class="action-btn action-btn--edit" @tap="goEditOrder">修改订单</button>
        <button v-if="order.payment_status !== 2" class="action-btn action-btn--payment" @tap="showPaymentDialog">录入回款</button>
        <button v-if="showOutboundBtn" class="action-btn action-btn--outbound" @tap="showOutboundDialog">申请出库</button>
      </template>
      <!-- 已驳回 -->
      <template v-if="order.order_status === 2">
        <button v-if="showEditBtn" class="action-btn action-btn--edit" @tap="goEditOrder">修改订单</button>
      </template>
    </view>

    <!-- 审核驳回弹窗 -->
    <view v-if="rejectDialogVisible" class="modal-mask" @tap="rejectDialogVisible = false">
      <view class="modal-content" @tap.stop>
        <text class="modal-title">驳回原因</text>
        <textarea
          v-model="rejectReason"
          class="modal-textarea"
          placeholder="请输入驳回原因"
          maxlength="200"
        />
        <view class="modal-actions">
          <button class="modal-btn modal-btn--cancel" @tap="rejectDialogVisible = false">取消</button>
          <button class="modal-btn modal-btn--confirm" @tap="handleReject">确认驳回</button>
        </view>
      </view>
    </view>

    <!-- 审核通过弹窗 -->
    <view v-if="approveDialogVisible" class="modal-mask" @tap="approveDialogVisible = false">
      <view class="modal-content" @tap.stop>
        <text class="modal-title">审核通过</text>
        <view class="modal-form-item">
          <text class="modal-form-label">订单成交价</text>
          <text class="modal-form-value">{{ order.final_amount || '0.00' }}元</text>
        </view>
        <view class="modal-form-item">
          <text class="modal-form-label">订金金额(元)</text>
          <input v-model="depositAmount" class="input-field" type="digit" placeholder="请输入订金金额（选填）" />
        </view>
        <view class="modal-actions">
          <button class="modal-btn modal-btn--cancel" @tap="approveDialogVisible = false">取消</button>
          <button class="modal-btn modal-btn--confirm" @tap="confirmApprove">确定</button>
        </view>
      </view>
    </view>

    <!-- 回款录入弹窗 -->
    <view v-if="paymentDialogVisible" class="modal-mask" @tap="paymentDialogVisible = false">
      <view class="modal-content" @tap.stop>
        <text class="modal-title">录入回款</text>
        <view class="modal-form-item">
          <text class="modal-form-label">回款金额(元)</text>
          <input v-model="paymentAmount" class="input-field" type="digit" placeholder="请输入回款金额" />
        </view>
        <view class="modal-form-item">
          <text class="modal-form-label">备注</text>
          <input v-model="paymentRemark" class="input-field" placeholder="请输入备注(选填)" />
        </view>
        <view class="modal-form-item">
          <text class="modal-form-label">付款凭证</text>
          <view class="voucher-upload-area" @tap="chooseVoucherImage">
            <image v-if="paymentVoucherUrl" :src="getFullImageUrl(paymentVoucherUrl)" mode="aspectFill" class="voucher-preview" />
            <view v-else class="voucher-upload-placeholder">
              <text class="voucher-upload-icon">+</text>
              <text class="voucher-upload-text">上传凭证</text>
            </view>
          </view>
        </view>
        <view class="modal-actions">
          <button class="modal-btn modal-btn--cancel" @tap="paymentDialogVisible = false">取消</button>
          <button class="modal-btn modal-btn--confirm" @tap="handlePayment">确认录入</button>
        </view>
      </view>
    </view>

    <!-- 申请出库弹窗 -->
    <view v-if="outboundDialogVisible" class="modal-mask" @tap="outboundDialogVisible = false">
      <view class="modal-content" @tap.stop>
        <text class="modal-title">申请出库</text>
        <!-- 订单信息 -->
        <view class="modal-form-item">
          <text class="modal-form-label">订单号</text>
          <text class="modal-form-value">{{ order.order_no || '--' }}</text>
        </view>
        <view class="modal-form-item">
          <text class="modal-form-label">客户名称</text>
          <text class="modal-form-value">{{ order.customer_name || '--' }}</text>
        </view>
        <view class="modal-form-item">
          <text class="modal-form-label">订单金额</text>
          <text class="modal-form-value text-danger">{{ order.final_amount || '0.00' }}元</text>
        </view>
        <view class="modal-form-item">
          <text class="modal-form-label">库存状态</text>
          <text class="tag" :class="getStockStatusClass(order.stock_status)">
            {{ getStockStatusText(order.stock_status) }}
          </text>
        </view>
        <view class="modal-form-item">
          <text class="modal-form-label">尾款金额</text>
          <text class="modal-form-value" :class="{ 'text-danger': remainingRate > 0 }">
            {{ order.remaining_amount || '0.00' }}元 ({{ remainingRate.toFixed(1) }}%)
          </text>
        </view>
        <!-- 尾款警告 -->
        <view v-if="remainingRate > 20" class="outbound-warning">
          <text class="warning-text">尾款比例超过20%，提交后将进入审批流程</text>
        </view>
        <!-- 审批备注预览 -->
        <view v-if="remainingRate > 20" class="outbound-remark-preview">
          <text class="remark-preview-label">审批备注：</text>
          <text class="remark-preview-text">此订单尾款还有{{ order.remaining_amount || '0.00' }}元，送货后由业务员{{ order.salesman?.real_name || '未知' }}负责收回尾款</text>
        </view>
        <view class="modal-actions">
          <button class="modal-btn modal-btn--cancel" @tap="outboundDialogVisible = false">取消</button>
          <button class="modal-btn modal-btn--confirm" :disabled="outboundSubmitting" @tap="handleOutboundRequest">
            {{ outboundSubmitting ? '提交中...' : '提交申请' }}
          </button>
        </view>
      </view>
    </view>

    <!-- 利润提成详情弹窗 -->
    <view v-if="commissionDialogVisible" class="modal-mask" @tap="commissionDialogVisible = false">
      <view class="modal-content commission-modal" @tap.stop>
        <text class="modal-title">利润提成详情</text>
        <view v-if="commissionLoading" class="commission-loading">
          <text class="loading-text">加载中...</text>
        </view>
        <view v-else-if="commissionData" class="commission-body">
          <!-- 订单类型 -->
          <view class="commission-type-tag">
            <text>{{ commissionData.order_type_name || '--' }}</text>
          </view>
          <!-- 汇总数据 -->
          <view class="commission-summary">
            <view class="commission-summary-item">
              <text class="commission-summary-label">成交价</text>
              <text class="commission-summary-value">¥{{ commissionData.final_amount || '0.00' }}</text>
            </view>
            <view class="commission-summary-item">
              <text class="commission-summary-label">总成本</text>
              <text class="commission-summary-value">¥{{ commissionData.total_cost || '0.00' }}</text>
            </view>
            <view class="commission-summary-item" v-if="Number(commissionData.gift_cost) > 0">
              <text class="commission-summary-label">礼品成本</text>
              <text class="commission-summary-value">¥{{ commissionData.gift_cost || '0.00' }}</text>
            </view>
            <view class="commission-summary-item commission-summary-item--profit">
              <text class="commission-summary-label">实际利润</text>
              <text class="commission-summary-value commission-value--profit">¥{{ commissionData.actual_profit || '0.00' }}</text>
            </view>
          </view>
          <!-- 分割线 -->
          <view class="commission-divider"></view>
          <!-- 提成信息 -->
          <view class="commission-rate-row">
            <text class="commission-rate-label">提成比例</text>
            <text class="commission-rate-value">{{ formatRate(commissionData.commission_rate) }}</text>
          </view>
          <view class="commission-amount-row">
            <text class="commission-amount-label">预估提成</text>
            <text class="commission-amount-value">¥{{ commissionData.commission_amount || '0.00' }}</text>
          </view>
          <!-- 商品成本明细 -->
          <view v-if="commissionData.items && commissionData.items.length > 0" class="commission-items">
            <text class="commission-items-title">商品成本明细</text>
            <view v-for="(cItem, cIdx) in commissionData.items" :key="cIdx" class="commission-item-card">
              <view class="commission-item-header">
                <text class="commission-item-name">{{ cItem.product_name || '--' }}</text>
                <text class="commission-item-sku">{{ cItem.sku_name || '--' }}</text>
              </view>
              <view class="commission-item-row">
                <text class="commission-item-label">数量</text>
                <text class="commission-item-value">x{{ cItem.quantity || 0 }}</text>
              </view>
              <view class="commission-item-row">
                <text class="commission-item-label">销售价</text>
                <text class="commission-item-value">¥{{ cItem.sale_price || '0.00' }}</text>
              </view>
              <view class="commission-item-row">
                <text class="commission-item-label">单位成本</text>
                <text class="commission-item-value">¥{{ cItem.unit_cost || '0.00' }}</text>
              </view>
              <view class="commission-item-row">
                <text class="commission-item-label">总成本</text>
                <text class="commission-item-value commission-value--cost">¥{{ cItem.total_cost || '0.00' }}</text>
              </view>
            </view>
          </view>
          <!-- 成本未计算提示 -->
          <view v-if="Number(commissionData.total_cost) === 0" class="commission-tip">
            <text class="commission-tip-text">成本将在送货出库后计算，当前显示为预估数据</text>
          </view>
        </view>
        <view class="modal-actions">
          <button class="modal-btn modal-btn--confirm" style="flex:1" @tap="commissionDialogVisible = false">关闭</button>
        </view>
      </view>
    </view>
  </view>
</template>

<script>
import { ref, computed, onMounted } from 'vue'
import { getOrderDetail, approveOrder, getOrderCommission } from '../../api/order'
import { post } from '../../api/request'
import { BASE_URL } from '../../api/request'
import { useUserStore } from '../../store/user'

export default {
  setup() {
    const userStore = useUserStore()
    const order = ref(null)
    const items = ref([])
    const gifts = ref([])
    const payments = ref([])
    const rejectDialogVisible = ref(false)
    const rejectReason = ref('')
    const approveDialogVisible = ref(false)
    const depositAmount = ref('')
    const paymentDialogVisible = ref(false)
    const paymentAmount = ref('')
    const paymentRemark = ref('')
    const paymentVoucherUrl = ref('')
    const submitting = ref(false)
    const commissionDialogVisible = ref(false)
    const commissionLoading = ref(false)
    const commissionData = ref(null)
    const outboundDialogVisible = ref(false)
    const outboundSubmitting = ref(false)

    const getStatusText = (status) => {
      const map = {
        0: '待审批',
        1: '已生效',
        2: '已驳回',
        3: '已取消',
        4: '已退货'
      }
      return map[status] || status || '未知'
    }

    const getStatusClassName = (status) => {
      const map = {
        0: 'pending_approval',
        1: 'effective',
        2: 'rejected',
        3: 'cancelled',
        4: 'returned'
      }
      return map[status] || ''
    }

    const getPaymentStatusText = (status) => {
      const map = { 0: '未回款', 1: '部分回款', 2: '已回款' }
      return map[status] || '未回款'
    }

    // 回款记录状态（0=待确认, 1=已确认, 2=已驳回）
    const getPaymentRecordStatusText = (status) => {
      const map = { 0: '待确认', 1: '已确认', 2: '已驳回' }
      return map[status] || '待确认'
    }

    const getPaymentRecordStatusClass = (status) => {
      const map = { 0: 'tag-warning', 1: 'tag-success', 2: 'tag-danger' }
      return map[status] || 'tag-warning'
    }

    const getDeliveryStatusText = (status) => {
      const map = { 0: '未配送', 1: '配送中', 2: '已配送' }
      return map[status] || '未配送'
    }

    // 出库申请状态
    const getOutboundStatus = (order) => {
      const req = order.outbound_request
      if (!req) return 'none' // 无申请
      const statusMap = {
        1: 'pending_supervisor', // 待主管审批
        2: 'pending_finance',    // 待财务审批
        3: 'rejected',           // 已驳回
        4: 'approved'            // 已通过
      }
      return statusMap[req.status] || 'none'
    }

    const getOutboundStatusText = (order) => {
      const req = order.outbound_request
      if (!req) return '待申请'
      const textMap = {
        1: '待主管审批',
        2: '待财务审批',
        3: '已驳回',
        4: '已通过'
      }
      return textMap[req.status] || '待申请'
    }

    // 库存状态
    const getStockStatusText = (status) => {
      const map = { 0: '库存充足', 1: '部分缺货', 2: '全部缺货' }
      return map[status] ?? '未知'
    }

    const getStockStatusClass = (status) => {
      const map = { 0: 'tag-success', 1: 'tag-warning', 2: 'tag-danger' }
      return map[status] ?? 'tag-info'
    }

    // 尾款比例
    const remainingRate = computed(() => {
      if (!order.value) return 0
      const finalAmount = parseFloat(order.value.final_amount) || 0
      const remainingAmount = parseFloat(order.value.remaining_amount) || 0
      if (finalAmount <= 0) return 0
      return (remainingAmount / finalAmount) * 100
    })

    // 是否显示申请出库按钮
    const showOutboundBtn = computed(() => {
      if (!order.value) return false
      // 已生效、未配送
      if (order.value.order_status !== 1 || order.value.delivery_status !== 0) return false
      // 库存充足
      if (order.value.stock_status !== 0) return false
      // 无申请或已被驳回（status=3为已驳回）
      const req = order.value.outbound_request
      if (req && req.status !== 3) return false
      return true
    })

    // 打开申请出库弹窗
    const showOutboundDialog = () => {
      outboundDialogVisible.value = true
    }

    // 提交出库申请
    const handleOutboundRequest = async () => {
      if (outboundSubmitting.value) return
      outboundSubmitting.value = true
      try {
        const { createOutboundRequest } = await import('@/api/outbound-request')
        await createOutboundRequest({ order_id: order.value.id })
        uni.showToast({ title: '申请已提交', icon: 'success' })
        outboundDialogVisible.value = false
        // 刷新订单详情
        loadDetail(order.value.id)
      } catch (e) {
        uni.showToast({ title: e.message || '提交失败', icon: 'none' })
      } finally {
        outboundSubmitting.value = false
      }
    }

    const formatTime = (timeStr) => {
      if (!timeStr) return '--'
      const date = new Date(timeStr)
      const year = date.getFullYear()
      const month = String(date.getMonth() + 1).padStart(2, '0')
      const day = String(date.getDate()).padStart(2, '0')
      const hour = String(date.getHours()).padStart(2, '0')
      const minute = String(date.getMinutes()).padStart(2, '0')
      const second = String(date.getSeconds()).padStart(2, '0')
      return `${year}-${month}-${day} ${hour}:${minute}:${second}`
    }

    // 是否有审核权限（老板、主管、店长、财务）
    const hasApprovePermission = computed(() => {
      const roles = userStore.userInfo?.roles || []
      const roleCodes = roles.map(r => r.role_code || r.RoleCode || '')
      return roleCodes.includes('BOSS') || roleCodes.includes('SUPERVISOR') ||
             roleCodes.includes('STORE_MANAGER') || roleCodes.includes('FINANCE')
    })

    // 是否显示操作按钮区域
    const showActions = computed(() => {
      if (!order.value) return false
      const status = order.value.order_status
      // 待审批且有审核权限
      if (status === 0 && hasApprovePermission.value) return true
      // 已生效：显示录入回款、修改订单按钮
      if (status === 1) return true
      // 已驳回且未送货：显示修改订单按钮
      if (status === 2 && order.value.delivery_status < 1) return true
      return false
    })

    // 是否显示修改订单按钮
    const showEditBtn = computed(() => {
      if (!order.value) return false
      // 已配送不能修改
      if (order.value.delivery_status >= 1) return false
      // 已驳回可以修改
      if (order.value.order_status === 2) return true
      // 已生效但未发起出库申请或申请被驳回，可以修改
      if (order.value.order_status === 1) {
        const req = order.value.outbound_request
        // 无申请或已驳回可以修改
        if (!req || req.status === 3) return true
        return false
      }
      return false
    })

    // 审核通过 - 打开弹窗
    const handleApprove = () => {
      depositAmount.value = ''
      approveDialogVisible.value = true
    }

    // 确认审核通过
    const confirmApprove = async () => {
      if (submitting.value) return
      submitting.value = true
      try {
        const data = { approved: true }
        if (depositAmount.value && Number(depositAmount.value) > 0) {
          data.deposit_amount = depositAmount.value.toString()
        }
        await approveOrder(order.value.id, data)
        uni.showToast({ title: '审核通过', icon: 'success' })
        approveDialogVisible.value = false
        loadDetail(order.value.id)
      } catch (e) {
        console.error('审核失败:', e)
        if (e.code === 403) {
          approveDialogVisible.value = false
        }
      } finally {
        submitting.value = false
      }
    }

    // 显示驳回弹窗
    const showRejectDialog = () => {
      rejectReason.value = ''
      rejectDialogVisible.value = true
    }

    // 驳回
    const handleReject = async () => {
      if (!rejectReason.value) {
        uni.showToast({ title: '请输入驳回原因', icon: 'none' })
        return
      }
      if (submitting.value) return
      submitting.value = true
      try {
        await approveOrder(order.value.id, { approved: false, reason: rejectReason.value })
        uni.showToast({ title: '已驳回', icon: 'success' })
        rejectDialogVisible.value = false
        loadDetail(order.value.id)
      } catch (e) {
        console.error('驳回失败:', e)
        if (e.code === 403) {
          rejectDialogVisible.value = false
        }
      } finally {
        submitting.value = false
      }
    }

    // 显示回款弹窗
    const showPaymentDialog = () => {
      paymentAmount.value = ''
      paymentRemark.value = ''
      paymentVoucherUrl.value = ''
      paymentDialogVisible.value = true
    }

    // 录入回款
    const handlePayment = async () => {
      if (!paymentAmount.value || Number(paymentAmount.value) <= 0) {
        uni.showToast({ title: '请输入有效金额', icon: 'none' })
        return
      }
      if (submitting.value) return
      submitting.value = true
      try {
        await post('/payments', {
          order_id: order.value.id,
          amount: paymentAmount.value,
          remark: paymentRemark.value || undefined,
          voucher_url: paymentVoucherUrl.value || undefined
        })
        uni.showToast({ title: '录入成功', icon: 'success' })
        paymentDialogVisible.value = false
        loadDetail(order.value.id)
      } catch (e) {
        console.error('录入回款失败:', e)
        paymentDialogVisible.value = false
        setTimeout(() => {
          uni.showToast({ title: e.message || e.errMsg || '录入失败', icon: 'none', duration: 3000 })
        }, 100)
      } finally {
        submitting.value = false
      }
    }

    onMounted(() => {
      const pages = getCurrentPages()
      const currentPage = pages[pages.length - 1]
      const id = currentPage.$page?.options?.id || currentPage.options?.id
      if (id) {
        loadDetail(id)
      }
    })

    const loadDetail = async (id) => {
      try {
        const res = await getOrderDetail(id)
        // 后端返回结构: { order, items, gifts, payments }
        order.value = res.data?.order || null
        items.value = res.data?.items || []
        gifts.value = res.data?.gifts || []
        payments.value = res.data?.payments || []
      } catch (e) {
        console.error('加载订单详情失败:', e)
      }
    }

    // 格式化提成比例（0.20 → 20%）
    const formatRate = (rate) => {
      if (!rate && rate !== 0) return '--'
      const num = Number(rate)
      if (isNaN(num)) return '--'
      return (num * 100).toFixed(1) + '%'
    }

    // 打开利润提成弹窗
    const openCommissionDialog = async () => {
      if (!order.value) return
      commissionDialogVisible.value = true
      commissionLoading.value = true
      commissionData.value = null
      try {
        const res = await getOrderCommission(order.value.id)
        commissionData.value = res.data || null
      } catch (e) {
        console.error('加载利润提成详情失败:', e)
        uni.showToast({ title: '加载失败', icon: 'none' })
      } finally {
        commissionLoading.value = false
      }
    }

    // 跳转修改订单
    const goEditOrder = () => {
      if (!order.value) return
      uni.navigateTo({ url: `/pages/orders/edit?id=${order.value.id}` })
    }

    // 上传凭证图片
    const chooseVoucherImage = () => {
      uni.chooseImage({
        count: 1,
        sizeType: ['compressed'],
        sourceType: ['album', 'camera'],
        success: (res) => {
          const tempPath = res.tempFilePaths[0]
          uni.uploadFile({
            url: BASE_URL + '/upload/image',
            filePath: tempPath,
            name: 'file',
            header: { 'Authorization': 'Bearer ' + uni.getStorageSync('token') },
            success: (uploadRes) => {
              const data = JSON.parse(uploadRes.data)
              if (data.errno === 0 && data.data && data.data.url) {
                paymentVoucherUrl.value = data.data.url
              } else {
                uni.showToast({ title: '上传失败', icon: 'none' })
              }
            },
            fail: () => {
              uni.showToast({ title: '上传失败', icon: 'none' })
            }
          })
        }
      })
    }

    // 获取完整图片URL
    const getFullImageUrl = (url) => {
      if (!url) return ''
      if (url.startsWith('http')) return url
      return BASE_URL.replace('/api', '') + url
    }

    // 预览凭证图片
    const previewVoucher = (url) => {
      uni.previewImage({
        urls: [getFullImageUrl(url)],
        current: getFullImageUrl(url)
      })
    }

    return {
      order,
      items,
      gifts,
      payments,
      showActions,
      showEditBtn,
      showOutboundBtn,
      rejectDialogVisible,
      rejectReason,
      approveDialogVisible,
      depositAmount,
      paymentDialogVisible,
      paymentAmount,
      paymentRemark,
      paymentVoucherUrl,
      outboundDialogVisible,
      outboundSubmitting,
      commissionDialogVisible,
      commissionLoading,
      commissionData,
      remainingRate,
      getStatusText,
      getStatusClassName,
      getPaymentStatusText,
      getPaymentRecordStatusText,
      getPaymentRecordStatusClass,
      getDeliveryStatusText,
      getOutboundStatus,
      getOutboundStatusText,
      getStockStatusText,
      getStockStatusClass,
      formatTime,
      handleApprove,
      confirmApprove,
      showRejectDialog,
      handleReject,
      showPaymentDialog,
      handlePayment,
      showOutboundDialog,
      handleOutboundRequest,
      openCommissionDialog,
      goEditOrder,
      formatRate,
      chooseVoucherImage,
      getFullImageUrl,
      previewVoucher
    }
  }
}
</script>

<style lang="scss" scoped>
.detail-page {
  min-height: 100vh;
  background-color: #f5f5f5;
  padding-bottom: 0;
}

/* 状态横幅 */
.status-banner {
  padding: 40rpx 30rpx;
  display: flex;
  flex-direction: column;
  align-items: center;

  &.status-bg--pending_approval {
    background: linear-gradient(135deg, #faad14 0%, #d48806 100%);
  }

  &.status-bg--effective {
    background: linear-gradient(135deg, #52c41a 0%, #389e0d 100%);
  }

  &.status-bg--rejected {
    background: linear-gradient(135deg, #ff4d4f 0%, #cf1322 100%);
  }

  &.status-bg--cancelled {
    background: linear-gradient(135deg, #999999 0%, #666666 100%);
  }

  &.status-bg--returned {
    background: linear-gradient(135deg, #722ed1 0%, #531dab 100%);
  }
}

.status-tag-large {
  font-size: 36rpx;
  font-weight: bold;
  color: #ffffff;
  margin-bottom: 12rpx;
}

.status-order-no {
  font-size: 24rpx;
  color: rgba(255, 255, 255, 0.8);
}

/* 回款和配送状态 */
.status-extra {
  display: flex;
  gap: 24rpx;
  margin-top: 20rpx;
}

.status-extra-item {
  display: flex;
  align-items: center;
  gap: 8rpx;
  padding: 8rpx 16rpx;
  background: rgba(255, 255, 255, 0.2);
  border-radius: 24rpx;
}

.status-extra-icon {
  font-size: 24rpx;
}

.status-extra-text {
  font-size: 22rpx;
  color: #ffffff;
}

/* 回款状态颜色 */
.status-extra-item.payment-0 {
  background: rgba(255, 77, 79, 0.3);
}
.status-extra-item.payment-1 {
  background: rgba(250, 140, 22, 0.3);
}
.status-extra-item.payment-2 {
  background: rgba(82, 196, 26, 0.3);
}

/* 配送状态颜色 */
.status-extra-item.delivery-0 {
  background: rgba(153, 153, 153, 0.3);
}
.status-extra-item.delivery-1 {
  background: rgba(24, 144, 255, 0.3);
}
.status-extra-item.delivery-2 {
  background: rgba(82, 196, 26, 0.3);
}

/* 出库申请状态颜色 */
.status-extra-item.outbound-none {
  background: rgba(153, 153, 153, 0.3);
}
.status-extra-item.outbound-pending_supervisor,
.status-extra-item.outbound-pending_finance {
  background: rgba(250, 140, 22, 0.3);
}
.status-extra-item.outbound-approved {
  background: rgba(82, 196, 26, 0.3);
}
.status-extra-item.outbound-rejected {
  background: rgba(255, 77, 79, 0.3);
}

/* 信息区域 */
.info-section {
  margin: 20rpx 24rpx;
  padding: 24rpx;
}

.section-title {
  font-size: 32rpx;
  font-weight: bold;
  color: #333333;
  margin-bottom: 20rpx;
}

.info-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12rpx 0;
  border-bottom: 1rpx solid #f5f5f5;

  &:last-child {
    border-bottom: none;
  }
}

.info-row--highlight {
  padding: 16rpx 0;
  background-color: #fffbe6;
  margin: 8rpx -24rpx 8rpx;
  padding-left: 24rpx;
  padding-right: 24rpx;
  border-bottom: none;
}

.info-label {
  font-size: 28rpx;
  color: #999999;

  &--bold {
    color: #333333;
    font-weight: bold;
  }
}

.info-value {
  font-size: 28rpx;
  color: #333333;
  max-width: 400rpx;
  text-align: right;
}

.info-discount {
  color: #ff4d4f;
}

.info-value--red {
  font-size: 32rpx;
  font-weight: bold;
  color: #ff4d4f;
}

.info-profit {
  color: #52c41a;
  font-weight: 500;
}

/* 商品明细 */
.product-detail {
  padding: 16rpx 0;
  border-bottom: 1rpx solid #f5f5f5;

  &:last-child {
    border-bottom: none;
  }
}

.product-detail-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12rpx;
}

.product-detail-name {
  font-size: 28rpx;
  font-weight: 500;
  color: #333333;
}

.product-detail-sku {
  font-size: 24rpx;
  color: #999999;
}

.product-detail-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 4rpx 0;
}

.product-detail-label {
  font-size: 26rpx;
  color: #999999;
}

.product-detail-value {
  font-size: 26rpx;
  color: #333333;
}

.empty-hint {
  padding: 30rpx 0;
  text-align: center;
}

.empty-hint-text {
  font-size: 26rpx;
  color: #cccccc;
}

/* 回款记录 */
.payment-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16rpx 0;
  border-bottom: 1rpx solid #f5f5f5;

  &:last-child {
    border-bottom: none;
  }
}

.payment-left {
  display: flex;
  flex-direction: column;
}

.payment-amount {
  font-size: 28rpx;
  font-weight: 500;
  color: #52c41a;
  margin-bottom: 4rpx;
}

.payment-time {
  font-size: 22rpx;
  color: #cccccc;
}

.payment-info-row {
  display: flex;
  align-items: center;
  gap: 12rpx;
}

.payment-type-tag {
  font-size: 20rpx;
  padding: 4rpx 12rpx;
  border-radius: 4rpx;
}

.payment-type-tag.type-1 {
  background: #fff7e6;
  color: #fa8c16;
}

.payment-type-tag.type-2 {
  background: #f6ffed;
  color: #52c41a;
}

.payment-type-tag.type-0 {
  background: #f0f0f0;
  color: #666;
}

/* 底部操作按钮 */
.bottom-actions {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  display: flex;
  gap: 20rpx;
  padding: 20rpx 24rpx;
  background-color: #ffffff;
  box-shadow: 0 -2rpx 12rpx rgba(0, 0, 0, 0.06);
  z-index: 100;
}

.action-btn {
  flex: 1;
  height: 80rpx;
  line-height: 80rpx;
  text-align: center;
  font-size: 28rpx;
  border-radius: 12rpx;
  border: none;

  &--reject {
    background-color: #ffffff;
    color: #ff4d4f;
    border: 2rpx solid #ff4d4f;

    &:active {
      background-color: #fff2f0;
    }
  }

  &--approve {
    background-color: #52c41a;
    color: #ffffff;

    &:active {
      background-color: #389e0d;
    }
  }

  &--payment {
    background-color: #1890ff;
    color: #ffffff;

    &:active {
      background-color: #096dd9;
    }
  }

  &--edit {
    background-color: #ffffff;
    color: #faad14;
    border: 2rpx solid #faad14;

    &:active {
      background-color: #fffbe6;
    }
  }
}

/* 弹窗 */
.modal-mask {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-content {
  width: 600rpx;
  background-color: #ffffff;
  border-radius: 16rpx;
  padding: 40rpx;
}

.modal-title {
  font-size: 32rpx;
  font-weight: bold;
  color: #333333;
  margin-bottom: 24rpx;
  display: block;
}

.modal-textarea {
  width: 100%;
  height: 200rpx;
  background-color: #f5f5f5;
  border: 2rpx solid #eeeeee;
  border-radius: 12rpx;
  padding: 20rpx;
  font-size: 28rpx;
  color: #333333;
  margin-bottom: 24rpx;
}

.modal-form-item {
  margin-bottom: 20rpx;
}

.modal-form-label {
  font-size: 26rpx;
  color: #666666;
  margin-bottom: 12rpx;
  display: block;
}

.modal-form-value {
  font-size: 28rpx;
  color: #ff4d4f;
  font-weight: 500;
  margin-bottom: 12rpx;
  display: block;
}

.input-field {
  width: 100%;
  height: 80rpx;
  background-color: #f5f5f5;
  border: 2rpx solid #eeeeee;
  border-radius: 12rpx;
  padding: 0 24rpx;
  font-size: 28rpx;
  color: #333333;
}

.modal-actions {
  display: flex;
  gap: 20rpx;
  margin-top: 30rpx;
}

.modal-btn {
  flex: 1;
  height: 76rpx;
  line-height: 76rpx;
  text-align: center;
  font-size: 28rpx;
  border-radius: 12rpx;
  border: none;

  &--cancel {
    background-color: #f5f5f5;
    color: #666666;

    &:active {
      background-color: #eeeeee;
    }
  }

  &--confirm {
    background-color: #1890ff;
    color: #ffffff;

    &:active {
      background-color: #096dd9;
    }
  }
}

.loading-state {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 60vh;
}

.loading-text {
  font-size: 28rpx;
  color: #999999;
}

/* 查看利润提成按钮 */
.commission-btn-row {
  display: flex;
  justify-content: center;
  padding: 24rpx 0 8rpx;
}

.commission-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 72rpx;
  padding: 0 48rpx;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 36rpx;
  border: none;

  &:active {
    opacity: 0.85;
  }
}

.commission-btn-text {
  font-size: 28rpx;
  color: #ffffff;
  font-weight: 500;
}

/* 利润提成弹窗 */
.commission-modal {
  width: 680rpx;
  max-height: 80vh;
  overflow-y: auto;
}

.commission-loading {
  padding: 60rpx 0;
  text-align: center;
}

.commission-body {
  max-height: 60vh;
  overflow-y: auto;
}

.commission-type-tag {
  display: inline-block;
  padding: 6rpx 20rpx;
  background-color: #f0f5ff;
  border-radius: 8rpx;
  margin-bottom: 24rpx;

  text {
    font-size: 24rpx;
    color: #1890ff;
  }
}

.commission-summary {
  background-color: #fafafa;
  border-radius: 12rpx;
  padding: 20rpx;
  margin-bottom: 20rpx;
}

.commission-summary-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10rpx 0;

  &--profit {
    padding: 14rpx 0;
    margin-top: 8rpx;
    border-top: 1rpx solid #eeeeee;
  }
}

.commission-summary-label {
  font-size: 26rpx;
  color: #999999;
}

.commission-summary-value {
  font-size: 28rpx;
  color: #333333;
  font-weight: 500;
}

.commission-value--profit {
  color: #52c41a;
  font-size: 32rpx;
  font-weight: bold;
}

.commission-value--cost {
  color: #ff4d4f;
}

.commission-divider {
  height: 1rpx;
  background-color: #eeeeee;
  margin: 16rpx 0;
}

.commission-rate-row,
.commission-amount-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14rpx 0;
}

.commission-rate-label,
.commission-amount-label {
  font-size: 28rpx;
  color: #666666;
}

.commission-rate-value {
  font-size: 28rpx;
  color: #1890ff;
  font-weight: 600;
}

.commission-amount-value {
  font-size: 36rpx;
  color: #ff4d4f;
  font-weight: bold;
}

.commission-items {
  margin-top: 20rpx;
}

.commission-items-title {
  font-size: 28rpx;
  font-weight: 500;
  color: #333333;
  margin-bottom: 16rpx;
  display: block;
}

.commission-item-card {
  background-color: #fafafa;
  border-radius: 12rpx;
  padding: 16rpx;
  margin-bottom: 16rpx;

  &:last-child {
    margin-bottom: 0;
  }
}

.commission-item-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12rpx;
}

.commission-item-name {
  font-size: 26rpx;
  font-weight: 500;
  color: #333333;
}

.commission-item-sku {
  font-size: 22rpx;
  color: #999999;
}

.commission-item-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 4rpx 0;
}

.commission-item-label {
  font-size: 24rpx;
  color: #999999;
}

.commission-item-value {
  font-size: 24rpx;
  color: #333333;
}

.commission-tip {
  margin-top: 20rpx;
  padding: 16rpx;
  background-color: #fffbe6;
  border-radius: 8rpx;
}

.commission-tip-text {
  font-size: 24rpx;
  color: #d48806;
}

/* 商品状态标签 */
.item-status-tag {
  font-size: 20rpx;
  padding: 4rpx 12rpx;
  border-radius: 4rpx;
  margin-left: 12rpx;
}

.item-status-tag.tag-new {
    background: #f6ffed;
    color: #52c41a;
}
.item-status-tag.tag-removed {
    background: #fff2f0;
    color: #ff4d4f;
}
.item-removed {
    opacity: 0.6;
}
.item-removed .product-detail-name,
.item-removed .product-detail-sku,
.item-removed .product-detail-value {
    text-decoration: line-through;
}

/* 修改次数标签 */
.edit-count-tag {
  font-size: 22rpx;
  color: #fa8c16;
  margin-left: 12rpx;
}

/* 凭证上传 */
.voucher-upload-area {
  margin-top: 10rpx;
}
.voucher-upload-placeholder {
  width: 160rpx;
  height: 160rpx;
  border: 2rpx dashed #ddd;
  border-radius: 8rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}
.voucher-upload-icon {
  font-size: 48rpx;
  color: #ccc;
}
.voucher-upload-text {
  font-size: 22rpx;
  color: #999;
  margin-top: 8rpx;
}
.voucher-preview {
  width: 160rpx;
  height: 160rpx;
  border-radius: 8rpx;
}

/* 回款凭证缩略图 */
.payment-voucher {
  margin-top: 8rpx;
}
.payment-voucher-img {
  width: 100rpx;
  height: 100rpx;
  border-radius: 8rpx;
}

/* 尾款提示 */
.remaining-tip {
  font-size: 22rpx;
  color: #999;
}

/* 标签样式 */
.tag {
  display: inline-block;
  padding: 4rpx 12rpx;
  border-radius: 6rpx;
  font-size: 22rpx;
}
.tag-success {
  background: #f6ffed;
  color: #52c41a;
  border: 1rpx solid #b7eb8f;
}
.tag-warning {
  background: #fffbe6;
  color: #faad14;
  border: 1rpx solid #ffe58f;
}
.tag-danger {
  background: #fff2f0;
  color: #ff4d4f;
  border: 1rpx solid #ffccc7;
}
.tag-info {
  background: #f5f5f5;
  color: #999;
  border: 1rpx solid #d9d9d9;
}

/* 文字颜色 */
.text-danger {
  color: #ff4d4f;
}

/* 申请出库按钮 */
.action-btn--outbound {
  background: linear-gradient(135deg, #1890ff 0%, #096dd9 100%);
  color: #ffffff;
}

/* 申请出库弹窗样式 */
.outbound-warning {
  background: #fff2f0;
  border: 1rpx solid #ffccc7;
  border-radius: 8rpx;
  padding: 16rpx;
  margin-top: 16rpx;
}
.warning-text {
  font-size: 24rpx;
  color: #ff4d4f;
}
.outbound-remark-preview {
  background: #fff7e6;
  border: 1rpx solid #ffd591;
  border-radius: 8rpx;
  padding: 16rpx;
  margin-top: 12rpx;
}
.remark-preview-label {
  font-size: 24rpx;
  color: #fa8c16;
  font-weight: 500;
}
.remark-preview-text {
  font-size: 24rpx;
  color: #666;
  line-height: 1.4;
}
</style>
