<template>
  <view class="create-page">
    <!-- 客户信息 -->
    <view class="form-section card">
      <view class="customer-header">
        <text class="section-title">客户信息</text>
        <view class="customer-select-btn" @tap="openCustomerSelect">
          <text class="customer-select-btn-text">{{ selectedCustomer ? '更换' : '选择' }}</text>
        </view>
      </view>
      <view v-if="selectedCustomer" class="customer-detail">
        <view class="customer-main">
          <text class="customer-name">{{ form.customerName }}</text>
          <text class="customer-phone">{{ form.customerPhone || '-' }}</text>
        </view>
        <view v-if="selectedCustomer.address" class="customer-address-row">
          <text class="customer-address-label">地址</text>
          <text class="customer-address-value">{{ selectedCustomer.address }}</text>
        </view>
        <view class="customer-salesman-row">
          <text class="customer-salesman-label">业务员</text>
          <text class="customer-salesman-value">{{ form.salesmanName || selectedCustomer.salesman?.real_name || selectedCustomer.salesman_name || '无' }}</text>
        </view>
      </view>
      <view v-else class="customer-empty" @tap="openCustomerSelect">
        <text class="customer-empty-text">点击选择客户</text>
        <text class="customer-empty-hint">客户姓名、手机号、地址将自动填充</text>
      </view>
    </view>

    <!-- 同行信息（可选） -->
    <view class="form-section card">
      <view class="section-header-row" @tap="showCompanion = !showCompanion">
        <text class="section-title">同行信息（选填）</text>
        <text class="toggle-text">{{ showCompanion ? '收起' : '展开' }}</text>
      </view>
      <view v-if="showCompanion">
        <view class="form-item">
          <text class="form-label">同行名称</text>
          <view class="select-customer-row" @tap="openPeerSelect">
            <input v-model="form.companionName" class="input-field" placeholder="点击选择同行" disabled />
            <text class="select-btn">选择</text>
          </view>
        </view>
        <view class="form-item">
          <text class="form-label">同行手机号</text>
          <input v-model="form.companionPhone" class="input-field" type="number" maxlength="11" placeholder="选择同行后自动填充" disabled />
        </view>
      </view>
    </view>

    <!-- 订单备注 -->
    <view class="form-section card">
      <view class="section-header-row">
        <text class="section-title">订单备注</text>
      </view>
      <view class="form-item">
        <textarea
          v-model="form.remark"
          class="textarea-field"
          placeholder="请输入订单备注（选填）"
          :maxlength="500"
        />
      </view>
    </view>

    <!-- 商品列表 -->
    <view class="form-section card">
      <view class="section-header-row">
        <text class="section-title">商品列表</text>
        <view class="add-btn" @tap="addProduct">
          <text class="add-btn-text">+ 添加商品</text>
        </view>
      </view>
      <view v-if="form.products.length > 0">
        <view v-for="(product, pIndex) in form.products" :key="pIndex" class="product-card" @tap="openProductSelect(pIndex)">
          <!-- 第一行：序号 + 商品名称 + 删除 -->
          <view class="pc-row pc-row--header">
            <text class="pc-index">{{ pIndex + 1 }}</text>
            <text class="pc-name" @tap.stop="showProductName(product.name)">{{ product.name || '点击选择商品' }}</text>
            <text class="pc-delete" @tap.stop="removeProduct(pIndex)">删除</text>
          </view>
          <!-- 第二行：SKU + 库存 -->
          <view class="pc-row pc-row--meta">
            <text class="pc-sku">{{ product.sku || '-' }}</text>
            <text class="pc-stock" :class="{ 'pc-stock--empty': (product.stock || 0) <= 0 }">库存: {{ product.stock || 0 }}</text>
          </view>
          <!-- 第三行：数量 + 挂牌价 + 销售价 -->
          <view class="pc-row pc-row--price" @tap.stop>
            <view class="pc-qty">
              <text class="pc-label">数量</text>
              <view class="pc-qty-control">
                <view class="pc-qty-btn" @tap="changeQty(pIndex, -1)"><text class="pc-qty-btn-text">-</text></view>
                <input v-model.number="product.quantity" class="pc-qty-input" type="number" placeholder="0" @tap.stop />
                <view class="pc-qty-btn" @tap="changeQty(pIndex, 1)"><text class="pc-qty-btn-text">+</text></view>
              </view>
            </view>
            <view class="pc-price-item">
              <text class="pc-label">挂牌价</text>
              <text class="pc-price-value pc-price-value--list">¥{{ product.listPrice || '0.00' }}</text>
            </view>
            <view class="pc-price-item">
              <text class="pc-label">销售价</text>
              <input
                v-model.number="product.salePrice"
                class="pc-sale-input"
                :class="{ 'pc-sale-input--error': product.minPrice && product.salePrice < product.minPrice }"
                type="digit"
                placeholder="0.00"
                @blur="validateSalePrice(product)"
                @tap.stop
              />
            </view>
          </view>
          <!-- 最低价提示 -->
          <text v-if="product.minPrice && product.salePrice < product.minPrice" class="pc-error" @tap.stop>销售价不能低于最低价 ¥{{ product.minPrice }}</text>
        </view>
      </view>
      <view v-else class="empty-product">
        <text class="empty-text">暂未添加商品</text>
      </view>
    </view>

    <!-- 礼品区域（可选） -->
    <view class="form-section card">
      <view class="section-header-row" @tap="showGift = !showGift">
        <text class="section-title">礼品信息（选填）</text>
        <text class="toggle-text">{{ showGift ? '收起' : '展开' }}</text>
      </view>
      <view v-if="showGift">
        <view v-for="(gift, gIndex) in form.gifts" :key="gIndex" class="product-card" @tap="openGiftSelect(gIndex)">
          <!-- 第一行：序号 + 礼品名称 + 删除 -->
          <view class="pc-row pc-row--header">
            <text class="pc-index pc-index--gift">G{{ gIndex + 1 }}</text>
            <text class="pc-name">{{ gift.name || '点击选择礼品' }}</text>
            <text class="pc-delete" @tap.stop="removeGift(gIndex)">删除</text>
          </view>
          <!-- 第二行：成本价 + 库存 -->
          <view v-if="gift.name" class="pc-row pc-row--meta">
            <text class="pc-sku">成本价: ¥{{ gift.costPrice || '0.00' }}</text>
            <text class="pc-stock" :class="{ 'pc-stock--empty': gift.stockQuantity && gift.quantity > gift.stockQuantity }">
              库存: {{ gift.stockQuantity || '-' }}
            </text>
          </view>
          <!-- 第三行：数量 -->
          <view class="pc-row pc-row--price" @tap.stop>
            <view class="pc-qty">
              <text class="pc-label">数量</text>
              <view class="pc-qty-control">
                <view class="pc-qty-btn" @tap="changeGiftQty(gIndex, -1)"><text class="pc-qty-btn-text">-</text></view>
                <input v-model.number="gift.quantity" class="pc-qty-input" type="number" placeholder="0" @tap.stop />
                <view class="pc-qty-btn" @tap="changeGiftQty(gIndex, 1)"><text class="pc-qty-btn-text">+</text></view>
              </view>
            </view>
            <text v-if="gift.stockQuantity && gift.quantity > gift.stockQuantity" class="pc-error">数量超过库存上限</text>
          </view>
        </view>
        <view class="add-gift-btn" @tap="addGift">
          <text class="add-gift-text">+ 添加礼品</text>
        </view>
      </view>
    </view>

    <!-- 金额汇总 -->
    <view class="form-section card">
      <view class="section-header-row">
        <text class="section-title">金额汇总</text>
        <view class="estimate-btn" @tap="openCommissionEstimate">
          <text class="estimate-btn-text">核算提成</text>
        </view>
      </view>
      <view class="summary-row">
        <text class="summary-label">挂牌总价</text>
        <text class="summary-value">¥{{ listTotal }}</text>
      </view>
      <view class="summary-row">
        <text class="summary-label">销售总价</text>
        <text class="summary-value">¥{{ saleTotal }}</text>
      </view>
      <!-- 多品折扣 -->
      <view v-if="categoryCount >= 3" class="summary-row discount-highlight">
        <text class="summary-label">多品折扣(95折)</text>
        <text class="summary-value discount">-¥{{ multiCategoryDiscount }}</text>
      </view>
      <view class="summary-row">
        <text class="summary-label">折扣金额</text>
        <text class="summary-value summary-discount">¥{{ discountAmount }}</text>
      </view>
      <view class="divider"></view>
      <view class="summary-row summary-final">
        <text class="summary-label summary-final-label">最终成交价</text>
        <text class="summary-value summary-final-value">¥{{ finalAmount }}</text>
      </view>
    </view>

    <!-- 特殊审批 -->
    <view class="form-section card">
      <view class="switch-row">
        <text class="switch-label">特殊审批</text>
        <switch :checked="form.specialApproval" color="#1890ff" @change="form.specialApproval = $event.detail.value" />
      </view>
    </view>

    <!-- 提交按钮 -->
    <view class="submit-section">
      <button class="btn-secondary save-btn" :class="{ disabled: saving }" @tap="handleSave">
        {{ saving ? '保存中...' : '保存订单' }}
      </button>
      <button class="btn-primary submit-btn" :class="{ disabled: submitting }" @tap="handleSubmit">
        {{ submitting ? '提交中...' : '提交订单' }}
      </button>
    </view>

    <!-- 客户选择弹窗 -->
    <CustomerSelect
      :visible="showCustomerSelect"
      @close="closeCustomerSelect"
      @select="onCustomerSelect"
      @applyFollowUp="onApplyFollowUp"
    />

    <!-- 同行选择弹窗 -->
    <PeerSelect
      :visible="showPeerSelect"
      @close="closePeerSelect"
      @select="onPeerSelect"
    />

    <!-- 商品选择弹窗 -->
    <ProductSelect
      :visible="showProductSelect"
      :excludeIds="selectedSkuIds"
      @close="closeProductSelect"
      @select="onProductSelect"
    />

    <!-- 申请跟进审批弹窗 -->
    <FollowUpApproval
      :visible="showFollowUpApproval"
      :customer="pendingFollowUpCustomer"
      @close="closeFollowUpApproval"
      @approved="onFollowUpApproved"
      @rejected="onFollowUpRejected"
    />

    <!-- 提成预估弹窗 -->
    <CommissionEstimate
      :visible="showCommissionEstimate"
      :items="form.products"
      :gifts="form.gifts"
      :isPeerOrder="!!form.peerId"
      @close="closeCommissionEstimate"
    />

    <!-- 礼品选择弹窗 -->
    <GiftSelect
      :visible="showGiftSelect"
      :excludeIds="selectedGiftIds"
      @close="closeGiftSelect"
      @select="onGiftSelect"
    />
  </view>
