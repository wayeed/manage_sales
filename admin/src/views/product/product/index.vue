<template>
  <div class="product-list">
    <!-- 搜索区域 -->
    <el-card shadow="never" class="search-card">
      <el-form :inline="true" :model="searchForm" class="search-form">
        <el-form-item label="商品名称">
          <el-input
            v-model="searchForm.keyword"
            placeholder="请输入商品名称或编码"
            clearable
            @keyup.enter="handleSearch"
          />
        </el-form-item>
        <el-form-item label="商品分类">
          <el-select v-model="searchForm.category_id" placeholder="全部" clearable style="width: 150px">
            <el-option
              v-for="cat in categoryList"
              :key="cat.id"
              :label="cat.category_name"
              :value="cat.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" placeholder="全部" clearable style="width: 100px">
            <el-option label="上架" :value="1" />
            <el-option label="下架" :value="0" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">查询</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 列表区域 -->
    <el-card shadow="never" class="table-card">
      <template #header>
        <div class="card-header">
          <span class="title">商品列表</span>
          <div class="header-actions">
            <el-button type="primary" @click="handleAdd">
              <el-icon><Plus /></el-icon>
              新增商品
            </el-button>
          </div>
        </div>
      </template>

      <el-table :data="productList" border stripe v-loading="loading" style="width: 100%">
        <el-table-column label="图片" width="80" align="center">
          <template #default="{ row }">
            <el-image
              v-if="row.product_image"
              :src="row.product_image"
              :preview-src-list="[row.product_image]"
              fit="cover"
              style="width: 50px; height: 50px; border-radius: 4px"
              preview-teleported
            />
            <span v-else style="color: #c0c4cc; font-size: 12px">暂无</span>
          </template>
        </el-table-column>
        <el-table-column prop="product_code" label="商品编码" width="120" />
        <el-table-column prop="product_name" label="商品名称" min-width="180" show-overflow-tooltip />
        <el-table-column prop="category" label="分类" width="120">
          <template #default="{ row }">
            {{ row.category?.category_name || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="cost_price" label="成本价" width="100" align="right">
          <template #default="{ row }">
            ¥{{ Number(row.cost_price || 0).toFixed(2) }}
          </template>
        </el-table-column>
        <el-table-column prop="list_price" label="挂牌价" width="100" align="right">
          <template #default="{ row }">
            ¥{{ Number(row.list_price || 0).toFixed(2) }}
          </template>
        </el-table-column>
        <el-table-column prop="min_price" label="最低价" width="100" align="right">
          <template #default="{ row }">
            ¥{{ Number(row.min_price || 0).toFixed(2) }}
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">
              {{ row.status === 1 ? '上架' : '下架' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="170" />
        <el-table-column label="操作" width="160" align="center" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleEdit(row)">编辑</el-button>
            <el-button
              :type="row.status === 1 ? 'warning' : 'success'"
              link
              size="small"
              @click="handleToggleStatus(row)"
            >
              {{ row.status === 1 ? '下架' : '上架' }}
            </el-button>
            <el-button type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :total="pagination.total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="fetchProductList"
          @current-change="fetchProductList"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { getProductList, updateProduct, deleteProduct } from '@/api/product'
import { getCategoryList } from '@/api/category'

const router = useRouter()
const loading = ref(false)
const productList = ref([])
const categoryList = ref([])

const searchForm = reactive({
  keyword: '',
  category_id: null,
  status: null,
})

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0,
})

const fetchCategoryList = async () => {
  try {
    const res = await getCategoryList()
    categoryList.value = res.data || []
  } catch (error) {
    console.error('获取分类列表失败:', error)
  }
}

const fetchProductList = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      pageSize: pagination.pageSize,
      ...searchForm,
    }
    const res = await getProductList(params)
    if (res.data) {
      productList.value = res.data.list || []
      pagination.total = res.data.total || 0
    }
  } catch (error) {
    console.error('获取商品列表失败:', error)
    ElMessage.error('获取商品列表失败')
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  pagination.page = 1
  fetchProductList()
}

const handleReset = () => {
  searchForm.keyword = ''
  searchForm.category_id = null
  searchForm.status = null
  handleSearch()
}

const handleAdd = () => {
  router.push('/product/product/add')
}

const handleEdit = (row) => {
  router.push({ path: '/product/product/edit', query: { id: row.id } })
}

const handleToggleStatus = async (row) => {
  const action = row.status === 1 ? '下架' : '上架'
  try {
    await ElMessageBox.confirm(`确定要${action}商品"${row.product_name}"吗？`, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    })
    await updateProduct(row.id, { status: row.status === 1 ? 0 : 1 })
    ElMessage.success(`${action}成功`)
    fetchProductList()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('操作失败:', error)
      ElMessage.error('操作失败')
    }
  }
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(`确定要删除商品"${row.product_name}"吗？此操作不可恢复。`, '警告', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    })
    await deleteProduct(row.id)
    ElMessage.success('删除成功')
    fetchProductList()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除失败:', error)
      ElMessage.error('删除失败')
    }
  }
}

onMounted(() => {
  fetchCategoryList()
  fetchProductList()
})
</script>

<style lang="scss" scoped>
.product-list {
  .search-card {
    margin-bottom: 16px;
  }

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

  .pagination-wrapper {
    display: flex;
    justify-content: flex-end;
    margin-top: 16px;
  }
}
</style>
