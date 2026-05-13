<template>
  <div class="role-manage">
    <!-- 角色列表 -->
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <span class="title">角色管理</span>
          <el-button type="primary" icon="Plus" @click="handleAdd">新增角色</el-button>
        </div>
      </template>

      <el-table
        v-loading="loading"
        :data="roleList"
        border
        stripe
        style="width: 100%"
      >
        <el-table-column prop="role_name" label="角色名称" width="160" />
        <el-table-column prop="role_code" label="角色编码" width="160" />
        <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip />
        <el-table-column label="操作" width="240" align="center" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleEdit(row)">
              编辑
            </el-button>
            <el-button type="warning" link size="small" @click="handleAssignPermission(row)">
              分配权限
            </el-button>
            <el-button type="danger" link size="small" @click="handleDelete(row)">
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 新增/编辑弹窗 -->
    
<el-dialog v-dialog-drag
      v-model="formDialogVisible"
      :title="isEdit ? '编辑角色' : '新增角色'"
      width="520px"
      destroy-on-close
    >
      <el-form
        ref="roleFormRef"
        :model="roleForm"
        :rules="roleFormRules"
        label-width="80px"
      >
        <el-form-item label="角色名称" prop="role_name">
          <el-input v-model="roleForm.role_name" placeholder="请输入角色名称" />
        </el-form-item>
        <el-form-item label="角色编码" prop="role_code">
          <el-input
            v-model="roleForm.role_code"
            placeholder="请输入角色编码（如 admin）"
            :disabled="isEdit"
          />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input
            v-model="roleForm.description"
            type="textarea"
            :rows="3"
            placeholder="请输入角色描述"
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

    <!-- 权限分配弹窗 -->
    
<el-dialog v-dialog-drag
      v-model="permDialogVisible"
      title="分配权限"
      width="560px"
      destroy-on-close
    >
      <div class="perm-assign-info">
        <span>角色：</span>
        <strong>{{ currentRole?.role_name }}</strong>
        <el-tag size="small" type="info" style="margin-left: 8px">{{ currentRole?.role_code }}</el-tag>
      </div>
      <el-divider />
      <div v-loading="permTreeLoading" class="perm-tree-wrapper">
        <el-tree
          ref="permTreeRef"
          :data="permissionTree"
          :props="treeProps"
          show-checkbox
          node-key="id"
          default-expand-all
          check-strictly
          :default-checked-keys="checkedPermIds"
        >
          <template #default="{ node, data }">
            <div class="tree-node">
              <span class="node-label">{{ data.permission_name }}</span>
              <el-tag size="small" :type="getPermTypeTag(data.permission_type)" class="node-tag">
                {{ getPermTypeLabel(data.permission_type) }}
              </el-tag>
            </div>
          </template>
        </el-tree>
      </div>
      <template #footer>
        <el-button @click="permDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="permSubmitLoading" @click="handlePermSubmit">
          确定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, nextTick } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getRoleList,
  createRole,
  updateRole,
  deleteRole,
  assignPermissions,
  getRolePermissions,
} from '@/api/role'
import { getPermissionTree } from '@/api/permission'

// ==================== 角色列表 ====================
const loading = ref(false)
const roleList = ref([])

const fetchRoleList = async () => {
  loading.value = true
  try {
    const res = await getRoleList()
    roleList.value = res.data || []
  } catch (error) {
    console.error('获取角色列表失败:', error)
  } finally {
    loading.value = false
  }
}

// ==================== 新增/编辑 ====================
const formDialogVisible = ref(false)
const isEdit = ref(false)
const submitLoading = ref(false)
const roleFormRef = ref(null)

const roleForm = reactive({
  id: null,
  role_name: '',
  role_code: '',
  description: '',
})

const roleFormRules = {
  role_name: [{ required: true, message: '请输入角色名称', trigger: 'blur' }],
  role_code: [
    { required: true, message: '请输入角色编码', trigger: 'blur' },
    { pattern: /^[a-zA-Z_][a-zA-Z0-9_]*$/, message: '编码只能包含字母、数字和下划线', trigger: 'blur' },
  ],
}

const resetRoleForm = () => {
  roleForm.id = null
  roleForm.role_name = ''
  roleForm.role_code = ''
  roleForm.description = ''
}