</template>

<script>
import { ref, computed } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { createOrder, updateOrder, deleteOrder, getCustomerDraft } from '../../api/order'
import { getCustomerDetail } from '../../api/customer'
import ProductSelect from '../../components/ProductSelect/ProductSelect.vue'
import CustomerSelect from '../../components/CustomerSelect/CustomerSelect.vue'
import PeerSelect from '../../components/PeerSelect/PeerSelect.vue'
import GiftSelect from '../../components/GiftSelect/GiftSelect.vue'
import FollowUpApproval from '../../components/FollowUpApproval/FollowUpApproval.vue'
import CommissionEstimate from '../../components/CommissionEstimate/CommissionEstimate.vue'
import { useUserStore } from '../../store/user'

export default {
  components: {
    ProductSelect,
    CustomerSelect,
    PeerSelect,
    GiftSelect,
    FollowUpApproval,
    CommissionEstimate
  },
  setup() {
    const showCompanion = ref(false)
    const showGift = ref(false)
    const saving = ref(false)
    const submitting = ref(false)
    const userStore = useUserStore()

    // 弹窗控制
    const showCustomerSelect = ref(false)
    const showPeerSelect = ref(false)
    const showProductSelect = ref(false)
    const showGiftSelect = ref(false)
    const showFollowUpApproval = ref(false)
    const showCommissionEstimate = ref(false)
    const currentProductIndex = ref(-1)
    const currentGiftIndex = ref(-1)
    const selectedCustomer = ref(null)
    const pendingFollowUpCustomer = ref(null)

    const form = ref({
      customerId: null,
      customerName: '',
      customerPhone: '',
      customerAddress: '',
      salesmanId: null,
      salesmanName: '',
      peerId: null,
      companionName: '',
      companionPhone: '',
      products: [],
      gifts: [],
      specialApproval: false,
      remark: '',
      draftOrderId: null
    })

    // ========== 客户选择 ==========
    const openCustomerSelect = () => {
      showCustomerSelect.value = true
    }
    const closeCustomerSelect = () => {
      showCustomerSelect.value = false
    }
    const onCustomerSelect = async (customer) => {
      selectedCustomer.value = customer
      form.value.customerId = customer.id
      form.value.customerName = customer.customer_name || customer.name
      form.value.customerPhone = customer.phone
      form.value.customerAddress = customer.address || ''

      // 查询客户历史草稿订单
      try {
        const res = await getCustomerDraft(customer.id)
        if (res.data) {
          uni.showModal({
            title: '发现历史记录',
            content: `该客户有未完成的订单（${res.data.items?.length || 0}件商品），是否加载？`,
            confirmText: '加载',
            cancelText: '不加载',
            success: (modalRes) => {
              if (modalRes.confirm) {
                loadDraftOrder(res.data)
              }
            }
          })
        }
      } catch (e) {
        // 无草稿或查询失败，忽略
      }
    }

    // 加载草稿订单数据
    const loadDraftOrder = (draft) => {
      const order = draft.order || {}
      form.value.draftOrderId = order.id || draft.id
      // 恢复业务员信息（如果草稿有）
      if (order.salesman_id) {
        form.value.salesmanId = order.salesman_id
      }
      // 从 salesman 对象或 salesman_name 字段读取业务员名称
      if (order.salesman?.real_name) {
        form.value.salesmanName = order.salesman.real_name
      } else if (order.salesman_name) {
        form.value.salesmanName = order.salesman_name
      }
      // 恢复备注
      form.value.remark = order.remark || ''
      // 恢复同行人信息
      if (order.companion_name) {
        form.value.companionName = order.companion_name
        form.value.companionPhone = order.companion_phone || ''
      }
      // 恢复商品（优先使用 order.items，否则使用 draft.items）
      const items = order.items || draft.items || []
      form.value.products = items.map(item => ({
        skuId: item.sku_id,
        name: item.product_name,
        sku: item.sku_name,
        quantity: item.quantity,
        listPrice: item.list_price,
        salePrice: item.sale_price,
        costPrice: item.unit_cost,
        minPrice: item.min_price || 0,
      }))
      // 恢复赠品（优先使用 order.gifts，否则使用 draft.gifts）
      const gifts = order.gifts || draft.gifts || []
      form.value.gifts = gifts.map(g => ({
        giftId: g.gift_id,
        name: g.gift_name,
        costPrice: g.cost_price,
        quantity: g.quantity,
      }))
    }

    // 从客户详情页跳转时自动加载客户信息和草稿
    const loadCustomerInfo = async (customerId, loadDraft = false) => {
      try {
        const res = await getCustomerDetail(customerId)
        const customer = res.data
        if (customer) {
          selectedCustomer.value = customer
          form.value.customerId = customer.id
          form.value.customerName = customer.customer_name || customer.name
          form.value.customerPhone = customer.phone
          form.value.customerAddress = customer.address || ''
        }
      } catch (e) {
        console.error('加载客户信息失败:', e)
      }

      if (loadDraft) {
        try {
          const res = await getCustomerDraft(customerId)
          if (res.data) {
            loadDraftOrder(res.data)
          }
        } catch (e) {
          console.error('加载草稿失败:', e)
        }
      }
    }

    // 页面加载时处理跳转参数
    onLoad((options) => {
      if (options.customer_id) {
        const loadDraft = options.load_draft === '1'
        loadCustomerInfo(options.customer_id, loadDraft)
      }
    })
    const onApplyFollowUp = (customer) => {
      pendingFollowUpCustomer.value = customer
      showCustomerSelect.value = false
      showFollowUpApproval.value = true
    }
    const closeFollowUpApproval = () => {
      showFollowUpApproval.value = false
      pendingFollowUpCustomer.value = null
    }
    const onFollowUpApproved = (customer) => {
      // 审批通过后自动选择该客户
      selectedCustomer.value = customer
      form.value.customerId = customer.id
      form.value.customerName = customer.customer_name || customer.name
      form.value.customerPhone = customer.phone
      form.value.customerAddress = customer.address || ''
    }

    const onFollowUpRejected = (data) => {
      uni.showToast({ title: data?.reason || '审批被拒绝', icon: 'none', duration: 3000 })
    }

    // ========== 同行选择 ==========
    const openPeerSelect = () => {
      showPeerSelect.value = true
    }
    const closePeerSelect = () => {
      showPeerSelect.value = false
    }
    const onPeerSelect = (peer) => {
      form.value.peerId = peer.id
      form.value.companionName = peer.peer_name || peer.name
      form.value.companionPhone = peer.phone
    }

    // ========== 商品选择 ==========
    const openProductSelect = (index) => {
      currentProductIndex.value = index
      showProductSelect.value = true
    }
    const closeProductSelect = () => {
      showProductSelect.value = false
    }
    const onProductSelect = (product) => {
      const index = currentProductIndex.value
      if (index >= 0 && index < form.value.products.length) {
        form.value.products[index].skuId = product.id
        form.value.products[index].name = product.product?.product_name || product.sku_name || ''
        form.value.products[index].sku = product.sku_name || ''
        form.value.products[index].skuCode = product.sku_code || ''
        form.value.products[index].listPrice = Number(product.product?.list_price) || 0
        form.value.products[index].minPrice = Number(product.product?.min_price) || 0
        form.value.products[index].costPrice = Number(product.product?.reference_cost) || 0
        form.value.products[index].salePrice = Number(product.product?.list_price) || 0
        form.value.products[index].stock = product.available_stock || 0
      }
    }

    // 显示完整商品名称
    const showProductName = (name) => {
      if (!name) return
      uni.showModal({
        title: '商品名称',
        content: name,
        showCancel: false,
        confirmText: '关闭'
      })
    }

    // 价格校验
    const validateSalePrice = (product) => {
      if (product.minPrice && product.salePrice < product.minPrice) {
        uni.showToast({ title: `销售价不能低于最低价 ¥${product.minPrice}`, icon: 'none' })
      }
    }

    const addProduct = () => {
      form.value.products.push({
        skuId: null,
        name: '',
        sku: '',
        quantity: 1,
        listPrice: 0,
        minPrice: 0,
        costPrice: 0,
        salePrice: 0,
        stock: 0
      })
      openProductSelect(form.value.products.length - 1)
    }

    const removeProduct = (index) => {
      form.value.products.splice(index, 1)
    }

    const changeQty = (index, delta) => {
      const product = form.value.products[index]
      const qty = (product.quantity || 0) + delta
      if (qty < 1) return
      product.quantity = qty
    }

    // 已选SKU ID列表（传给ProductSelect禁止重复选择）
    const selectedSkuIds = computed(() => {
      return form.value.products
        .filter(p => p.skuId)
        .map(p => p.skuId)
    })

    // ========== 礼品 ==========
    const addGift = () => {
      form.value.gifts.push({ giftId: null, name: '', costPrice: 0, stockQuantity: 0, quantity: 1 })
    }
    const removeGift = (index) => {
      form.value.gifts.splice(index, 1)
    }
    const changeGiftQty = (index, delta) => {
      const gift = form.value.gifts[index]
      const maxStock = gift.stockQuantity || 0
      const qty = (gift.quantity || 0) + delta
      if (qty < 1) return
      if (maxStock > 0 && qty > maxStock) {
        uni.showToast({ title: '不能超过库存数量', icon: 'none' })
        return
      }
      gift.quantity = qty
    }

    // ========== 礼品选择 ==========
    const openGiftSelect = (index) => {
      currentGiftIndex.value = index
      showGiftSelect.value = true
    }
    const closeGiftSelect = () => {
      showGiftSelect.value = false
    }
    const onGiftSelect = (gift) => {
      const index = currentGiftIndex.value
      if (index >= 0 && index < form.value.gifts.length) {
        form.value.gifts[index].giftId = gift.id
        form.value.gifts[index].name = gift.gift_name
        form.value.gifts[index].costPrice = gift.cost_price
        form.value.gifts[index].stockQuantity = gift.stock_quantity || 0
        // 如果当前数量超过库存，自动调整为库存数量
        if (form.value.gifts[index].quantity > (gift.stock_quantity || 0)) {
          form.value.gifts[index].quantity = Math.max(1, gift.stock_quantity || 1)
        }
      }
      closeGiftSelect()
    }

    // 已选礼品ID列表
    const selectedGiftIds = computed(() => {
      return form.value.gifts
        .filter(g => g.giftId)
        .map(g => g.giftId)
    })

    // ========== 金额计算 ==========
    const listTotal = computed(() => {
      let total = 0
      form.value.products.forEach(p => {
        total += (Number(p.listPrice) || 0) * (Number(p.quantity) || 0)
      })
      return total.toFixed(2)
    })

    const saleTotal = computed(() => {
      let total = 0
      form.value.products.forEach(p => {
        total += (Number(p.salePrice) || 0) * (Number(p.quantity) || 0)
      })
      return total.toFixed(2)
    })

    // 品类数量（简化：按商品数计算）
    const categoryCount = computed(() => {
      return form.value.products.filter(p => p.name).length
    })

    // 多品折扣（3个及以上品类95折）
    const multiCategoryDiscount = computed(() => {
      if (categoryCount.value >= 3) {
        return (Number(saleTotal.value) * 0.05).toFixed(2)
      }
      return '0.00'
    })

    // 折扣金额
    const discountAmount = computed(() => {
      const list = Number(listTotal.value) || 0
      const sale = Number(saleTotal.value) || 0
      const baseDiscount = list - sale
      const multiDiscount = Number(multiCategoryDiscount.value) || 0
      return (baseDiscount + multiDiscount).toFixed(2)
    })

    // 最终金额
    const finalAmount = computed(() => {
      let total = Number(saleTotal.value)
      if (categoryCount.value >= 3) {
        total = total * 0.95
      }
      return total.toFixed(2)
    })

    // ========== 提成预估 ==========
    const openCommissionEstimate = () => {
      if (form.value.products.length === 0) {
        uni.showToast({ title: '请先添加商品', icon: 'none' })
        return
      }
      showCommissionEstimate.value = true
    }
    const closeCommissionEstimate = () => {
      showCommissionEstimate.value = false
    }

    // ========== 表单验证 ==========
    const validate = (isDraft = false) => {
      if (!isDraft) {
        if (!form.value.customerName) {
          uni.showToast({ title: '请选择客户', icon: 'none' })
          return false
        }
        if (!form.value.customerPhone) {
          uni.showToast({ title: '客户手机号不能为空', icon: 'none' })
          return false
        }
      }
      if (form.value.products.length === 0) {
        uni.showToast({ title: '请至少添加一个商品', icon: 'none' })
        return false
      }
      for (let i = 0; i < form.value.products.length; i++) {
        const p = form.value.products[i]
        if (!p.name) {
          uni.showToast({ title: `请选择第${i + 1}个商品`, icon: 'none' })
          return false
        }
        if (!p.quantity || p.quantity <= 0) {
          uni.showToast({ title: `请填写第${i + 1}个商品数量`, icon: 'none' })
          return false
        }
        if (!isDraft) {
          if (!p.salePrice || p.salePrice <= 0) {
            uni.showToast({ title: `请填写第${i + 1}个商品销售价`, icon: 'none' })
            return false
          }
          // 最低价校验
          if (p.minPrice && p.salePrice < p.minPrice) {
            uni.showToast({ title: `第${i + 1}个商品销售价不能低于最低价`, icon: 'none' })
            return false
          }
        }
      }
      return true
    }

    // ========== 保存订单（草稿） ==========
    const handleSave = async () => {
      if (saving.value) return
      if (!validate(true)) return

      saving.value = true
      try {
        const orderData = {
          store_id: userStore.userInfo?.store_id || 1,
          salesman_id: userStore.userInfo?.id || 1,
          customer_name: form.value.customerName || '',
          customer_phone: form.value.customerPhone || '',
          customer_address: form.value.customerAddress || '',
          is_peer_order: form.value.peerId ? 1 : 0,
          peer_id: form.value.peerId || null,
          is_special_approved: form.value.specialApproval ? 1 : 0,
          is_draft: 1, // 草稿
          remark: form.value.remark || '',
          items: form.value.products.map(p => ({
            sku_id: p.skuId || 0,
            product_name: p.name,
            sku_name: p.sku || p.name,
            sku_code: p.skuCode || '',
            quantity: Number(p.quantity) || 1,
            list_price: Number(p.listPrice) || 0,
            sale_price: Number(p.salePrice) || 0,
            cost_price: Number(p.costPrice) || 0,
          })),
          gifts: form.value.gifts.length > 0 ? form.value.gifts.map(g => ({
            gift_id: g.giftId || 0,
            gift_name: g.name,
            cost_price: Number(g.costPrice) || 0,
            quantity: Number(g.quantity) || 1,
          })) : [],
        }

        let result
        if (form.value.draftOrderId) {
          // 更新现有草稿
          result = await updateOrder(form.value.draftOrderId, orderData)
        } else {
          // 创建新草稿
          result = await createOrder(orderData)
        }

        // 保存返回的订单ID
        if (result.data?.id) {
          form.value.draftOrderId = result.data.id
        }

        uni.showToast({ title: '保存成功', icon: 'success' })
      } catch (e) {
        console.error('保存订单失败:', e)
        uni.showToast({ title: '保存失败', icon: 'none' })
      } finally {
        saving.value = false
      }
    }

    // ========== 提交订单 ==========
    const handleSubmit = async () => {
      if (submitting.value) return
      if (!validate(false)) return

      submitting.value = true
      try {
        const orderData = {
          store_id: userStore.userInfo?.store_id || 1,
          salesman_id: userStore.userInfo?.id || 1,
          customer_name: form.value.customerName,
          customer_phone: form.value.customerPhone,
          customer_address: form.value.customerAddress || '',
          is_peer_order: form.value.peerId ? 1 : 0,
          peer_id: form.value.peerId || null,
          is_special_approved: form.value.specialApproval ? 1 : 0,
          is_draft: 0, // 正式订单
          remark: form.value.remark || '',
          items: form.value.products.map(p => ({
            sku_id: p.skuId || 0,
            product_name: p.name,
            sku_name: p.sku || p.name,
            sku_code: p.skuCode || '',
            quantity: Number(p.quantity) || 1,
            list_price: Number(p.listPrice) || 0,
            sale_price: Number(p.salePrice) || 0,
            cost_price: Number(p.costPrice) || 0,
          })),
          gifts: form.value.gifts.length > 0 ? form.value.gifts.map(g => ({
            gift_id: g.giftId || 0,
            gift_name: g.name,
            cost_price: Number(g.costPrice) || 0,
            quantity: Number(g.quantity) || 1,
          })) : [],
        }

        await createOrder(orderData)

        // 如果是从草稿提交的，删除原草稿订单
        if (form.value.draftOrderId) {
          try {
            await deleteOrder(form.value.draftOrderId)
          } catch (e) {
            console.error('删除草稿订单失败:', e)
          }
        }

        uni.showToast({ title: '提交成功', icon: 'success' })
        setTimeout(() => uni.navigateBack(), 1500)
      } catch (e) {
        console.error('提交订单失败:', e)
      } finally {
        submitting.value = false
      }
    }

    return {
      form,
      showCompanion,
      showGift,
      saving,
      submitting,
      selectedCustomer,
      pendingFollowUpCustomer,
      showCustomerSelect,
      showPeerSelect,
      showProductSelect,
      showGiftSelect,
      showFollowUpApproval,
      showCommissionEstimate,
      categoryCount,
      listTotal,
      saleTotal,
      multiCategoryDiscount,
      discountAmount,
      finalAmount,
      openCustomerSelect,
      closeCustomerSelect,
      onCustomerSelect,
      onApplyFollowUp,
      closeFollowUpApproval,
      onFollowUpApproved,
      onFollowUpRejected,
      openPeerSelect,
      closePeerSelect,
      onPeerSelect,
      openProductSelect,
      closeProductSelect,
      onProductSelect,
      validateSalePrice,
      addProduct,
      removeProduct,
      changeQty,
      selectedSkuIds,
      addGift,
      removeGift,
      changeGiftQty,
      openGiftSelect,
      closeGiftSelect,
      onGiftSelect,
      selectedGiftIds,
      openCommissionEstimate,
      closeCommissionEstimate,
      showProductName,
      handleSave,
      handleSubmit
    }
  }
}
</script>

