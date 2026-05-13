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
  xData: { type: Array, default: () => [] },
  seriesData: { type: Array, default: () => [] },
  height: { type: String, default: '300px' },
})

const chartRef = ref(null)
let chartInstance = null

const renderChart = () => {
  if (!chartRef.value) return
  if (!chartInstance) {
    chartInstance = echarts.init(chartRef.value)
  }

  const series = props.seriesData.map((item) => ({
    name: item.name || '',
    type: 'bar',
    barWidth: item.barWidth || undefined,
    barMaxWidth: item.barMaxWidth || 40,
    data: item.data || [],
    stack: item.stack || undefined,
    itemStyle: item.itemStyle || {},
    label: item.label || undefined,
    yAxisIndex: item.yAxisIndex || 0,
  }))

  const yAxisList = []
  const hasMultiAxis = props.seriesData.some((s) => s.yAxisIndex === 1)
  yAxisList.push({
    type: 'value',
    axisLabel: { color: '#909399' },
    splitLine: { lineStyle: { type: 'dashed', color: '#E4E7ED' } },
  })
  if (hasMultiAxis) {
    yAxisList.push({
      type: 'value',
      axisLabel: { color: '#909399' },
      splitLine: { show: false },
    })
  }

  const option = {
    tooltip: {
      trigger: 'axis',
      confine: true,
    },
    legend: {
      data: props.seriesData.map((s) => s.name),
      bottom: 0,
    },
    grid: {
      left: '3%',
      right: hasMultiAxis ? '4%' : '3%',
      bottom: props.seriesData.length > 1 ? '12%' : '3%',
      top: '10px',
      containLabel: true,
    },
    xAxis: {
      type: 'category',
      data: props.xData,
      axisLabel: { color: '#909399' },
      axisLine: { lineStyle: { color: '#DCDFE6' } },
    },
    yAxis: yAxisList,
    series,
  }

  chartInstance.setOption(option, true)
}

const handleResize = () => {
  chartInstance?.resize()
}

watch(
  () => [props.xData, props.seriesData],
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
