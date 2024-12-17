<template>
  <el-container>
    <el-header>
      <h2>Layer2 监控面板</h2>
    </el-header>

    <el-main>
      <el-row :gutter="20">
        <el-col :span="8">
          <el-card class="chain-info">
            <template #header>
              <div class="card-header">
                <span>L1 链信息</span>
              </div>
            </template>
            <div class="info-item">
              <span>区块高度:</span>
              <span>{{ stats.l1_blocks || 0 }}</span>
            </div>
            <div class="info-item">
              <span>余额:</span>
              <span>{{ formatBalance(stats.l1_balance) || 0 }} ETH</span>
            </div>
          </el-card>
        </el-col>

        <el-col :span="8">
          <el-card class="chain-info">
            <template #header>
              <div class="card-header">
                <span>L2 链信息</span>
              </div>
            </template>
            <div class="info-item">
              <span>区块高度:</span>
              <span>{{ stats.l2_blocks || 0 }}</span>
            </div>
            <div class="info-item">
              <span>TPS:</span>
              <span>{{ (stats.l2_tps || 0).toFixed(2) }}</span>
            </div>
          </el-card>
        </el-col>

        <el-col :span="8">
          <el-card class="chain-info">
            <template #header>
              <div class="card-header">
                <span>性能指标</span>
              </div>
            </template>
            <div class="info-item">
              <span>当前TPS:</span>
              <span>{{ (stats.current_tps || 0).toFixed(2) }}</span>
            </div>
            <div class="info-item">
              <span>峰值TPS:</span>
              <span>{{ (stats.peak_tps || 0).toFixed(2) }}</span>
            </div>
          </el-card>
        </el-col>
      </el-row>

      <!-- TPS趋势图 -->
      <el-row :gutter="20" class="chart-row">
        <el-col :span="24">
          <el-card>
            <template #header>
              <div class="card-header">
                <span>TPS趋势</span>
              </div>
            </template>
            <div class="chart" ref="tpsChart"></div>
          </el-card>
        </el-col>
      </el-row>

      <!-- 交易列表 -->
      <el-card>
        <template #header>
          <div class="card-header">
            <span>最新交易</span>
          </div>
        </template>
        <el-table :data="transactions" style="width: 100%">
          <el-table-column prop="hash" label="交易哈希" width="220" />
          <el-table-column prop="from" label="发送方" width="180" />
          <el-table-column prop="to" label="接收方" width="180" />
          <el-table-column prop="value" label="金额" width="120" />
          <el-table-column prop="time" label="时间" width="180" />
          <el-table-column prop="status" label="状态">
            <template #default="scope">
              <el-tag :type="scope.row.status === 1 ? 'success' : 'danger'">
                {{ scope.row.status === 1 ? '成功' : '失败' }}
              </el-tag>
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </el-main>
  </el-container>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import * as echarts from 'echarts'
import { getStats, subscribeToUpdates } from '@/http/http.dashboard'
import type { StatsResp } from '@/model/dashboardModel'

const stats = ref<StatsResp>({
  current_tps: 0,
  peak_tps: 0,
  total_tx: 0,
  block_height: 0,
  active_users: 0,
  l1_blocks: 0,
  l2_blocks: 0,
  l1_balance: '',
  l2_tps: 0
})

const transactions = ref([])
const tpsChart = ref(null)

// 格式化余额显示
const formatBalance = (balance: number) => {
  if (!balance) return '0'
  return (balance / 1e18).toFixed(4)
}

// 初始化图表
const initCharts = () => {
  const tpsOption = {
    title: { text: 'TPS趋势' },
    tooltip: { trigger: 'axis' },
    xAxis: { type: 'time' },
    yAxis: { type: 'value' },
    series: [{
      name: 'TPS',
      type: 'line',
      smooth: true,
      data: []
    }]
  }

  if (tpsChart.value) {
    const chart = echarts.init(tpsChart.value)
    chart.setOption(tpsOption)
  }
}

// 获取初始数据
const fetchInitialData = async () => {
  const data = await getStats()
  if (data) {
    stats.value = data
  }
}

// 订阅实时更新
const subscribeUpdates = () => {
  subscribeToUpdates((data) => {
    if (data) {
      stats.value = data
      updateChart(data.current_tps)
    }
  })
}

// 更新图表数据
const updateChart = (tps: number) => {
  if (tpsChart.value) {
    const chart = echarts.getInstanceByDom(tpsChart.value)
    if (chart) {
      const now = new Date()
      const data = chart.getOption().series[0].data
      data.push([now, tps])
      if (data.length > 50) data.shift()
      chart.setOption({
        series: [{
          data: data
        }]
      })
    }
  }
}

onMounted(() => {
  initCharts()
  fetchInitialData()
  subscribeUpdates()
})
</script>

<style scoped>
.chain-info {
  margin-bottom: 20px;
}

.info-item {
  display: flex;
  justify-content: space-between;
  margin: 10px 0;
}

.chart {
  height: 400px;
}

.chart-row {
  margin: 20px 0;
}
</style>