<style lang="scss" scoped>
.create-page {
  min-height: 100vh;
  background-color: #f5f5f5;
  padding: 24rpx;
  padding-bottom: 200rpx;
}

.form-section {
  padding: 24rpx;
  margin-bottom: 20rpx;
}

.section-title {
  font-size: 32rpx;
  font-weight: bold;
  color: #333333;
  margin-bottom: 24rpx;
}

.section-header-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 24rpx;
  .section-title { margin-bottom: 0; }
}

.toggle-text {
  font-size: 24rpx;
  color: #1890ff;
}

.select-customer-row,
.select-product-row,
.select-gift-row {
  display: flex;
  align-items: center;
  gap: 16rpx;
  .input-field { flex: 1; }
  .select-btn {
    padding: 0 24rpx;
    height: 80rpx;
    line-height: 80rpx;
    background-color: #1890ff;
    color: #ffffff;
    font-size: 28rpx;
    border-radius: 12rpx;
  }
}

.cost-price {
  font-size: 28rpx;
  color: #ff4d4f;
  font-weight: 500;
}

.stock-info {
  font-size: 28rpx;
  color: #52c41a;
  font-weight: 500;
}

.stock-warning {
  font-size: 24rpx;
  color: #ff4d4f;
  margin-top: 8rpx;
  display: block;
}

