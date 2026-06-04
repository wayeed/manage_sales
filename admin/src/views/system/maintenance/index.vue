<template>
  <div class="system-maintenance">
    <!-- 数据备份卡片 -->
    <el-card class="maintenance-card" shadow="never">
      <template #header>
        <div class="card-header">
          <span class="title">数据备份</span>
          <el-button type="primary" :loading="backupLoading" @click="handleCreateBackup">
            <el-icon><Download /></el-icon>执行备份
          </el-button>
        </div>
      </template>

      <el-alert
        title="备份说明"
        type="info"
        :closable="false"
        description="执行数据库备份将生成SQL文件，可用于数据还原。建议定期备份以防数据丢失。"
        style="margin-bottom: 16px"
      />

      <!-- 备份列表 -->
      <el-table :data="backupList" v-loading="backupListLoading" stripe size="small">
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="file_name" label="文件名" />
        <el-table-column prop="file_size" label="大小" width="100">
          <template #default="{ row }">
            {{ formatFileSize(row.file_size) }}
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getBackupStatusType(row.status)" size="small">
              {{ getBackupStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="160">
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button
              link
              type="primary"
              :disabled="row.status !== 1"
              @click="handleRestoreBackup(row)"
            >
              还原
            </el-button>
            <el-button
              link
              type="danger"
              @click="handleDeleteBackup(row)"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-container">
        <el-pagination
          v-model:current-page="backupPagination.page"
          v-model:page-size="backupPagination.page_size"
          :page-sizes="[10, 20, 50]"
          :total="backupPagination.total"
          layout="total, sizes, prev, pager, next"
          @size-change="fetchBackupList"
          @current-change="fetchBackupList"
        />
      </div>
    </el-card>

    <!-- 清除业务数据卡片 -->
    <el-card class="maintenance-card" shadow="never">
      <template #header>
        <div class="card-header">
          <span class="title">清除业务数据</span>
          <el-button type="danger" :loading="clearLoading" @click="handleClearData">
            <el-icon><Delete /></el-icon>执行清除
          </el-button>
        </div>
      </template>

      <el-alert
        title="警告：此操作不可逆！"
        type="warning"
        :closable="false"
        description="清除业务数据前必须先执行数据备份（10分钟内），否则无法执行清除操作。"
        style="margin-bottom: 16px"
      />

      <!-- 备份状态检查 -->
      <div class="backup-check" style="margin-bottom: 16px">
        <el-tag :type="hasRecentBackup ? 'success' : 'danger'" size="large">
          {{ hasRecentBackup ? '✓ 10分钟内有备份，可以清除数据' : '✗ 10分钟内无备份，请先执行备份' }}
        </el-tag>
        <el-button link type="primary" @click="checkRecentBackup" style="margin-left: 12px">
          重新检查
        </el-button>
      </div>

      <!-- 数据表选择 -->
      <div class="table-selection">
        <div class="selection-header">
          <el-checkbox v-model="selectAll" @change="handleSelectAll">全选</el-checkbox>
          <span class="selection-count">已选择 {{ selectedCategories.length }} 项</span>
        </div>
        <el-checkbox-group v-model="selectedCategories" class="category-list">
          <el-card
            v-for="item in dataTableList"
            :key="item.category"
            class="category-card"
            :class="{ selected: selectedCategories.includes(item.category) }"
            shadow="never"
          >
            <el-checkbox :label="item.category">
              <div class="category-content">
                <div class="category-label">{{ item.label }}</div>
                <div class="category-desc">{{ item.description }}</div>
                <div class="category-tables">
                  <el-tag v-for="table in item.tables" :key="table" size="small" type="info">
                    {{ table }}
                  </el-tag>
                </div>
              </div>
            </el-checkbox>
          </el-card>
        </el-checkbox-group>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Download, Delete } from '@element-plus/icons-vue'
import {
  getDataTables,
  checkRecentBackup as checkRecentBackupAPI,
  getBackupList,
  createBackup,
  deleteBackup,
  restoreBackup,
  clearData,
} from '@/api/maintenance'

// ==================== 数据备份 ====================
const backupLoading = ref(false)
const backupListLoading = ref(false)
const backupList = ref([])
const backupPagination = reactive({
  page: 1,
  page_size: 10,
  total: 0,
})

const fetchBackupList = async () => {
  backupListLoading.value = true
  try {
    const res = await getBackupList({
      page: backupPagination.page,
      page_size: backupPagination.page_size,
    })
    if (res.data) {
      backupList.value = res.data.list || []
      backupPagination.total = res.data.total || 0
    }
  } catch (error) {
    console.error('获取备份列表失败:', error)
  } finally {
    backupListLoading.value = false
  }
}

const handleCreateBackup = async () => {
  backupLoading.value = true
  try {
    await createBackup({ backup_type: 'database' })
    ElMessage.success('备份任务已创建，请稍后刷新查看结果')
    setTimeout(fetchBackupList, 2000)
  } catch (error) {
    console.error('创建备份失败:', error)
  } finally {
    backupLoading.value = false
  }
}

const handleRestoreBackup = (row) => {
  ElMessageBox.confirm(
    `确定要使用备份 "${row.file_name}" 还原数据吗？<br><strong style="color: #f56c6c">警告：这将覆盖当前所有数据！</strong>`,
    '确认还原',
    {
      confirmButtonText: '确定还原',
      cancelButtonText: '取消',
      type: 'warning',
      dangerouslyUseHTMLString: true,
    }
  ).then(async () => {
    try {
      await restoreBackup(row.id)
      ElMessage.success('还原成功')
    } catch (error) {
      console.error('还原失败:', error)
    }
  }).catch(() => {})
}

const handleDeleteBackup = (row) => {
  ElMessageBox.confirm(`确定要删除备份 "${row.file_name}" 吗？`, '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning',
  }).then(async () => {
    try {
      await deleteBackup(row.id)
      ElMessage.success('删除成功')
      fetchBackupList()
    } catch (error) {
      console.error('删除失败:', error)
    }
  }).catch(() => {})
}

