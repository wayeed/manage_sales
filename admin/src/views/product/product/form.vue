<template>
  <div class="product-form-page">
    <el-card shadow="never">
      <template #header>
        <div class="page-header">
          <el-button type="text" @click="goBack">
            <el-icon><ArrowLeft /></el-icon>
            返回列表
          </el-button>
          <span class="title">{{ isEdit ? '编辑商品' : '新增商品' }}</span>
        </div>
      </template>

      <el-form
        ref="formRef"
        :model="formData"
        :rules="formRules"
        label-width="120px"
        v-loading="loading"
      >
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="商品编码" prop="product_code">
              <el-input v-model="formData.product_code" placeholder="请输入商品编码" :disabled="isEdit" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="商品名称" prop="product_name">
              <el-input v-model="formData.product_name" placeholder="请输入商品名称" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="品牌">
              <el-input v-model="formData.brand" placeholder="请输入品牌" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="商品分类" prop="category_id">
              <el-select v-model="formData.category_id" placeholder="请选择分类" style="width: 100%">
                <el-option
                  v-for="cat in categoryList"
                  :key="cat.id"
                  :label="cat.category_name"
                  :value="cat.id"
                />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>

        <!-- 商品图片 -->
        <el-row :gutter="20">
          <el-col :span="24">
            <el-form-item label="商品图片">
              <el-upload
                class="product-image-uploader"
                action="/api/upload/image"
                :headers="uploadHeaders"
                name="file"
                :show-file-list="false"
                :on-success="onImageUploadSuccess"
                :before-upload="beforeImageUpload"
                accept="image/*"
              >
                <el-image
                  v-if="formData.product_image"
                  :src="formData.product_image"
                  fit="cover"
                  class="product-image"
                />
                <el-icon v-else class="product-image-uploader-icon"><Plus /></el-icon>
              </el-upload>
              <div class="image-tip">支持 jpg/png/gif/webp，最大 5MB</div>
            </el-form-item>
          </el-col>
        </el-row>

        <!-- 价格区域 -->
        <el-divider content-position="left">价格信息</el-divider>

        <el-row :gutter="20">
          <el-col :span="6">
            <el-form-item label="进货价">
              <el-input-number
                v-model="formData.reference_cost"
                :min="0"
                :precision="2"
                :step="10"
                controls-position="right"
                style="width: 100%"
                @change="onReferenceCostChange"
              />
            </el-form-item>
          </el-col>
          <el-col :span="6">
            <el-form-item label="成本系数">
              <el-input-number
                v-model="formData.total_cost_rate"
                :min="0"
                :max="99.99"
                :precision="4"
                :step="0.01"
                controls-position="right"
                style="width: 100%"
                @change="onCostRateChange"
              />
            </el-form-item>
          </el-col>
          <el-col :span="6">
            <el-form-item label="成本价">
              <el-input-number
                v-model="formData.cost_price"
                :min="0"
                :precision="2"
                :step="10"
                controls-position="right"
                style="width: 100%"
                @change="onCostPriceChange"
              />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="6">
            <el-form-item label="挂牌价" prop="list_price">
              <el-input-number
                v-model="formData.list_price"
                :min="0"
                :precision="2"
                :step="10"
                controls-position="right"
                style="width: 100%"
                @change="onListPriceChange"
              />
            </el-form-item>
          </el-col>
          <el-col :span="6">
            <el-form-item label="折扣系数">
              <el-input-number
                v-model="discountRate"
                :min="0"
                :max="1"
                :precision="4"
                :step="0.01"
                controls-position="right"
                style="width: 100%"
                @change="onDiscountRateChange"
              />
            </el-form-item>
          </el-col>
          <el-col :span="6">
            <el-form-item label="最低价">
              <el-input-number
                v-model="formData.min_price"
                :min="0"
                :precision="2"
                :step="10"
                controls-position="right"
                style="width: 100%"
                @change="onMinPriceChange"
              />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="8">
            <el-form-item label="库存预警">
              <el-input-number
                v-model="formData.warning_stock"
                :min="0"
                :precision="0"
                :step="10"
                controls-position="right"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
        </el-row>

        <!-- 商品描述 -->
        <el-divider content-position="left">商品描述</el-divider>

        <el-form-item label="商品描述" prop="description">
          <div style="border: 1px solid #ccc; width: 100%">
            <Toolbar
              style="border-bottom: 1px solid #ccc"
              :editor="editorRef"
              :defaultConfig="toolbarConfig"
              mode="simple"
            />
            <Editor
              style="height: 300px; overflow-y: hidden"
              v-model="formData.description"
              :defaultConfig="editorConfig"
              mode="simple"
              @onCreated="handleCreated"
            />
          </div>
        </el-form-item>

        <!-- SKU规格管理（仅编辑模式） -->
        <template v-if="isEdit">
          <el-divider content-position="left">SKU规格管理</el-divider>

          <el-form-item label="">
            <div class="sku-section">
              <div class="sku-header">
                <el-button type="primary" size="small" @click="openSkuDialog()">
                  <el-icon><Plus /></el-icon>
                  新增SKU
                </el-button>
              </div>

              <el-table :data="skuList" border stripe size="small" style="width: 100%; margin-top: 12px">
                <el-table-column prop="sku_code" label="SKU编码" width="120" />
                <el-table-column prop="sku_name" label="SKU名称" min-width="150" show-overflow-tooltip />
                <el-table-column prop="attributes" label="规格属性" min-width="150">
                  <template #default="{ row }">
                    {{ formatAttributes(row.attributes) }}
                  </template>
                </el-table-column>
                <el-table-column prop="barcode" label="条码" width="120">
                  <template #default="{ row }">
                    {{ row.barcode || '-' }}
                  </template>
                </el-table-column>
                <el-table-column prop="status" label="状态" width="80" align="center">
                  <template #default="{ row }">
                    <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">
                      {{ row.status === 1 ? '启用' : '禁用' }}
                    </el-tag>
                  </template>
                </el-table-column>
                <el-table-column label="操作" width="120" align="center">
                  <template #default="{ row }">
                    <el-button type="primary" link size="small" @click="openSkuDialog(row)">编辑</el-button>
                    <el-button type="danger" link size="small" @click="handleDeleteSku(row)">删除</el-button>
                  </template>
                </el-table-column>
              </el-table>
            </div>
          </el-form-item>
        </template>

        <el-form-item>
          <el-button type="primary" @click="handleSubmit" :loading="submitLoading">
            {{ isEdit ? '保存修改' : '立即创建' }}
          </el-button>
          <el-button @click="goBack">取消</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- SKU编辑弹窗 -->
    <SkuFormDialog
      v-model="skuDialogVisible"
      :sku-data="currentSku"
      :product-id="productId"
      @success="handleSkuSuccess"
    />
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onBeforeUnmount, shallowRef } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowLeft } from '@element-plus/icons-vue'
import { Editor, Toolbar } from '@wangeditor/editor-for-vue'
import '@wangeditor/editor/dist/css/style.css'
import { getProductDetail, createProduct, updateProduct, getSkuList, createSku, updateSku, deleteSku } from '@/api/product'
import { getCategoryList } from '@/api/category'
import { getConfig } from '@/api/config'
import { getToken } from '@/utils/auth'
import SkuFormDialog from './components/SkuFormDialog.vue'

