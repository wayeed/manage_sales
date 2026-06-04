<template>
  <div class="app-version-manage">
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <span class="title">APP版本管理</span>
          <el-button type="primary" @click="handleAdd">
            <el-icon><Plus /></el-icon>新增版本
          </el-button>
        </div>
      </template>

      <!-- 筛选条件 -->
      <el-form :inline="true" :model="queryParams" class="search-form">
        <el-form-item label="平台">
          <el-select v-model="queryParams.platform" placeholder="全部平台" clearable style="width: 120px">
            <el-option label="iOS" value="ios" />
            <el-option label="Android" value="android" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="queryParams.status" placeholder="全部状态" clearable style="width: 120px">
            <el-option label="启用" :value="1" />
            <el-option label="禁用" :value="0" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">查询</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>

      <!-- 数据表格 -->
      <el-table :data="tableData" v-loading="loading" stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="platform" label="平台" width="100">
          <template #default="{ row }">
            <el-tag :type="row.platform === 'ios' ? 'success' : 'primary'">
              {{ row.platform === 'ios' ? 'iOS' : 'Android' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="version_code" label="版本号" width="120" />
        <el-table-column prop="version_name" label="版本名称" />
        <el-table-column prop="update_type" label="更新类型" width="100">
          <template #default="{ row }">
            <el-tag :type="row.update_type === 'wgt' ? 'warning' : 'primary'" size="small">
              {{ row.update_type === 'wgt' ? '热更新' : '整包更新' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="file_size" label="文件大小" width="120">
          <template #default="{ row }">
            {{ formatFileSize(row.file_size) }}
          </template>
        </el-table-column>
        <el-table-column prop="is_force_update" label="强制更新" width="100">
          <template #default="{ row }">
            <el-tag :type="row.is_force_update ? 'danger' : 'info'" size="small">
              {{ row.is_force_update ? '是' : '否' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="handleEdit(row)">编辑</el-button>
            <el-button link type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-container">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.page_size"
          :page-sizes="[10, 20, 50, 100]"
          :total="pagination.total"
          layout="total, sizes, prev, pager, next"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </el-card>

    <!-- 新增/编辑弹窗 -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? '编辑版本' : '新增版本'"
      width="600px"
      destroy-on-close
    >
      <el-form
        ref="formRef"
        :model="formData"
        :rules="formRules"
        label-width="100px"
      >
        <el-form-item label="平台" prop="platform">
          <el-select v-model="formData.platform" style="width: 100%" :disabled="isEdit">
            <el-option label="iOS" value="ios" />
            <el-option label="Android" value="android" />
          </el-select>
        </el-form-item>
        <el-form-item label="版本号" prop="version_code">
          <el-input v-model="formData.version_code" placeholder="如: 1.0.0" :disabled="isEdit" />
        </el-form-item>
        <el-form-item label="版本名称" prop="version_name">
          <el-input v-model="formData.version_name" placeholder="如: v1.0.0正式版" />
        </el-form-item>
        <el-form-item label="下载地址" prop="download_url">
          <el-input v-model="formData.download_url" placeholder="安装包下载链接" />
        </el-form-item>
        <el-form-item label="更新类型" prop="update_type">
          <el-radio-group v-model="formData.update_type">
            <el-radio label="full">整包更新</el-radio>
            <el-radio label="wgt">热更新(wgt)</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="文件大小">
          <el-input-number v-model="formData.file_size" :min="0" style="width: 100%" placeholder="字节" />
          <div class="form-tip">填写下载地址后，文件大小将自动获取</div>
        </el-form-item>
        <el-form-item label="强制更新">
          <el-switch v-model="formData.is_force_update" :active-value="1" :inactive-value="0" />
        </el-form-item>
        <el-form-item label="最低版本">
          <el-input v-model="formData.min_version" placeholder="如: 1.0.0" />
        </el-form-item>
        <el-form-item label="更新内容">
          <el-input
            v-model="formData.update_content"
            type="textarea"
            :rows="4"
            placeholder="每行一条更新内容"
          />
        </el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="formData.status">
            <el-radio :label="1">启用</el-radio>
            <el-radio :label="0">禁用</el-radio>
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
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import {
  getAppVersionList,
  createAppVersion,
  updateAppVersion,
  deleteAppVersion,
} from '@/api/app-version'

// ==================== 数据 ====================
const loading = ref(false)
const tableData = ref([])
const queryParams = reactive({
  platform: '',
  status: null,
})
const pagination = reactive({
  page: 1,
  page_size: 10,
  total: 0,
})

// ==================== 弹窗表单 ====================
const dialogVisible = ref(false)
const isEdit = ref(false)
const submitLoading = ref(false)
const formRef = ref()
const formData = reactive({
  id: null,
  platform: 'android',
  version_code: '',
  version_name: '',
  download_url: '',
  file_size: 0,
  update_type: 'full',
  is_force_update: 0,
  min_version: '',
  update_content: '',
  status: 1,
})

const formRules = {
  platform: [{ required: true, message: '请选择平台', trigger: 'change' }],
  version_code: [{ required: true, message: '请输入版本号', trigger: 'blur' }],
  version_name: [{ required: true, message: '请输入版本名称', trigger: 'blur' }],
  download_url: [
    { required: true, message: '请输入下载地址', trigger: 'blur' },
    { type: 'url', message: '请输入有效的URL', trigger: 'blur' },
  ],
  update_type: [{ required: true, message: '请选择更新类型', trigger: 'change' }],
}

// ==================== 方法 ====================
const fetchList = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      page_size: pagination.page_size,
      ...queryParams,
    }
    const res = await getAppVersionList(params)
    if (res.data) {
      tableData.value = res.data.list || []
      pagination.total = res.data.total || 0
    }
  } catch (error) {
    console.error('获取版本列表失败:', error)
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  pagination.page = 1
  fetchList()
}

const handleReset = () => {
  queryParams.platform = ''
  queryParams.status = null
  pagination.page = 1
  fetchList()
}

const handleSizeChange = (val) => {
  pagination.page_size = val
  fetchList()
}

const handleCurrentChange = (val) => {
  pagination.page = val
  fetchList()
}

const resetForm = () => {
  formData.id = null
  formData.platform = 'android'
  formData.version_code = ''
  formData.version_name = ''
  formData.download_url = ''
  formData.file_size = 0
  formData.update_type = 'full'
  formData.is_force_update = 0
  formData.min_version = ''
  formData.update_content = ''
  formData.status = 1
}

const handleAdd = () => {
  isEdit.value = false
  resetForm()
  dialogVisible.value = true
}

const handleEdit = (row) => {
  isEdit.value = true
  Object.assign(formData, row)
  dialogVisible.value = true
}

const handleSubmit = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  submitLoading.value = true
  try {
    if (isEdit.value) {
      await updateAppVersion(formData.id, formData)
      ElMessage.success('更新成功')
    } else {
      await createAppVersion(formData)
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    fetchList()
  } catch (error) {
    console.error('提交失败:', error)
  } finally {
    submitLoading.value = false
  }
}

const handleDelete = (row) => {
  ElMessageBox.confirm(`确定要删除版本 ${row.version_code} 吗？`, '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning',
  }).then(async () => {
    try {
      await deleteAppVersion(row.id)
      ElMessage.success('删除成功')
      fetchList()
    } catch (error) {
      console.error('删除失败:', error)
    }
  }).catch(() => {})
}

// ==================== 工具函数 ====================
const formatFileSize = (bytes) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const formatDate = (date) => {
  if (!date) return '-'
  return new Date(date).toLocaleString('zh-CN')
}

onMounted(() => {
  fetchList()
})
</script>

<style lang="scss" scoped>
.app-version-manage {
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;

    .title {
      font-size: 16px;
      font-weight: 600;
    }
  }

  .search-form {
    margin-bottom: 20px;
  }

  .pagination-container {
    margin-top: 20px;
    display: flex;
    justify-content: flex-end;
  }

  .form-tip {
    font-size: 12px;
    color: #909399;
    margin-top: 4px;
  }
}
</style>
