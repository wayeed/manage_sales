<template>
  <el-dialog
    v-model="visible"
    :title="isEdit ? '编辑SKU' : '新增SKU'"
    width="600px"
    :close-on-click-modal="false"
    @close="handleClose"
  >
    <el-form
      ref="formRef"
      :model="formData"
      :rules="formRules"
      label-width="100px"
    >
      <el-form-item label="SKU编码" prop="sku_code">
        <el-input
          v-model="formData.sku_code"
          placeholder="请输入SKU编码"
          :disabled="isEdit"
        />
      </el-form-item>
      <el-form-item label="SKU名称" prop="sku_name">
        <el-input v-model="formData.sku_name" placeholder="请输入SKU名称" />
      </el-form-item>
      <el-form-item label="条码">
        <el-input v-model="formData.barcode" placeholder="请输入条码" />
      </el-form-item>
      <el-form-item label="规格属性">
        <div class="attr-editor">
          <div v-for="(item, index) in attrList" :key="index" class="attr-row">
            <el-input v-model="item.key" placeholder="属性名" style="width: 120px" />
            <span style="margin: 0 8px">:</span>
            <el-input v-model="item.value" placeholder="属性值" style="flex: 1" />
            <el-button type="danger" link @click="removeAttr(index)" style="margin-left: 8px">
              删除
            </el-button>
          </div>
          <el-button type="primary" link @click="addAttr">
            <el-icon><Plus /></el-icon>
            添加属性
          </el-button>
        </div>
      </el-form-item>
      <el-form-item label="状态">
        <el-radio-group v-model="formData.status">
          <el-radio :value="1">启用</el-radio>
          <el-radio :value="0">禁用</el-radio>
        </el-radio-group>
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="handleClose">取消</el-button>
      <el-button type="primary" @click="handleSubmit" :loading="loading">确定</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, reactive, watch, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false,
  },
  skuData: {
    type: Object,
    default: null,
  },
  productId: {
    type: Number,
    default: 0,
  },
})

const emit = defineEmits(['update:modelValue', 'success'])

const visible = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val),
})

const formRef = ref(null)
const loading = ref(false)
const isEdit = computed(() => !!props.skuData?.id)

const formData = reactive({
  sku_code: '',
  sku_name: '',
  barcode: '',
  status: 1,
})

const attrList = ref([])

const formRules = {
  sku_code: [{ required: true, message: '请输入SKU编码', trigger: 'blur' }],
  sku_name: [{ required: true, message: '请输入SKU名称', trigger: 'blur' }],
}

// 监听弹窗打开，初始化表单
watch(visible, (val) => {
  if (val) {
    if (props.skuData) {
      // 编辑模式
      formData.sku_code = props.skuData.sku_code || ''
      formData.sku_name = props.skuData.sku_name || ''
      formData.barcode = props.skuData.barcode || ''
      formData.status = props.skuData.status ?? 1
      // 解析属性
      if (props.skuData.attributes) {
        try {
          const attrs = typeof props.skuData.attributes === 'string'
            ? JSON.parse(props.skuData.attributes)
            : props.skuData.attributes
          attrList.value = Object.entries(attrs).map(([key, value]) => ({
            key,
            value: String(value),
          }))
        } catch {
          attrList.value = []
        }
      } else {
        attrList.value = []
      }
    } else {
      // 新增模式
      resetForm()
    }
  }
})

const resetForm = () => {
  formData.sku_code = ''
  formData.sku_name = ''
  formData.barcode = ''
  formData.status = 1
  attrList.value = []
}

const addAttr = () => {
  attrList.value.push({ key: '', value: '' })
}

const removeAttr = (index) => {
  attrList.value.splice(index, 1)
}

const getAttributes = () => {
  const attrs = {}
  attrList.value.forEach((item) => {
    if (item.key && item.value) {
      attrs[item.key] = item.value
    }
  })
  return Object.keys(attrs).length > 0 ? JSON.stringify(attrs) : ''
}

const handleClose = () => {
  visible.value = false
  formRef.value?.resetFields()
}

const handleSubmit = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  loading.value = true
  try {
    const data = {
      ...formData,
      attributes: getAttributes(),
    }
    emit('success', data, isEdit.value)
    handleClose()
  } finally {
    loading.value = false
  }
}
</script>

<style lang="scss" scoped>
.attr-editor {
  width: 100%;

  .attr-row {
    display: flex;
    align-items: center;
    margin-bottom: 8px;
  }
}
</style>
