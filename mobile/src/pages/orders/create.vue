<template>
  <view class="create-page">
    <!-- 客户信息 -->
    <view class="form-section card">
      <text class="section-title">客户信息</text>
      <view class="form-item">
        <text class="form-label">客户姓名 <text class="required">*</text></text>
        <view class="select-customer-row" @tap="openCustomerSelect">
          <input
            v-model="form.customerName"
            class="input-field"
            placeholder="点击选择客户"
            disabled
          />
          <text class="select-btn">选择</text>
        </view>
      </view>
      <view class="form-item">
        <text class="form-label">手机号</text>
        <input v-model="form.customerPhone" class="input-field" type="number" maxlength="11" placeholder="选择客户后自动填充" disabled />
      </view>
      <view v-if="selectedCustomer" class="customer-info">
        <text class="customer-salesman">负责业务员: {{ selectedCustomer.salesman_name || '无' }}</text>
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

    <!-- 商品列表 -->
    <view class="form-section card">
      <view class="section-header-row">
        <text class="section-title">商品列表</text>
        <view class="add-btn" @tap="addProduct">
          <text class="add-btn-text">+ 添加商品</text>
        </view>
      </view>
      <view v-if="form.products.length > 0">
        <view v-for="(product, pIndex) in form.products" :key="pIndex" class="product-card">
          <view class="product-header">
            <text class="product-index">商品{{ pIndex + 1 }}</text>
            <text class="product-delete" @tap="removeProduct(pIndex)">删除</text>
          </view>
          <view class="form-item">
            <text class="form-label">商品名称 <text class="required">*</text></text>
            <view class="select-product-row" @tap="openProductSelect(pIndex)">
              <input v-model="product.name" class="input-field" placeholder="点击选择商品" disabled />
              <text class="select-btn">选择</text>
            </view>
          </view>
          <view class="form-item">
            <text class="form-label">SKU</text>
            <input v-model="product.sku" class="input-field" placeholder="选择商品后自动填充" disabled />
          </view>
          <view class="form-row">
            <view class="form-item form-item-half">
              <text class="form-label">数量</text>
              <view class="quantity-control">
                <view class="qty-btn" @tap="changeQty(pIndex, -1)"><text class="qty-btn-text">-</text></view>
                <input v-model.number="product.quantity" class="qty-input" type="number" placeholder="0" />
                <view class="qty-btn" @tap="changeQty(pIndex, 1)"><text class="qty-btn-text">+</text></view>
              </view>
            </view>
            <view class="form-item form-item-half">
              <text class="form-label">挂牌价(元)</text>
              <input v-model.number="product.listPrice" class="input-field" type="digit" placeholder="挂牌价" disabled />
            </view>
          </view>
          <view class="form-item">
            <text class="form-label">销售价(元) <text class="required">*</text></text>
            <input 
              v-model.number="product.salePrice" 
              class="input-field" 
              :class="{ 'input-error': product.minPrice && product.salePrice < product.minPrice }"
              type="digit" 
              placeholder="请输入销售价" 
              @blur="validateSalePrice(product)"
            />
            <text v-if="product.minPrice" class="price-hint">最低价: ¥{{ product.minPrice }}</text>
            <text v-if="product.minPrice && product.salePrice < product.minPrice" class="price-error">销售价不能低于最低价</text>
          </view>
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
        <view v-for="(gift, gIndex) in form.gifts" :key="gIndex" class="gift-card">
          <view class="gift-header">
            <text class="gift-index">礼品{{ gIndex + 1 }}</text>
            <text class="gift-delete" @tap="removeGift(gIndex)">删除</text>
          </view>
          <view class="form-item">
            <text class="form-label">礼品名称</text>
            <input v-model="gift.name" class="input-field" placeholder="请输入礼品名称" />
          </view>
          <view class="form-item">
            <text class="form-label">数量</text>
            <view class="quantity-control">
              <view class="qty-btn" @tap="changeGiftQty(gIndex, -1)"><text class="qty-btn-text">-</text></view>
              <input v-model.number="gift.quantity" class="qty-input" type="number" placeholder="0" />
              <view class="qty-btn" @tap="changeGiftQty(gIndex, 1)"><text class="qty-btn-text">+</text></view>
            </view>
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
      :isPeerOrder="!!form.companionName"
      @close="closeCommissionEstimate"
    />
  </view>
</template>

<script>
import { ref, computed } from 'vue'
import { createOrder } from '../../api/order'
import ProductSelect from '../../components/ProductSelect/ProductSelect.vue'
import CustomerSelect from '../../components/CustomerSelect/CustomerSelect.vue'
import PeerSelect from '../../components/PeerSelect/PeerSelect.vue'
import FollowUpApproval from '../../components/FollowUpApproval/FollowUpApproval.vue'
import CommissionEstimate from '../../components/CommissionEstimate/CommissionEstimate.vue'
import { useUserStore } from '../../store/user'