const router = useRouter()
const route = useRoute()
const formRef = ref(null)
const loading = ref(false)
const submitLoading = ref(false)
const categoryList = ref([])
const editorRef = shallowRef()

const isEdit = ref(false)
const productId = ref(null)

// SKU相关
const skuList = ref([])
const skuDialogVisible = ref(false)
const currentSku = ref(null)

// 折扣系数（从系统配置获取）
const discountRate = ref(1.0)

// 标记是否正在程序化修改值，避免触发循环
let updating = false

const formData = reactive({
  product_code: '',
  product_name: '',
  category_id: null,
  cost_price: 0,
  list_price: 0,
  min_price: 0,
  description: '',
  brand: '',
  product_image: '',
  reference_cost: 0,
  total_cost_rate: 1.2,
  warning_stock: 10,
})

const uploadHeaders = {
  Authorization: `Bearer ${getToken()}`,
}

const beforeImageUpload = (file) => {
  const isImage = file.type.startsWith('image/')
  const isLt5M = file.size / 1024 / 1024 < 5
  if (!isImage) {
    ElMessage.error('只能上传图片文件!')
    return false
  }
  if (!isLt5M) {
    ElMessage.error('图片大小不能超过 5MB!')
    return false
  }
  return true
}

const onImageUploadSuccess = (response) => {
  if (response.errno === 0 && response.data) {
    formData.product_image = response.data.url
    ElMessage.success('图片上传成功')
  } else {
    ElMessage.error('图片上传失败')
  }
}