const handleAdd = () => {
  isEdit.value = false
  resetRoleForm()
  formDialogVisible.value = true
}

const handleEdit = (row) => {
  isEdit.value = true
  roleForm.id = row.id
  roleForm.role_name = row.role_name
  roleForm.role_code = row.role_code
  roleForm.description = row.description
  formDialogVisible.value = true
}

const handleSubmit = async () => {
  const valid = await roleFormRef.value?.validate().catch(() => false)
  if (!valid) return

  submitLoading.value = true
  try {
    const data = {
      role_name: roleForm.role_name,
      role_code: roleForm.role_code,
      description: roleForm.description,
    }
    if (isEdit.value) {
      await updateRole(roleForm.id, data)
      ElMessage.success('更新成功')
    } else {
      await createRole(data)
      ElMessage.success('创建成功')
    }
    formDialogVisible.value = false
    fetchRoleList()
  } catch (error) {
    console.error('保存角色失败:', error)
  } finally {
    submitLoading.value = false
  }
}

// ==================== 权限分配 ====================
const permDialogVisible = ref(false)
const permSubmitLoading = ref(false)
const permTreeLoading = ref(false)
const currentRole = ref(null)
const permissionTree = ref([])
const checkedPermIds = ref([])
const permTreeRef = ref(null)

const treeProps = {
  label: 'permission_name',
  children: 'children',
}

const getPermTypeLabel = (type) => {
  const map = { 1: '菜单', 2: '按钮', 3: '接口' }
  return map[type] || '未知'
}

const getPermTypeTag = (type) => {
  const map = { 1: '', 2: 'success', 3: 'warning' }
  return map[type] || 'info'
}

const handleAssignPermission = async (row) => {
  currentRole.value = row
  permDialogVisible.value = true
  permTreeLoading.value = true
  checkedPermIds.value = []

  try {
    // 并行获取权限树和角色已有权限
    const [treeRes, permRes] = await Promise.all([
      getPermissionTree(),
      getRolePermissions(row.id),
    ])
    permissionTree.value = treeRes.data || []
    // 从权限对象数组中提取 ID 数组
    checkedPermIds.value = (permRes.data || []).map(p => p.id)
  } catch (error) {
    console.error('获取权限数据失败:', error)
  } finally {
    permTreeLoading.value = false
  }
}

const handlePermSubmit = async () => {
  permSubmitLoading.value = true
  try {
    const checkedKeys = permTreeRef.value?.getCheckedKeys() || []
    const halfCheckedKeys = permTreeRef.value?.getHalfCheckedKeys() || []
    const allKeys = [...checkedKeys, ...halfCheckedKeys]
    await assignPermissions(currentRole.value.id, allKeys)
    ElMessage.success('权限分配成功')
    permDialogVisible.value = false
  } catch (error) {
    console.error('权限分配失败:', error)
  } finally {
    permSubmitLoading.value = false
  }
}

// ==================== 删除 ====================
const handleDelete = (row) => {
  ElMessageBox.confirm(
    `确定要删除角色 "${row.role_name}" 吗？删除后该角色下的用户将失去对应权限。`,
    '删除确认',
    {
      confirmButtonText: '确定删除',
      cancelButtonText: '取消',
      type: 'warning',
    }
  ).then(async () => {
    try {
      await deleteRole(row.id)
      ElMessage.success('删除成功')
      fetchRoleList()
    } catch (error) {
      console.error('删除角色失败:', error)
    }
  }).catch(() => {})
}

// ==================== 初始化 ====================
onMounted(() => {
  fetchRoleList()
})
</script>

<style lang="scss" scoped>
.role-manage {
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

  .perm-assign-info {
    font-size: 14px;
    color: #606266;
  }

  .perm-tree-wrapper {
    max-height: 400px;
    overflow-y: auto;
    border: 1px solid #ebeef5;
    border-radius: 4px;
    padding: 8px;
  }

  .tree-node {
    display: flex;
    align-items: center;
    justify-content: space-between;
    flex: 1;
    padding-right: 8px;

    .node-label {
      font-size: 14px;
    }

    .node-tag {
      margin-left: 8px;
    }
  }
}
</style>
