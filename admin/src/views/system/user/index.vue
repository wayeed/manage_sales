<template>
  <div class="user-manage">
    <!-- 搜索栏 -->
    <el-card class="search-card" shadow="never">
      <el-form :model="searchForm" inline>
        <el-form-item label="关键词">
          <el-input
            v-model="searchForm.keyword"
            placeholder="姓名/手机号/工号"
            clearable
            style="width: 200px"
            @keyup.enter="handleSearch"
          />
        </el-form-item>
        <el-form-item label="门店">
          <el-select
            v-model="searchForm.store_id"
            placeholder="全部门店"
            clearable
            style="width: 160px"
          >
            <el-option
              v-for="s in storeOptions"
              :key="s.id"
              :label="s.store_name"
              :value="s.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="角色">
          <el-select
            v-model="searchForm.role_id"
            placeholder="全部角色"
            clearable
            style="width: 160px"
          >
            <el-option
              v-for="role in roleOptions"
              :key="role.id"
              :label="role.role_name"
              :value="role.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select
            v-model="searchForm.status"
            placeholder="全部状态"
            clearable
            style="width: 120px"
          >
            <el-option label="启用" :value="1" />
            <el-option label="禁用" :value="0" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" icon="Search" @click="handleSearch">搜索</el-button>
          <el-button icon="Refresh" @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 数据表格 -->
    <el-card class="table-card" shadow="never">
      <template #header>
        <div class="card-header">
          <span class="title">用户列表</span>
          <el-button type="primary" icon="Plus" @click="handleAdd">新增用户</el-button>
        </div>
      </template>

      <el-table
        v-loading="loading"
        :data="userList"
        border
        stripe
        style="width: 100%"
      >
        <el-table-column prop="employee_no" label="工号" width="100" align="center" />
        <el-table-column prop="username" label="用户名" width="110" />
        <el-table-column prop="real_name" label="姓名" width="100" />
        <el-table-column prop="phone" label="手机号" width="130" />
        <el-table-column label="门店" width="120">
          <template #default="{ row }">
            {{ row.store?.store_name || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="角色" min-width="140">
          <template #default="{ row }">
            <el-tag
              v-for="role in row.roles"
              :key="role.id"
              size="small"
              class="role-tag"
            >
              {{ role.role_name }}
            </el-tag>
            <span v-if="!row.roles || row.roles.length === 0" class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column label="等级" width="90" align="center">
          <template #default="{ row }">
            <el-tag :type="getLevelTag(row.level)" size="small">
              {{ getLevelLabel(row.level) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="转正" width="70" align="center">
          <template #default="{ row }">
            <el-tag :type="row.is_formal === 1 ? 'success' : 'info'" size="small">
              {{ row.is_formal === 1 ? '是' : '否' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="70" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="入职日期" width="110" align="center">
          <template #default="{ row }">
            {{ row.entry_date ? row.entry_date.substring(0, 10) : '-' }}
          </template>
        </el-table-column>
        <el-table-column label="引荐人" width="100" align="center">
          <template #default="{ row }">
            {{ row.referrer?.real_name || row.referrer?.username || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="320" align="center" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleEdit(row)">
              编辑
            </el-button>
            <el-button type="warning" link size="small" @click="handleAssignRole(row)">
              分配角色
            </el-button>
            <el-button
              :type="row.status === 1 ? 'danger' : 'success'"
              link
              size="small"
              @click="handleToggleStatus(row)"
            >
              {{ row.status === 1 ? '禁用' : '启用' }}
            </el-button>
            <el-button type="info" link size="small" @click="handleResetPassword(row)">
              重置密码
            </el-button>
            <el-button type="danger" link size="small" @click="handleDelete(row)">
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.page_size"
          :page-sizes="[10, 20, 50, 100]"
          :total="pagination.total"
          layout="total, sizes, prev, pager, next, jumper"
          background
          @size-change="fetchUserList"
          @current-change="fetchUserList"
        />
      </div>
    </el-card>

    <!-- 新增/编辑弹窗 -->
    <el-dialog
      v-model="formDialogVisible"
      :title="isEdit ? '编辑用户' : '新增用户'"
      width="680px"
      destroy-on-close
    >
      <el-form
        ref="userFormRef"
        :model="userForm"
        :rules="userFormRules"
        label-width="90px"
      >
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="用户名" prop="username">
              <el-input v-model="userForm.username" placeholder="请输入用户名" :disabled="isEdit" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="工号" prop="employee_no">
              <el-input v-model="userForm.employee_no" placeholder="请输入工号" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="姓名" prop="real_name">
              <el-input v-model="userForm.real_name" placeholder="请输入姓名" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="手机号" prop="phone">
              <el-input v-model="userForm.phone" placeholder="请输入手机号" :disabled="isEdit" maxlength="11" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="邮箱">
              <el-input v-model="userForm.email" placeholder="请输入邮箱" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="所属门店">
              <el-select v-model="userForm.store_id" placeholder="请选择门店" clearable style="width: 100%">
                <el-option
                  v-for="s in storeOptions"
                  :key="s.id"
                  :label="s.store_name"
                  :value="s.id"
                />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="入职日期">
              <el-date-picker
                v-model="userForm.entry_date"
                type="date"
                placeholder="选择日期"
                value-format="YYYY-MM-DD"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="试用期至">
              <el-date-picker
                v-model="userForm.probation_end_date"
                type="date"
                placeholder="选择日期"
                value-format="YYYY-MM-DD"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="等级">
              <el-select v-model="userForm.level" placeholder="请选择等级" style="width: 100%">
                <el-option label="初级业务员" :value="1" />
                <el-option label="中级业务员" :value="2" />
                <el-option label="高级业务员" :value="3" />
              </el-select>
              <div class="level-remark">{{ levelRemark }}</div>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="基本工资">
              <el-input-number v-model="userForm.base_salary" :min="0" :precision="2" placeholder="0.00" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="是否转正">
              <el-radio-group v-model="userForm.is_formal">
                <el-radio :value="1">是</el-radio>
                <el-radio :value="0">否</el-radio>
              </el-radio-group>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="状态">
              <el-radio-group v-model="userForm.status">
                <el-radio :value="1">启用</el-radio>
                <el-radio :value="0">禁用</el-radio>
              </el-radio-group>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="上级主管">
              <el-select v-model="userForm.parent_id" placeholder="请选择上级" clearable style="width: 100%">
                <el-option
                  v-for="u in userList"
                  :key="u.id"
                  :label="u.real_name || u.username"
                  :value="u.id"
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="引荐人">
              <el-select v-model="userForm.referrer_id" placeholder="请选择引荐人" clearable style="width: 100%">
                <el-option
                  v-for="u in userList"
                  :key="u.id"
                  :label="u.real_name || u.username"
                  :value="u.id"
                />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="角色">
              <el-select v-model="userForm.role_ids" multiple placeholder="请选择角色" style="width: 100%">
                <el-option
                  v-for="role in roleOptions"
                  :key="role.id"
                  :label="role.role_name"
                  :value="role.id"
                />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-divider content-position="left">银行信息</el-divider>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="身份证号">
              <el-input v-model="userForm.id_card" placeholder="请输入身份证号" maxlength="18" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="银行卡号">
              <el-input v-model="userForm.bank_account" placeholder="请输入银行卡号" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="开户银行">
              <el-input v-model="userForm.bank_name" placeholder="请输入开户银行" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item v-if="!isEdit" label="密码" prop="password">
          <el-input
            v-model="userForm.password"
            type="password"
            placeholder="请输入密码（至少6位）"
            show-password
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmit">
          确定
        </el-button>
      </template>
    </el-dialog>

    <!-- 角色分配弹窗 -->
    <el-dialog
      v-model="roleDialogVisible"
      title="分配角色"
      width="480px"
      destroy-on-close
    >
      <div class="role-assign-info">
        <span>用户：</span>
        <strong>{{ currentUser?.real_name }}</strong>
      </div>
      <el-divider />
      <el-checkbox-group v-model="assignedRoleIds">
        <el-checkbox
          v-for="role in roleOptions"
          :key="role.id"
          :value="role.id"
          class="role-checkbox"
        >
          <div>
            <div class="role-name">{{ role.role_name }}</div>
            <div class="role-desc">{{ role.description || '暂无描述' }}</div>
          </div>
        </el-checkbox>
      </el-checkbox-group>
      <template #footer>
        <el-button @click="roleDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="roleSubmitLoading" @click="handleRoleSubmit">
          确定
        </el-button>
      </template>
    </el-dialog>

    <!-- 密码重置确认弹窗 -->
    <el-dialog
      v-model="resetPwdDialogVisible"
      title="重置密码"
      width="400px"
      destroy-on-close
    >
      <div class="reset-pwd-content">
        <el-icon :size="48" color="#E6A23C"><WarningFilled /></el-icon>
        <p>确定要重置用户 <strong>{{ currentUser?.real_name }}</strong> 的密码吗？</p>
        <p class="text-muted">重置后密码将恢复为默认密码</p>
      </div>
      <template #footer>
        <el-button @click="resetPwdDialogVisible = false">取消</el-button>
        <el-button type="warning" :loading="resetPwdLoading" @click="confirmResetPassword">
          确认重置
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getUserList,
  createUser,
  updateUser,
  deleteUser,
  resetPassword,
  assignRole,
  updateUserStatus,
} from '@/api/user'
import { getRoleList } from '@/api/role'
import { getStoreList } from '@/api/store'
import { getConfigList } from '@/api/config'

// ==================== 搜索与列表 ====================
const loading = ref(false)
const userList = ref([])
const roleOptions = ref([])
const storeOptions = ref([])

const searchForm = reactive({
  keyword: '',
  store_id: '',
  role_id: '',
  status: '',
})

const pagination = reactive({
  page: 1,
  page_size: 20,
  total: 0,
})

const fetchUserList = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      page_size: pagination.page_size,
    }
    if (searchForm.keyword) params.keyword = searchForm.keyword
    if (searchForm.store_id) params.store_id = searchForm.store_id
    if (searchForm.role_id) params.role_id = searchForm.role_id
    if (searchForm.status !== '' && searchForm.status !== null) params.status = searchForm.status

    const res = await getUserList(params)
    userList.value = res.data?.list || res.data || []
    pagination.total = res.data?.total || 0
  } catch (error) {
    console.error('获取用户列表失败:', error)
  } finally {
    loading.value = false
  }
}

const fetchRoleOptions = async () => {
  try {
    const res = await getRoleList()
    roleOptions.value = res.data || []
  } catch (error) {
    console.error('获取角色列表失败:', error)
  }
}

const fetchStoreOptions = async () => {
  try {
    const res = await getStoreList()
    storeOptions.value = res.data || []
  } catch (error) {
    console.error('获取门店列表失败:', error)
  }
}

// 等级标签和标签类型
const getLevelLabel = (level) => {
  const map = { 1: '初级', 2: '中级', 3: '高级' }
  return map[level] || '初级'
}

const getLevelTag = (level) => {
  const map = { 1: 'info', 2: 'warning', 3: 'success' }
  return map[level] || 'info'
}

// 等级配置（从系统配置读取）
const levelConfigs = ref({})

const fetchLevelConfigs = async () => {
  try {
    const res = await getConfigList()
    const configs = res.data || []
    const configMap = {}
    configs.forEach(c => { configMap[c.config_key] = c.config_value })
    levelConfigs.value = configMap
  } catch (e) {
    console.error('获取系统配置失败:', e)
  }
}

// 等级备注（包含底薪建议和提成比例，从系统配置读取）
const levelRemark = computed(() => {
  const level = userForm.level
  const single = levelConfigs.value[`commission_rate_level${level}_single`]
  const multi = levelConfigs.value[`commission_rate_level${level}_multi`]
  const remark = levelConfigs.value[`commission_rate_level${level}_remark`]

  const rateText = single && multi
    ? `单品${(Number(single) * 100).toFixed(0)}% / 多品${(Number(multi) * 100).toFixed(0)}%`
    : ''
  const remarkText = remark || ''

  return [remarkText, rateText].filter(Boolean).join(' | ')
})

const handleSearch = () => {
  pagination.page = 1
  fetchUserList()
}

const handleReset = () => {
  searchForm.keyword = ''
  searchForm.store_id = ''
  searchForm.role_id = ''
  searchForm.status = ''
  pagination.page = 1
  fetchUserList()
}

// ==================== 新增/编辑 ====================
const formDialogVisible = ref(false)
const isEdit = ref(false)
const submitLoading = ref(false)
const userFormRef = ref(null)
const currentUser = ref(null)

const userForm = reactive({
  id: null,
  username: '',
  employee_no: '',
  real_name: '',
  phone: '',
  email: '',
  store_id: null,
  department_id: null,
  entry_date: '',
  probation_end_date: '',
  is_formal: 0,
  level: 1, // 默认初级业务员
  parent_id: null,
  referrer_id: null,
  base_salary: 0,
  id_card: '',
  bank_account: '',
  bank_name: '',
  password: '',
  role_ids: [],
  status: 1,
})

const userFormRules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  real_name: [{ required: true, message: '请输入姓名', trigger: 'blur' }],
  phone: [
    { required: true, message: '请输入手机号', trigger: 'blur' },
    { pattern: /^1[3-9]\d{9}$/, message: '请输入正确的手机号', trigger: 'blur' },
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码长度不能少于6位', trigger: 'blur' },
  ],
}

const resetUserForm = () => {
  userForm.id = null
  userForm.username = ''
  userForm.employee_no = ''
  userForm.real_name = ''
  userForm.phone = ''
  userForm.email = ''
  userForm.store_id = null
  userForm.department_id = null
  userForm.entry_date = ''
  userForm.probation_end_date = ''
  userForm.is_formal = 0
  userForm.level = 1 // 默认初级业务员
  userForm.parent_id = null
  userForm.referrer_id = null
  userForm.base_salary = 0
  userForm.id_card = ''
  userForm.bank_account = ''
  userForm.bank_name = ''
  userForm.password = ''
  userForm.role_ids = []
  userForm.status = 1
}

const handleAdd = () => {
  isEdit.value = false
  resetUserForm()
  formDialogVisible.value = true
}

const handleEdit = (row) => {
  isEdit.value = true
  currentUser.value = row
  userForm.id = row.id
  userForm.username = row.username || ''
  userForm.employee_no = row.employee_no || ''
  userForm.real_name = row.real_name || ''
  userForm.phone = row.phone || ''
  userForm.email = row.email || ''
  userForm.store_id = row.store_id || null
  userForm.department_id = row.department_id || null
  userForm.entry_date = row.entry_date ? row.entry_date.substring(0, 10) : ''
  userForm.probation_end_date = row.probation_end_date ? row.probation_end_date.substring(0, 10) : ''
  userForm.is_formal = row.is_formal || 0
  userForm.level = row.level || 1
  userForm.parent_id = row.parent_id || null
  userForm.referrer_id = row.referrer_id || null
  userForm.base_salary = row.base_salary || 0
  userForm.id_card = row.id_card || ''
  userForm.bank_account = row.bank_account || ''
  userForm.bank_name = row.bank_name || ''
  userForm.password = ''
  userForm.role_ids = row.roles?.map((r) => r.id) || []
  userForm.status = row.status
  formDialogVisible.value = true
}

const handleSubmit = async () => {
  const valid = await userFormRef.value?.validate().catch(() => false)
  if (!valid) return

  submitLoading.value = true
  try {
    const data = {
      username: userForm.username,
      employee_no: userForm.employee_no,
      real_name: userForm.real_name,
      phone: userForm.phone,
      email: userForm.email,
      store_id: userForm.store_id || null,
      department_id: userForm.department_id || null,
      entry_date: userForm.entry_date,
      probation_end_date: userForm.probation_end_date,
      is_formal: userForm.is_formal,
      level: userForm.level,
      parent_id: userForm.parent_id || null,
      referrer_id: userForm.referrer_id || null,
      base_salary: userForm.base_salary,
      id_card: userForm.id_card,
      bank_account: userForm.bank_account,
      bank_name: userForm.bank_name,
      role_ids: userForm.role_ids,
      status: userForm.status,
    }
    if (isEdit.value) {
      await updateUser(userForm.id, data)
      ElMessage.success('更新成功')
    } else {
      data.password = userForm.password
      await createUser(data)
      ElMessage.success('创建成功')
    }
    formDialogVisible.value = false
    fetchUserList()
  } catch (error) {
    console.error('保存用户失败:', error)
  } finally {
    submitLoading.value = false
  }
}

// ==================== 角色分配 ====================
const roleDialogVisible = ref(false)
const roleSubmitLoading = ref(false)
const assignedRoleIds = ref([])

const handleAssignRole = (row) => {
  currentUser.value = row
  assignedRoleIds.value = row.roles?.map((r) => r.id) || []
  roleDialogVisible.value = true
}

const handleRoleSubmit = async () => {
  roleSubmitLoading.value = true
  try {
    await assignRole(currentUser.value.id, assignedRoleIds.value)
    ElMessage.success('角色分配成功')
    roleDialogVisible.value = false
    fetchUserList()
  } catch (error) {
    console.error('角色分配失败:', error)
  } finally {
    roleSubmitLoading.value = false
  }
}

// ==================== 密码重置 ====================
const resetPwdDialogVisible = ref(false)
const resetPwdLoading = ref(false)

const handleResetPassword = (row) => {
  currentUser.value = row
  resetPwdDialogVisible.value = true
}

const confirmResetPassword = async () => {
  resetPwdLoading.value = true
  try {
    await resetPassword(currentUser.value.id)
    ElMessage.success('密码重置成功')
    resetPwdDialogVisible.value = false
  } catch (error) {
    console.error('密码重置失败:', error)
  } finally {
    resetPwdLoading.value = false
  }
}

// ==================== 启用/禁用 ====================
const handleToggleStatus = (row) => {
  const action = row.status === 1 ? '禁用' : '启用'
  ElMessageBox.confirm(
    `确定要${action}用户 "${row.real_name}" 吗？`,
    `${action}确认`,
    {
      confirmButtonText: `确定${action}`,
      cancelButtonText: '取消',
      type: 'warning',
    }
  ).then(async () => {
    try {
      await updateUserStatus(row.id, row.status === 1 ? 0 : 1)
      ElMessage.success(`${action}成功`)
      fetchUserList()
    } catch (error) {
      console.error(`${action}用户失败:`, error)
    }
  }).catch(() => {})
}

// ==================== 删除 ====================
const handleDelete = (row) => {
  ElMessageBox.confirm(
    `确定要删除用户 "${row.real_name}" 吗？此操作不可恢复。`,
    '删除确认',
    {
      confirmButtonText: '确定删除',
      cancelButtonText: '取消',
      type: 'warning',
    }
  ).then(async () => {
    try {
      await deleteUser(row.id)
      ElMessage.success('删除成功')
      fetchUserList()
    } catch (error) {
      console.error('删除用户失败:', error)
    }
  }).catch(() => {})
}

// ==================== 初始化 ====================
onMounted(() => {
  fetchUserList()
  fetchRoleOptions()
  fetchStoreOptions()
  fetchLevelConfigs()
})
</script>

<style lang="scss" scoped>
.user-manage {
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

  .table-card {
    .pagination-wrapper {
      display: flex;
      justify-content: flex-end;
      margin-top: 16px;
    }
  }

  .role-tag {
    margin-right: 4px;
    margin-bottom: 2px;
  }

  .text-muted {
    color: #909399;
    font-size: 13px;
  }

  .role-assign-info {
    font-size: 14px;
    color: #606266;
    margin-bottom: 8px;
  }

  .role-checkbox {
    width: 100%;
    margin-bottom: 12px;

    .role-name {
      font-weight: 500;
      color: #303133;
    }

    .role-desc {
      font-size: 12px;
      color: #909399;
      margin-top: 2px;
    }
  }

  .reset-pwd-content {
    text-align: center;
    padding: 16px 0;

    p {
      margin-top: 12px;
      font-size: 15px;
      color: #303133;
    }
  }

  .level-remark {
    margin-top: 4px;
    font-size: 12px;
    color: #909399;
    line-height: 1.4;
  }
}
</style>
