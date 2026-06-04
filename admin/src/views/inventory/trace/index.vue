<template>
  <div class="inventory-trace">
    <!-- 搜索栏 -->
    <el-card class="search-card" shadow="never">
      <el-tabs v-model="activeTab" @tab-change="handleTabChange">
        <el-tab-pane label="正向穿透（订单溯源）" name="forward" />
        <el-tab-pane label="反向穿透（批次去向）" name="backward" />
      </el-tabs>

      <el-form :model="searchForm" inline style="margin-top: 16px">
        <!-- 正向穿透 -->
        <template v-if="activeTab === 'forward'">
          <el-form-item label="订单号">
            <el-input
              v-model="searchForm.order_no"
              placeholder="请输入订单号"
              clearable
              style="width: 260px"
              @keyup.enter="handleForwardTrace"
            />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" icon="Search" :loading="loading" @click="handleForwardTrace">查询</el-button>
          </el-form-item>
        </template>

        <!-- 反向穿透 -->
        <template v-if="activeTab === 'backward'">
          <el-form-item label="批次号">
            <el-input
              v-model="searchForm.batch_no"
              placeholder="请输入批次号"
              clearable
              style="width: 260px"
              @keyup.enter="handleBackwardTrace"
            />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" icon="Search" :loading="loading" @click="handleBackwardTrace">查询</el-button>
          </el-form-item>
        </template>
      </el-form>
    </el-card>

    <!-- 正向穿透结果 -->
    <el-card v-if="activeTab === 'forward' && forwardResult" class="result-card" shadow="never">
      <template #header>
        <div class="card-header">
          <span>正向穿透结果</span>
          <el-tag type="info">订单号: {{ forwardResult.order_no }}</el-tag>
        </div>
      </template>

      <div class="trace-summary">
        <span>客户: <strong>{{ forwardResult.customer_name }}</strong></span>
        <span>电话: <strong>{{ forwardResult.customer_phone || '--' }}</strong></span>
        <span>地址: <strong>{{ forwardResult.customer_address || '--' }}</strong></span>
        <span>业务员: <strong>{{ forwardResult.salesman_name }}</strong></span>
      </div>

      <div v-for="(item, index) in forwardResult.items" :key="index" class="trace-chain">
        <div class="chain-title">
          <el-tag type="primary" size="small">商品 {{ index + 1 }}</el-tag>
          <span class="sku-info">{{ item.sku_name }}（{{ item.sku_code || '--' }}）</span>
          <span class="cost-info">数量: {{ item.quantity }} | 成本: ¥{{ item.total_cost }}</span>
        </div>

        <!-- 出库记录 -->
        <div v-if="item.delivery" class="chain-node delivery-node">
          <div class="node-label">出库记录</div>
          <div class="node-content">
            <p>送货单号: <strong>{{ item.delivery.delivery_no }}</strong></p>
            <p>送货时间: {{ item.delivery.delivery_time }}</p>
            <p>出库数量: {{ item.delivery.quantity }}</p>
          </div>
        </div>
        <div v-if="item.delivery" class="chain-arrow">↓</div>

        <!-- 库存批次 -->
        <div v-if="item.batch" class="chain-node batch-node">
          <div class="node-label">库存批次</div>
          <div class="node-content">
            <p>批次号: <strong>{{ item.batch.batch_no }}</strong></p>
            <p>采购价: ¥{{ item.batch.purchase_price }} | 初始: {{ item.batch.initial_quantity }} | 剩余: {{ item.batch.remaining_quantity }}</p>
            <p>仓库: {{ item.batch.warehouse_name || '--' }} | 入库日期: {{ item.batch.entry_date || '--' }}</p>
          </div>
        </div>
        <div v-if="item.batch" class="chain-arrow">↓</div>

        <!-- 采购来源 -->
        <div v-if="item.purchase" class="chain-node purchase-node">
          <div class="node-label">采购来源</div>
          <div class="node-content">
            <p>采购单号: <strong>{{ item.purchase.purchase_no }}</strong></p>
            <p>供应商: {{ item.purchase.supplier_name }}</p>
            <p>采购价: ¥{{ item.purchase.purchase_price }} | 采购数量: {{ item.purchase.purchase_quantity }}</p>
            <p>入库日期: {{ item.purchase.receipt_date }}</p>
          </div>
        </div>

        <!-- 无批次信息 -->
        <div v-if="!item.batch && !item.purchase" class="chain-node empty-node">
          <el-empty description="暂无批次关联信息（订单尚未锁定库存）" :image-size="60" />
        </div>
      </div>

      <el-empty v-if="forwardResult && forwardResult.items.length === 0" description="该订单暂无商品明细" />
    </el-card>

    <!-- 反向穿透结果 -->
    <el-card v-if="activeTab === 'backward' && backwardResult" class="result-card" shadow="never">
      <template #header>
        <div class="card-header">
          <span>反向穿透结果</span>
          <el-tag type="info">批次号: {{ backwardResult.batch.batch_no }}</el-tag>
        </div>
      </template>

      <!-- 批次信息 -->
      <div class="trace-chain">
        <div class="chain-title">
          <el-tag type="warning" size="small">批次信息</el-tag>
          <span class="sku-info">{{ backwardResult.batch.sku_name }}（{{ backwardResult.batch.sku_code || '--' }}）</span>
        </div>
        <div class="chain-node batch-node">
          <div class="node-content">
            <p>采购单号: <strong>{{ backwardResult.batch.purchase_no || '--' }}</strong></p>
            <p>供应商: {{ backwardResult.batch.supplier_name || '--' }}</p>
            <p>采购价: ¥{{ backwardResult.batch.purchase_price }} | 初始数量: {{ backwardResult.batch.initial_quantity }} | 剩余数量: {{ backwardResult.batch.remaining_quantity }}</p>
            <p>仓库: {{ backwardResult.batch.warehouse_name || '--' }} | 入库日期: {{ backwardResult.batch.entry_date || '--' }}</p>
          </div>
        </div>
      </div>

      <!-- 汇总 -->
      <div v-if="backwardResult.summary" class="trace-summary-bar">
        <div class="summary-item">
          <span class="summary-label">锁定总量</span>
          <span class="summary-value warning">{{ backwardResult.summary.total_locked }}</span>
        </div>
        <div class="summary-item">
          <span class="summary-label">已出库</span>
          <span class="summary-value success">{{ backwardResult.summary.total_delivered }}</span>
        </div>
        <div class="summary-item">
          <span class="summary-label">剩余</span>
          <span class="summary-value primary">{{ backwardResult.summary.total_remaining }}</span>
        </div>
      </div>

      <!-- 使用记录 -->
      <div class="section-title">使用记录</div>
      <el-table :data="backwardResult.transactions" border stripe style="width: 100%">
        <el-table-column prop="transaction_type_name" label="类型" width="140" />
        <el-table-column prop="quantity" label="数量" width="80" align="center" />
        <el-table-column prop="created_at" label="时间" width="170" />
        <el-table-column label="关联订单" min-width="200">
          <template #default="{ row }">
            <template v-if="row.order">
              <p>订单号: <strong>{{ row.order.order_no }}</strong></p>
              <p>客户: {{ row.order.customer_name }} | 业务员: {{ row.order.salesman_name }}</p>
            </template>
            <span v-else>--</span>
          </template>
        </el-table-column>
        <el-table-column label="出库信息" width="200">
          <template #default="{ row }">
            <template v-if="row.delivery">
              <p>送货单: {{ row.delivery.delivery_no }}</p>
              <p>{{ row.delivery.delivery_time }}</p>
            </template>
            <span v-else>--</span>
          </template>
        </el-table-column>
      </el-table>

      <el-empty v-if="backwardResult && backwardResult.transactions.length === 0" description="该批次暂无使用记录" />
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { ElMessage } from 'element-plus'
import { forwardTrace, backwardTrace } from '@/api/inventoryTrace'