const formRules = {
  product_name: [{ required: true, message: '请输入商品名称', trigger: 'blur' }],
  category_id: [{ required: true, message: '请选择商品分类', trigger: 'change' }],
  list_price: [{ required: true, message: '请输入挂牌价', trigger: 'blur' }],
}

const toolbarConfig = {}
const editorConfig = {
  placeholder: '请输入商品描述...',
  MENU_CONF: {
    uploadImage: {
      server: '/api/upload/image',
      fieldName: 'file',
      maxFileSize: 5 * 1024 * 1024,
      allowedFileTypes: ['image/*'],
      headers: {
        Authorization: `Bearer ${getToken()}`,
      },
    },
  },
}

const handleCreated = (editor) => {
  editorRef.value = editor
}

const goBack = () => {
  router.push('/product/product')
}

// ========== 价格联动计算 ==========

// 改进货价 → 自动按成本系数计算成本价
const onReferenceCostChange = (val) => {
  if (updating) return
  updating = true
  formData.cost_price = round2(Number(val) * Number(formData.total_cost_rate))
  updating = false
}

// 改成本价 → 反算成本系数 = 成本价 / 进货价
const onCostPriceChange = (val) => {
  if (updating) return
  updating = true
  if (Number(formData.reference_cost) > 0) {
    formData.total_cost_rate = round4(Number(val) / Number(formData.reference_cost))
  }
  updating = false
}

// 改成本系数 → 自动重算成本价 = 进货价 × 成本系数
const onCostRateChange = (val) => {
  if (updating) return
  updating = true
  formData.cost_price = round2(Number(formData.reference_cost) * Number(val))
  updating = false
}

// 改挂牌价 → 自动按折扣系数计算最低价
const onListPriceChange = (val) => {
  if (updating) return
  updating = true
  formData.min_price = round2(Number(val) * Number(discountRate.value))
  updating = false
}

// 改最低价 → 反算折扣系数 = 最低价 / 挂牌价
const onMinPriceChange = (val) => {
  if (updating) return
  updating = true
  if (Number(formData.list_price) > 0) {
    discountRate.value = round4(Number(val) / Number(formData.list_price))
  }
  updating = false
}

// 改折扣系数 → 自动重算最低价 = 挂牌价 × 折扣系数
const onDiscountRateChange = (val) => {
  if (updating) return
  updating = true
  formData.min_price = round2(Number(formData.list_price) * Number(val))
  updating = false
}

// 保留两位小数
const round2 = (num) => Math.round(num * 100) / 100
// 保留四位小数
const round4 = (num) => Math.round(num * 10000) / 10000

// ========== 数据获取 ==========

const fetchCategoryList = async () => {
  try {
    const res = await getCategoryList()
    categoryList.value = res.data || []
  } catch (error) {
    console.error('获取分类列表失败:', error)
  }
}

const fetchDiscountRate = async () => {
  try {
    const res = await getConfig('min_discount_rate')
    if (res.data && res.data.config_value) {
      discountRate.value = Number(res.data.config_value)
    }
  } catch (error) {
    console.error('获取折扣系数失败:', error)
  }
}

