<template>
  <div class="store-manage">
    <!-- 操作栏 -->
    <el-card shadow="never" class="filter-card">
      <el-row :gutter="16" align="middle">
        <el-col :span="8">
          <el-input
            v-model="searchKeyword"
            placeholder="搜索门店名称/编码"
            clearable
            @keyup.enter="handleSearch"
          />
        </el-col>
        <el-col :span="4">
          <el-button type="primary" @click="handleSearch">搜索</el-button>
        </el-col>
        <el-col :span="12" style="text-align: right;">
          <el-button type="primary" @click="handleAdd">
            <el-icon><Plus /></el-icon>
            新增门店
          </el-button>
        </el-col>
      </el-row>
    </el-card>

    <!-- 门店列表 -->
    <el-card shadow="hover" style="margin-top: 16px;">
      <el-table :data="storeList" border stripe v-loading="loading" style="width: 100%">
        <el-table-column prop="store_code" label="门店编码" width="120" />
        <el-table-column prop="store_name" label="门店名称" min-width="150" />
        <el-table-column prop="address" label="地址" min-width="220" show-overflow-tooltip />
        <el-table-column prop="contact_phone" label="联系电话" width="140" />
        <el-table-column prop="status" label="状态" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">
              {{ row.status === 1 ? '启用' : '停用' }}
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
              {{ row.status === 1 ? '停用' : '启用' }}
            </el-button>
            <el-button type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 新增/编辑弹窗 -->
    
<el-dialog v-dialog-drag
      v-model="dialogVisible"
      :title="dialogTitle"
      width="560px"
      destroy-on-close
    >
      <el-form
        ref="formRef"
        :model="formData"
        :rules="formRules"
        label-width="100px"
      >
        <el-form-item label="门店编码" prop="store_code">
          <el-input v-model="formData.store_code" placeholder="请输入门店编码" :disabled="!!formData.id" />
        </el-form-item>
        <el-form-item label="门店名称" prop="store_name">
          <el-input v-model="formData.store_name" placeholder="请输入门店名称" />
        </el-form-item>
        <el-form-item label="地址" prop="address">
          <el-input v-model="formData.address" type="textarea" :rows="2" placeholder="请输入门店地址" />
        </el-form-item>
        <el-form-item label="联系电话" prop="contact_phone">
          <el-input v-model="formData.contact_phone" placeholder="请输入联系电话" />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-radio-group v-model="formData.status">
            <el-radio :value="1">启用</el-radio>
            <el-radio :value="0">停用</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { getStoreList, createStore, updateStore, deleteStore } from '@/api/store'

// ==================== 搜索 ====================
const searchKeyword = ref('')
const loading = ref(false)

const handleSearch = () => {
  fetchStoreList()
}

// ==================== 门店列表 ====================
const storeList = ref([])

const fetchStoreList = async () => {
  loading.value = true
  try {
    const res = await getStoreList()
    if (res.data) {
      let list = res.data
      if (searchKeyword.value) {
        const keyword = searchKeyword.value.toLowerCase()
        list = list.filter(
          (item) =>
            item.store_name?.toLowerCase().includes(keyword) ||
            item.store_code?.toLowerCase().includes(keyword)
        )
      }
      storeList.value = list
    }
  } catch (error) {
    console.error('获取门店列表失败:', error)
    initMockData()
  } finally {
    loading.value = false
  }
}