export default {
  components: {
    ProductSelect,
    CustomerSelect,
    PeerSelect,
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
    const showFollowUpApproval = ref(false)
    const showCommissionEstimate = ref(false)
    const currentProductIndex = ref(-1)
    const selectedCustomer = ref(null)
    const pendingFollowUpCustomer = ref(null)

    const form = ref({
      customerId: null,
      customerName: '',
      customerPhone: '',
      companionName: '',
      companionPhone: '',
      products: [],
      gifts: [],
      specialApproval: false
    })

    // ========== 客户选择 ==========
    const openCustomerSelect = () => {
      showCustomerSelect.value = true
    }
    const closeCustomerSelect = () => {
      showCustomerSelect.value = false
    }
    const onCustomerSelect = (customer) => {
      selectedCustomer.value = customer
      form.value.customerId = customer.id
      form.value.customerName = customer.customer_name || customer.name
      form.value.customerPhone = customer.phone
    }
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
        form.value.products[index].name = product.sku_name || product.product?.product_name || ''
        form.value.products[index].sku = product.sku_code
        form.value.products[index].listPrice = product.product?.list_price || 0
        form.value.products[index].minPrice = product.product?.min_price || 0
        form.value.products[index].costPrice = product.product?.cost_price || 0
        form.value.products[index].salePrice = product.product?.list_price || ''
        form.value.products[index].stock = product.available_stock || 0
        // 数量不能超过库存
        const stock = product.available_stock || 0
        if (stock > 0 && form.value.products[index].quantity > stock) {
          form.value.products[index].quantity = stock
        }
      }
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
        salePrice: '',
        stock: 0
      })
      openProductSelect(form.value.products.length - 1)
    }

    const removeProduct = (index) => {
      form.value.products.splice(index, 1)
    }

    const changeQty = (index, delta) => {
      const product = form.value.products[index]
      const maxStock = product.stock || 0
      const qty = (product.quantity || 0) + delta
      if (qty < 1) return
      if (maxStock > 0 && qty > maxStock) {
        uni.showToast({ title: '不能超过库存数量', icon: 'none' })
        return
      }
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
      form.value.gifts.push({ name: '', quantity: 1 })
    }
    const removeGift = (index) => {
      form.value.gifts.splice(index, 1)
    }
    const changeGiftQty = (index, delta) => {
      const qty = (form.value.gifts[index].quantity || 0) + delta
      if (qty >= 0) form.value.gifts[index].quantity = qty
    }

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
          customer_address: '',
          is_peer_order: form.value.companionName ? 1 : 0,
          is_special_approved: form.value.specialApproval ? 1 : 0,
          is_draft: 1, // 草稿
          remark: form.value.companionName ? `陪同人: ${form.value.companionName} ${form.value.companionPhone || ''}` : '',
          items: form.value.products.map(p => ({
            sku_id: p.skuId || 0,
            product_name: p.name,
            sku_name: p.sku || p.name,
            quantity: Number(p.quantity) || 1,
            list_price: Number(p.listPrice) || 0,
            sale_price: Number(p.salePrice) || 0,
          })),
          gifts: form.value.gifts.length > 0 ? form.value.gifts.map(g => ({
            gift_id: 0,
            gift_name: g.name,
            cost_price: 0,
            quantity: Number(g.quantity) || 1,
          })) : [],
        }

        await createOrder(orderData)
        uni.showToast({ title: '保存成功', icon: 'success' })
      } catch (e) {
        console.error('保存订单失败:', e)
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
          customer_address: '',
          is_peer_order: form.value.companionName ? 1 : 0,
          is_special_approved: form.value.specialApproval ? 1 : 0,
          is_draft: 0, // 正式订单
          remark: form.value.companionName ? `陪同人: ${form.value.companionName} ${form.value.companionPhone || ''}` : '',
          items: form.value.products.map(p => ({
            sku_id: p.skuId || 0,
            product_name: p.name,
            sku_name: p.sku || p.name,
            quantity: Number(p.quantity) || 1,
            list_price: Number(p.listPrice) || 0,
            sale_price: Number(p.salePrice) || 0,
          })),
          gifts: form.value.gifts.length > 0 ? form.value.gifts.map(g => ({
            gift_id: 0,
            gift_name: g.name,
            cost_price: 0,
            quantity: Number(g.quantity) || 1,
          })) : [],
        }

        await createOrder(orderData)
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
      openCommissionEstimate,
      closeCommissionEstimate,
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
.select-product-row {
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

.customer-info {
  padding: 16rpx;
  background: #f7f7f7;
  border-radius: 12rpx;
  margin-top: 16rpx;
}

.customer-salesman {
  font-size: 26rpx;
  color: #666;
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

/* 商品卡片 */
.product-card {
  background-color: #fafafa;
  border-radius: 12rpx;
  padding: 20rpx;
  margin-bottom: 20rpx;
}

.product-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16rpx;
}

.product-index {
  font-size: 26rpx;
  font-weight: 500;
  color: #333333;
}

.product-delete {
  font-size: 24rpx;
  color: #ff4d4f;
  padding: 4rpx 16rpx;
  &:active { opacity: 0.6; }
}

/* 数量控制 */
.quantity-control {
  display: flex;
  align-items: center;
  height: 80rpx;
  background-color: #f5f5f5;
  border: 2rpx solid #eeeeee;
  border-radius: 12rpx;
  overflow: hidden;
}

.qty-btn {
  width: 80rpx;
  height: 80rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: #eeeeee;
  &:active { background-color: #dddddd; }
}

.qty-btn-text {
  font-size: 32rpx;
  color: #333333;
  line-height: 1;
}

.qty-input {
  flex: 1;
  height: 80rpx;
  text-align: center;
  font-size: 28rpx;
  color: #333333;
  background-color: transparent;
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
  line-height: 88rpx;
  font-size: 30rpx;
  border-radius: 12rpx;
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