const getBackupStatusType = (status) => {
  const map = { 0: 'warning', 1: 'success', 2: 'danger' }
  return map[status] || 'info'
}

const getBackupStatusText = (status) => {
  const map = { 0: '进行中', 1: '成功', 2: '失败' }
  return map[status] || '未知'
}

// ==================== 清除业务数据 ====================
const clearLoading = ref(false)
const hasRecentBackup = ref(false)
const dataTableList = ref([])
const selectedCategories = ref([])

const selectAll = computed({
  get: () => dataTableList.value.length > 0 && selectedCategories.value.length === dataTableList.value.length,
  set: (val) => {
    selectedCategories.value = val ? dataTableList.value.map(item => item.category) : []
  }
})

const handleSelectAll = (val) => {
  selectedCategories.value = val ? dataTableList.value.map(item => item.category) : []
}

const fetchDataTables = async () => {
  try {
    const res = await getDataTables()
    if (res.data) {
      dataTableList.value = res.data
    }
  } catch (error) {
    console.error('获取数据表列表失败:', error)
  }
}

const checkRecentBackup = async () => {
  try {
    const res = await checkRecentBackupAPI()
    if (res.data) {
      hasRecentBackup.value = res.data.has_backup
    }
  } catch (error) {
    console.error('检查备份状态失败:', error)
  }
}

const handleClearData = async () => {
  if (!hasRecentBackup.value) {
    ElMessage.warning('10分钟内无备份，请先执行数据备份')
    return
  }

  if (selectedCategories.value.length === 0) {
    ElMessage.warning('请至少选择一项要清除的数据')
    return
  }

  const selectedLabels = dataTableList.value
    .filter(item => selectedCategories.value.includes(item.category))
    .map(item => item.label)
    .join('、')

  ElMessageBox.confirm(
    `确定要清除以下数据吗？<br><strong>${selectedLabels}</strong><br><br><span style="color: #f56c6c">警告：此操作不可逆，数据将被永久删除！</span>`,
    '确认清除数据',
    {
      confirmButtonText: '确定清除',
      cancelButtonText: '取消',
      type: 'danger',
      dangerouslyUseHTMLString: true,
    }
  ).then(async () => {
    clearLoading.value = true
    try {
      await clearData({ table_categories: selectedCategories.value })
      ElMessage.success('数据清除成功')
      selectedCategories.value = []
    } catch (error) {
      console.error('清除数据失败:', error)
    } finally {
      clearLoading.value = false
    }
  }).catch(() => {})
}

// ==================== 工具函数 ====================
const formatFileSize = (bytes) => {
  if (!bytes || bytes === 0) return '0 B'
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
  fetchBackupList()
  fetchDataTables()
  checkRecentBackup()
})
</script>

<style lang="scss" scoped>
.system-maintenance {
  .maintenance-card {
    margin-bottom: 16px;

    .card-header {
      display: flex;
      justify-content: space-between;
      align-items: center;

      .title {
        font-size: 16px;
        font-weight: 600;
      }
    }
  }

  .pagination-container {
    margin-top: 16px;
    display: flex;
    justify-content: flex-end;
  }

  .table-selection {
    .selection-header {
      display: flex;
      align-items: center;
      gap: 12px;
      margin-bottom: 12px;
      padding-bottom: 12px;
      border-bottom: 1px solid #e4e7ed;

      .selection-count {
        color: #606266;
        font-size: 14px;
      }
    }

    .category-list {
      display: flex;
      flex-wrap: wrap;
      gap: 12px;

      :deep(.el-checkbox) {
        margin-right: 0;
        height: auto;
      }

      .category-card {
        width: calc(33.333% - 8px);
        min-width: 280px;
        cursor: pointer;
        transition: all 0.3s;

        &.selected {
          border-color: #409eff;
          background-color: #f0f9ff;
        }

        &:hover {
          border-color: #409eff;
        }

        :deep(.el-card__body) {
          padding: 12px 16px;
        }

        .category-content {
          .category-label {
            font-size: 15px;
            font-weight: 600;
            color: #303133;
            margin-bottom: 4px;
          }

          .category-desc {
            font-size: 12px;
            color: #909399;
            margin-bottom: 8px;
          }

          .category-tables {
            display: flex;
            flex-wrap: wrap;
            gap: 4px;

            .el-tag {
              font-size: 11px;
            }
          }
        }
      }
    }
  }
}
</style>