const initMockData = () => {
  let list = [
    { id: 1, store_code: 'STORE001', store_name: '总店', address: '北京市朝阳区建国路88号', contact_phone: '010-88888888', status: 1, created_at: '2024-01-01 10:00:00' },
    { id: 2, store_code: 'STORE002', store_name: '上海分店', address: '上海市黄浦区南京东路100号', contact_phone: '021-66666666', status: 1, created_at: '2024-02-15 14:30:00' },
    { id: 3, store_code: 'STORE003', store_name: '广州分店', address: '广州市天河区天河路200号', contact_phone: '020-33333333', status: 1, created_at: '2024-03-20 09:00:00' },
    { id: 4, store_code: 'STORE004', store_name: '深圳分店', address: '深圳市福田区华强北路50号', contact_phone: '0755-22222222', status: 0, created_at: '2024-04-10 16:00:00' },
    { id: 5, store_code: 'STORE005', store_name: '成都分店', address: '成都市锦江区春熙路88号', contact_phone: '028-11111111', status: 1, created_at: '2024-05-05 11:00:00' },
  ]
  if (searchKeyword.value) {
    const keyword = searchKeyword.value.toLowerCase()
    list = list.filter(
      (item) =>
        item.store_name.toLowerCase().includes(keyword) ||
        item.store_code.toLowerCase().includes(keyword)
    )
  }
  storeList.value = list
}

// ==================== 新增/编辑弹窗 ====================
const dialogVisible = ref(false)
const submitLoading = ref(false)
const formRef = ref(null)
const currentId = ref(null)

const dialogTitle = computed(() => (currentId.value ? '编辑门店' : '新增门店'))

const formData = reactive({
  id: null,
  store_code: '',
  store_name: '',
  address: '',
  contact_phone: '',
  status: 1,
})

const formRules = {
  store_code: [{ required: true, message: '请输入门店编码', trigger: 'blur' }],
  store_name: [{ required: true, message: '请输入门店名称', trigger: 'blur' }],
  address: [{ required: true, message: '请输入门店地址', trigger: 'blur' }],
  contact_phone: [
    { required: true, message: '请输入联系电话', trigger: 'blur' },
    { pattern: /^1[3-9]\d{9}$|^0\d{2,3}-?\d{7,8}$/, message: '请输入正确的电话号码', trigger: 'blur' },
  ],
}

const resetForm = () => {
  formData.id = null
  formData.store_code = ''
  formData.store_name = ''
  formData.address = ''
  formData.contact_phone = ''
  formData.status = 1
}

const handleAdd = () => {
  resetForm()
  currentId.value = null
  dialogVisible.value = true
}

const handleEdit = (row) => {
  currentId.value = row.id
  Object.assign(formData, {
    id: row.id,
    store_code: row.store_code,
    store_name: row.store_name,
    address: row.address,
    contact_phone: row.contact_phone,
    status: row.status,
  })
  dialogVisible.value = true
}

const handleSubmit = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  submitLoading.value = true
  try {
    if (currentId.value) {
      await updateStore(currentId.value, formData)
      ElMessage.success('门店更新成功')
    } else {
      await createStore(formData)
      ElMessage.success('门店创建成功')
    }
    dialogVisible.value = false
    fetchStoreList()
  } catch (error) {
    console.error('保存门店失败:', error)
  } finally {
    submitLoading.value = false
  }
}

// ==================== 状态切换 ====================
const handleToggleStatus = async (row) => {
  const action = row.status === 1 ? '停用' : '启用'
  try {
    await ElMessageBox.confirm(`确定要${action}门店"${row.store_name}"吗？`, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    })
    await updateStore(row.id, { status: row.status === 1 ? 0 : 1 })
    ElMessage.success(`${action}成功`)
    fetchStoreList()
  } catch (error) {
    // 用户取消操作
  }
}

// ==================== 删除 ====================
const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(`确定要删除门店"${row.store_name}"吗？此操作不可恢复。`, '警告', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    })
    await deleteStore(row.id)
    ElMessage.success('删除成功')
    fetchStoreList()
  } catch (error) {
    // 用户取消操作
  }
}

// ==================== 初始化 ====================
onMounted(() => {
  fetchStoreList()
})
</script>

<style lang="scss" scoped>
.store-manage {
  .filter-card {
    :deep(.el-card__body) {
      padding: 16px 20px;
    }
  }
}
</style>