const activeTab = ref('forward')
const loading = ref(false)
const forwardResult = ref(null)
const backwardResult = ref(null)

const searchForm = reactive({
  order_id: '',
  batch_no: ''
})

const handleTabChange = () => {
  forwardResult.value = null
  backwardResult.value = null
}

const handleForwardTrace = async () => {
  if (!searchForm.order_no) {
    ElMessage.warning('请输入订单号')
    return
  }
  loading.value = true
  forwardResult.value = null
  try {
    const res = await forwardTrace(searchForm.order_no)
    forwardResult.value = res.data || res
  } catch (error) {
    console.error('正向穿透查询失败:', error)
  } finally {
    loading.value = false
  }
}

const handleBackwardTrace = async () => {
  if (!searchForm.batch_no) {
    ElMessage.warning('请输入批次号')
    return
  }
  loading.value = true
  backwardResult.value = null
  try {
    const res = await backwardTrace(searchForm.batch_no)
    backwardResult.value = res.data || res
  } catch (error) {
    console.error('反向穿透查询失败:', error)
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.inventory-trace {
  padding: 20px;
}

.search-card {
  margin-bottom: 20px;
}

.result-card {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.trace-summary {
  display: flex;
  gap: 24px;
  margin-bottom: 20px;
  padding: 12px 16px;
  background: #f5f7fa;
  border-radius: 6px;
  font-size: 14px;
  color: #606266;
}

.trace-summary strong {
  color: #303133;
}

.trace-chain {
  margin-bottom: 24px;
  padding: 16px;
  border: 1px solid #ebeef5;
  border-radius: 8px;
}

.chain-title {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
  font-size: 15px;
}

.sku-info {
  font-weight: 600;
  color: #303133;
}

.cost-info {
  margin-left: auto;
  font-size: 13px;
  color: #909399;
}

.chain-arrow {
  text-align: center;
  font-size: 20px;
  color: #409eff;
  margin: 8px 0;
}

.chain-node {
  border-radius: 6px;
  padding: 12px 16px;
  margin-bottom: 4px;
}

.node-label {
  font-size: 12px;
  font-weight: 600;
  margin-bottom: 8px;
  text-transform: uppercase;
  letter-spacing: 1px;
}

.node-content p {
  margin: 4px 0;
  font-size: 13px;
  color: #606266;
}

.node-content strong {
  color: #303133;
}

.delivery-node {
  background: #ecf5ff;
  border-left: 3px solid #409eff;
}

.delivery-node .node-label {
  color: #409eff;
}

.batch-node {
  background: #fdf6ec;
  border-left: 3px solid #e6a23c;
}

.batch-node .node-label {
  color: #e6a23c;
}

.purchase-node {
  background: #f0f9eb;
  border-left: 3px solid #67c23a;
}

.purchase-node .node-label {
  color: #67c23a;
}

.empty-node {
  background: #f5f7fa;
  text-align: center;
  padding: 20px;
}

.trace-summary-bar {
  display: flex;
  gap: 40px;
  margin: 20px 0;
  padding: 16px 24px;
  background: #f5f7fa;
  border-radius: 8px;
}

.summary-item {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.summary-label {
  font-size: 12px;
  color: #909399;
  margin-bottom: 4px;
}

.summary-value {
  font-size: 24px;
  font-weight: 700;
}

.summary-value.primary {
  color: #409eff;
}

.summary-value.success {
  color: #67c23a;
}

.summary-value.warning {
  color: #e6a23c;
}

.section-title {
  font-size: 15px;
  font-weight: 600;
  color: #303133;
  margin: 20px 0 12px;
  padding-left: 8px;
  border-left: 3px solid #409eff;
}
</style>