const fetchProductDetail = async (id) => {
  loading.value = true
  try {
    const res = await getProductDetail(id)
    if (res.data) {
      Object.assign(formData, {
        product_code: res.data.product_code || '',
        product_name: res.data.product_name || '',
        category_id: res.data.category_id || null,
        unit: res.data.unit || '件',
        cost_price: Number(res.data.cost_price) || 0,
        list_price: Number(res.data.list_price) || Number(res.data.sale_price) || 0,
        min_price: Number(res.data.min_price) || 0,
        description: res.data.description || '',
        brand: res.data.brand || '',
        product_image: res.data.product_image || '',
        reference_cost: Number(res.data.reference_cost) || 0,
        total_cost_rate: Number(res.data.total_cost_rate) || 1.2,
        warning_stock: res.data.warning_stock || 10,
      })
      // 如果有挂牌价和最低价，反算折扣系数
      if (formData.list_price > 0 && formData.min_price > 0) {
        discountRate.value = round4(formData.min_price / formData.list_price)
      }
    }
  } catch (error) {
    console.error('获取商品详情失败:', error)
    ElMessage.error('获取商品详情失败')
  } finally {
    loading.value = false
  }
}

const handleSubmit = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  submitLoading.value = true
  try {
    if (isEdit.value) {
      await updateProduct(productId.value, formData)
      ElMessage.success('修改成功')
    } else {
      await createProduct(formData)
      ElMessage.success('创建成功')
    }
    goBack()
  } catch (error) {
    console.error('保存商品失败:', error)
    ElMessage.error(isEdit.value ? '修改失败' : '创建失败')
  } finally {
    submitLoading.value = false
  }
}

// ========== SKU管理 ==========

const fetchSkuList = async () => {
  if (!productId.value) return
  try {
    const res = await getSkuList(productId.value)
    skuList.value = res.data || []
  } catch (error) {
    console.error('获取SKU列表失败:', error)
  }
}

const openSkuDialog = (sku = null) => {
  currentSku.value = sku
  skuDialogVisible.value = true
}

const handleSkuSuccess = async (data, isEditSku) => {
  try {
    if (isEditSku) {
      await updateSku(currentSku.value.id, data)
      ElMessage.success('SKU修改成功')
    } else {
      await createSku(productId.value, data)
      ElMessage.success('SKU创建成功')
    }
    fetchSkuList()
  } catch (error) {
    console.error('保存SKU失败:', error)
    ElMessage.error(isEditSku ? 'SKU修改失败' : 'SKU创建失败')
  }
}

const handleDeleteSku = async (row) => {
  try {
    await ElMessageBox.confirm(`确定要删除SKU"${row.sku_name}"吗？`, '警告', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    })
    await deleteSku(row.id)
    ElMessage.success('删除成功')
    fetchSkuList()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除SKU失败:', error)
      ElMessage.error('删除失败')
    }
  }
}

const formatAttributes = (attrs) => {
  if (!attrs) return '-'
  try {
    const obj = typeof attrs === 'string' ? JSON.parse(attrs) : attrs
    return Object.entries(obj).map(([k, v]) => `${k}: ${v}`).join(', ')
  } catch {
    return '-'
  }
}

onMounted(() => {
  fetchCategoryList()
  fetchDiscountRate()

  // 判断是否为编辑模式
  if (route.query.id) {
    isEdit.value = true
    productId.value = Number(route.query.id)
    fetchProductDetail(productId.value)
    fetchSkuList()
  }
})

// 组件销毁时，也及时销毁编辑器
onBeforeUnmount(() => {
  const editor = editorRef.value
  if (editor == null) return
  editor.destroy()
})
</script>

<style lang="scss" scoped>
.product-form-page {
  .page-header {
    display: flex;
    align-items: center;
    gap: 16px;

    .title {
      font-size: 16px;
      font-weight: 600;
      color: #303133;
    }
  }

  .product-image-uploader {
    :deep(.el-upload) {
      border: 1px dashed #d9d9d9;
      border-radius: 6px;
      cursor: pointer;
      position: relative;
      overflow: hidden;
      transition: border-color 0.3s;

      &:hover {
        border-color: #409eff;
      }
    }
  }

  .product-image {
    width: 148px;
    height: 148px;
    display: block;
  }

  .product-image-uploader-icon {
    font-size: 28px;
    color: #8c939d;
    width: 148px;
    height: 148px;
    text-align: center;
    line-height: 148px;
  }

  .image-tip {
    font-size: 12px;
    color: #909399;
    margin-top: 4px;
  }

  .sku-section {
    width: 100%;

    .sku-header {
      display: flex;
      justify-content: flex-end;
    }
  }
}
</style>
