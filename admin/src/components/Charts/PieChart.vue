<template>
  <div class="chart-wrapper">
    <div v-if="title" class="chart-title">{{ title }}</div>
    <div ref="chartRef" :style="{ height: height || '300px', width: '100%' }"></div>
  </div>
</template>

<script setup>
import { ref, watch, onMounted, onBeforeUnmount, nextTick } from 'vue'
import * as echarts from 'echarts'

const props = defineProps({
  title: { type: String, default: '' },
  seriesData: { type: Array, default: () => [] },
  height: { type: String, default: '300px' },
})

const chartRef = ref(null)
let chartInstance = null

const defaultColors = [
  '#1890ff', '#52c41a', '#faad14', '#f5222d', '#722ed1',
  '#13c2c2', '#eb2f96', '#fa8c16', '#a0d911', '#2f54eb',
]

const renderChart = () => {
  if (!chartRef.value) return
  if (!chartInstance) {
    chartInstance = echarts.init(chartRef.value)
  }

  const data = props.seriesData.map((item) => ({
    name: item.name,
    value: item.value,
    itemStyle: item.itemStyle || {},
  }))

  const option = {
    tooltip: {
      trigger: 'item',
      confine: true,
      formatter: '{b}: {c} ({d}%)',
    },
    legend: {
      orient: 'vertical',
      right: '5%',
      top: 'center',
      itemWidth: 12,
      itemHeight: 12,
      textStyle: { color: '#606266', fontSize: 13 },
    },
    color: defaultColors,
    series: [
      {
        type: 'pie',
        radius: props.seriesData.length > 5 ? ['35%', '65%'] : ['0%', '65%'],
        center: ['35%', '50%'],
        avoidLabelOverlap: true,
        padAngle: 2,
        itemStyle: {
          borderRadius: props.seriesData.length > 5 ? 4 : 0,
        },
        label: {
          show: props.seriesData.length <= 5,
          formatter: '{b}\n{d}%',
          fontSize: 12,
          color: '#606266',
        },
        emphasis: {
          label: { show: true, fontSize: 14, fontWeight: 'bold' },
          itemStyle: { shadowBlur: 10, shadowOffsetX: 0, shadowColor: 'rgba(0,0,0,0.2)' },
        },
        data,
      },
    ],
  }

  chartInstance.setOption(option, true)
}

const handleResize = () => {
  chartInstance?.resize()
}

watch(
  () => props.seriesData,
  () => {
    nextTick(renderChart)
  },
  { deep: true }
)

onMounted(() => {
  nextTick(renderChart)
  window.addEventListener('resize', handleResize)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', handleResize)
  chartInstance?.dispose()
  chartInstance = null
})
</script>

<style lang="scss" scoped>
.chart-wrapper {
  width: 100%;

  .chart-title {
    font-size: 16px;
    font-weight: 600;
    color: #303133;
    margin-bottom: 12px;
  }
}
</style>
