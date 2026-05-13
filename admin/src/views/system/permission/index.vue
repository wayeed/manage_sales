<template>
  <div class="permission-manage">
    <!-- 权限列表 -->
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <span class="title">权限管理</span>
          <div class="header-actions">
            <el-radio-group v-model="viewMode" size="small">
              <el-radio-button value="tree">树形视图</el-radio-button>
              <el-radio-button value="table">列表视图</el-radio-button>
            </el-radio-group>
          </div>
        </div>
      </template>

      <!-- 树形视图 -->
      <div v-if="viewMode === 'tree'" v-loading="loading" class="perm-tree-wrapper">
        <el-table
          :data="permissionTree"
          border
          row-key="id"
          :tree-props="{ children: 'children' }"
          default-expand-all
          style="width: 100%"
        >
          <el-table-column prop="permission_name" label="权限名称" min-width="200" />
          <el-table-column prop="permission_code" label="权限编码" min-width="200" show-overflow-tooltip />
          <el-table-column label="类型" width="100" align="center">
            <template #default="{ row }">
              <el-tag :type="getPermTypeTag(row.permission_type)" size="small">
                {{ getPermTypeLabel(row.permission_type) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="path" label="路由路径" min-width="180" show-overflow-tooltip>
            <template #default="{ row }">
              {{ row.path || '-' }}
            </template>
          </el-table-column>
          <el-table-column label="状态" width="80" align="center">
            <template #default="{ row }">
              <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">
                {{ row.status === 1 ? '启用' : '禁用' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="sort_order" label="排序" width="80" align="center" />
        </el-table>
      </div>

      <!-- 列表视图 -->
      <div v-else v-loading="loading">
        <el-table
          :data="permissionList"
          border
          stripe
          style="width: 100%"
        >
          <el-table-column prop="permission_name" label="权限名称" min-width="180" />
          <el-table-column prop="permission_code" label="权限编码" min-width="200" show-overflow-tooltip />
          <el-table-column label="类型" width="100" align="center">
            <template #default="{ row }">
              <el-tag :type="getPermTypeTag(row.permission_type)" size="small">
                {{ getPermTypeLabel(row.permission_type) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="parent_name" label="上级权限" width="140">
            <template #default="{ row }">
              {{ row.parent_name || '-' }}
            </template>
          </el-table-column>
          <el-table-column prop="path" label="路由路径" min-width="160" show-overflow-tooltip>
            <template #default="{ row }">
              {{ row.path || '-' }}
            </template>
          </el-table-column>
          <el-table-column label="状态" width="80" align="center">
            <template #default="{ row }">
              <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">
                {{ row.status === 1 ? '启用' : '禁用' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="sort_order" label="排序" width="80" align="center" />
        </el-table>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import { getPermissionList, getPermissionTree } from '@/api/permission'

const loading = ref(false)
const viewMode = ref('tree')
const permissionTree = ref([])
const permissionList = ref([])

const getPermTypeLabel = (type) => {
  const map = { 1: '菜单', 2: '按钮', 3: '接口' }
  return map[type] || '未知'
}

const getPermTypeTag = (type) => {
  const map = { 1: '', 2: 'success', 3: 'warning' }
  return map[type] || 'info'
}

const fetchPermissionTree = async () => {
  loading.value = true
  try {
    const res = await getPermissionTree()
    permissionTree.value = res.data || []
  } catch (error) {
    console.error('获取权限树失败:', error)
  } finally {
    loading.value = false
  }
}

const fetchPermissionList = async () => {
  loading.value = true
  try {
    const res = await getPermissionList()
    permissionList.value = res.data || []
  } catch (error) {
    console.error('获取权限列表失败:', error)
  } finally {
    loading.value = false
  }
}

const fetchData = () => {
  if (viewMode.value === 'tree') {
    fetchPermissionTree()
  } else {
    fetchPermissionList()
  }
}

watch(viewMode, () => {
  fetchData()
})

onMounted(() => {
  fetchData()
})
</script>

<style lang="scss" scoped>
.permission-manage {
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

  .perm-tree-wrapper {
    :deep(.el-table__indent) {
      padding-left: 16px !important;
    }
  }
}
</style>