/* 客户信息卡片 */
.customer-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20rpx;
}

.customer-select-btn {
  padding: 8rpx 24rpx;
  background: #1890ff;
  border-radius: 8rpx;
  &:active { opacity: 0.8; }
}

.customer-select-btn-text {
  font-size: 24rpx;
  color: #fff;
}

.customer-detail {
  background: #f7f9fc;
  border-radius: 12rpx;
  padding: 20rpx;
}

.customer-main {
  display: flex;
  align-items: baseline;
  gap: 16rpx;
  margin-bottom: 12rpx;
}

.customer-name {
  font-size: 32rpx;
  font-weight: 600;
  color: #1a1a1a;
}

.customer-phone {
  font-size: 26rpx;
  color: #666;
}

.customer-address-row,
.customer-salesman-row {
  display: flex;
  align-items: flex-start;
  gap: 12rpx;
  margin-top: 8rpx;
}

.customer-address-label,
.customer-salesman-label {
  font-size: 24rpx;
  color: #999;
  flex-shrink: 0;
  width: 80rpx;
}

.customer-address-value {
  font-size: 24rpx;
  color: #666;
  line-height: 1.5;
}

.customer-salesman-value {
  font-size: 24rpx;
  color: #1890ff;
  font-weight: 500;
}

.customer-empty {
  padding: 40rpx 24rpx;
  text-align: center;
  background: #f7f9fc;
  border-radius: 12rpx;
  border: 2rpx dashed #ddd;
  &:active { background: #f0f2f5; }
}

.customer-empty-text {
  font-size: 28rpx;
  color: #1890ff;
  font-weight: 500;
  display: block;
}

.customer-empty-hint {
  font-size: 22rpx;
  color: #999;
  margin-top: 8rpx;
  display: block;
}

.form-item {
  margin-bottom: 24rpx;
  &:last-child { margin-bottom: 0; }
}

.form-label {
  font-size: 28rpx;
  color: #666666;
  margin-bottom: 12rpx;
  display: block;
}

.required { color: #ff4d4f; }

.form-row {
  display: flex;
  gap: 20rpx;
}

.form-item-half { flex: 1; }

.input-field {
  width: 100%;
  height: 80rpx;
  background-color: #f5f5f5;
  border: 2rpx solid #eeeeee;
  border-radius: 12rpx;
  padding: 0 24rpx;
  font-size: 28rpx;
  color: #333333;
  &.input-error {
    border-color: #ff4d4f;
    background-color: #fff2f0;
  }
}

.textarea-field {
  width: 100%;
  height: 120rpx;
  background-color: #f5f5f5;
  border: 2rpx solid #eeeeee;
  border-radius: 12rpx;
  padding: 16rpx 24rpx;
  font-size: 28rpx;
  color: #333333;
  box-sizing: border-box;
}

.price-hint {
  font-size: 24rpx;
  color: #999;
  margin-top: 8rpx;
  display: block;
}

.price-error {
  font-size: 24rpx;
  color: #ff4d4f;
  margin-top: 8rpx;
  display: block;
}

/* 商品卡片 - 紧凑布局 */
.product-card {
  background-color: #fff;
  border-radius: 16rpx;
  padding: 24rpx;
  margin-bottom: 16rpx;
  box-shadow: 0 2rpx 12rpx rgba(0, 0, 0, 0.06);
}

.pc-row {
  display: flex;
  align-items: center;

  &--header {
    justify-content: space-between;
    margin-bottom: 12rpx;
  }

  &--meta {
    justify-content: space-between;
    margin-bottom: 16rpx;
  }

  &--price {
    gap: 16rpx;
    flex-wrap: wrap;
  }
}

.pc-index {
  width: 40rpx;
  height: 40rpx;
  line-height: 40rpx;
  text-align: center;
  background: #1890ff;
  color: #fff;
  font-size: 22rpx;
  font-weight: 600;
  border-radius: 8rpx;
  flex-shrink: 0;

  &--gift {
    background: #faad14;
  }
}

.pc-name {
  flex: 1;
  min-width: 0;
  font-size: 28rpx;
  font-weight: 500;
  color: #1a1a1a;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  margin: 0 16rpx;
  padding: 4rpx 0;
}

.pc-delete {
  font-size: 24rpx;
  color: #ff4d4f;
  padding: 4rpx 16rpx;
  flex-shrink: 0;
  &:active { opacity: 0.6; }
}

.pc-sku {
  font-size: 24rpx;
  color: #999;
  background: #f5f5f5;
  padding: 4rpx 16rpx;
  border-radius: 6rpx;
}

.pc-stock {
  font-size: 24rpx;
  color: #52c41a;
  font-weight: 500;

  &--empty {
    color: #ff4d4f;
  }
}

.pc-label {
  font-size: 22rpx;
  color: #999;
  display: block;
  margin-bottom: 8rpx;
}

.pc-qty {
  flex-shrink: 0;
}

.pc-qty-control {
  display: flex;
  align-items: center;
  height: 64rpx;
  background-color: #f7f7f7;
  border-radius: 10rpx;
  overflow: hidden;
  border: 2rpx solid #eee;
}

.pc-qty-btn {
  width: 64rpx;
  height: 64rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: #eee;
  &:active { background-color: #ddd; }
}

.pc-qty-btn-text {
  font-size: 28rpx;
  color: #333;
  line-height: 1;
}

.pc-qty-input {
  width: 80rpx;
  height: 64rpx;
  text-align: center;
  font-size: 28rpx;
  color: #333;
  background-color: transparent;
}

.pc-price-item {
  flex: 1;
  min-width: 0;
}

.pc-price-value {
  font-size: 26rpx;
  color: #999;
  font-weight: 400;

  &--list {
    color: #666;
  }
}

.pc-sale-input {
  width: 100%;
  height: 64rpx;
  background-color: #fff8f0;
  border: 2rpx solid #ffd591;
  border-radius: 10rpx;
  padding: 0 16rpx;
  font-size: 28rpx;
  color: #e8453c;
  font-weight: 600;

  &--error {
    border-color: #ff4d4f;
    background-color: #fff2f0;
  }
}

.pc-error {
  font-size: 22rpx;
  color: #ff4d4f;
  margin-top: 8rpx;
  display: block;
}

/* 礼品 */
.gift-card {
  background-color: #fafafa;
  border-radius: 12rpx;
  padding: 20rpx;
  margin-bottom: 16rpx;
}

.gift-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16rpx;
}

.gift-index {
  font-size: 26rpx;
  font-weight: 500;
  color: #333333;
}

.gift-delete {
  font-size: 24rpx;
  color: #ff4d4f;
  padding: 4rpx 16rpx;
  &:active { opacity: 0.6; }
}

.add-btn {
  padding: 8rpx 20rpx;
  background-color: #e6f7ff;
  border-radius: 8rpx;
  &:active { opacity: 0.8; }
}

.add-btn-text {
  font-size: 24rpx;
  color: #1890ff;
}

.add-gift-btn {
  padding: 16rpx 0;
  text-align: center;
  &:active { opacity: 0.8; }
}

.add-gift-text {
  font-size: 24rpx;
  color: #1890ff;
}

.empty-product {
  padding: 40rpx 0;
  text-align: center;
}

.empty-text {
  font-size: 26rpx;
  color: #cccccc;
}

/* 金额汇总 */
.summary-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12rpx 0;
}

.summary-label {
  font-size: 28rpx;
  color: #666666;
}

.summary-value {
  font-size: 28rpx;
  color: #333333;
  font-weight: 500;
}

.summary-discount { color: #ff4d4f; }

.discount-highlight {
  background: #fff7e6;
  margin: 0 -24rpx;
  padding: 12rpx 24rpx;
  border-radius: 8rpx;
  .discount { color: #ff9500; font-weight: 600; }
}

.summary-final { padding-top: 16rpx; }

.summary-final-label {
  font-size: 30rpx;
  font-weight: bold;
  color: #333333;
}

.summary-final-value {
  font-size: 36rpx;
  font-weight: bold;
  color: #ff4d4f;
}

.divider {
  height: 1rpx;
  background-color: #eeeeee;
  margin: 16rpx 0;
}

.estimate-btn {
  padding: 8rpx 20rpx;
  background-color: #1890ff;
  border-radius: 8rpx;
  &:active { opacity: 0.8; }
}

.estimate-btn-text {
  font-size: 24rpx;
  color: #ffffff;
}

/* 开关 */
.switch-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.switch-label {
  font-size: 28rpx;
  color: #333333;
}

/* 提交 */
.submit-section {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  padding: 20rpx 24rpx;
  padding-bottom: calc(20rpx + env(safe-area-inset-bottom));
  background: #fff;
  display: flex;
  gap: 20rpx;
  box-shadow: 0 -2rpx 10rpx rgba(0, 0, 0, 0.05);
}

.save-btn,
.submit-btn {
  flex: 1;
  height: 88rpx;
  font-size: 30rpx;
  border-radius: 12rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  line-height: 1;
  &.disabled { opacity: 0.6; }
}

.save-btn {
  background-color: #f5f5f5;
  color: #666666;
}

.submit-btn {
  background-color: #1890ff;
  color: #ffffff;
}
</style>
