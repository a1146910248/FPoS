<template>
  <el-container>
    <el-header>
      <h2 class="dashboard-title">Layer2 监控面板</h2>
    </el-header>

    <el-main>
      <!-- 状态卡片行 -->
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

      <!-- 趋势图和交易列表行 -->
      <el-row :gutter="20" class="data-row">
        <!-- TPS趋势图 -->
        <el-col :span="12">
          <el-card class="chart-card">
            <template #header>
              <div class="card-header">
                <span>TPS趋势</span>
              </div>
            </template>
            <div class="chart" ref="tpsChart"></div>
          </el-card>
        </el-col>

        <!-- 交易列表 -->
        <el-col :span="12">
          <el-card class="transaction-card">
            <template #header>
              <div class="card-header">
                <span>最新交易</span>
              </div>
            </template>
            <el-table
              :data="transactionList"
              style="width: 100%"
              v-loading="loading"
              :height="tableHeight"
            >
              <!-- 交易哈希 -->
              <el-table-column
                prop="hash"
                label="交易哈希"
                width="180"
                fixed="left"
              >
                <template #default="{ row }">
                  <div class="hash-container">
                    <el-tooltip :content="row.hash" placement="top" effect="light">
                      <span class="hash-text">{{ formatHash(row.hash) }}</span>
                    </el-tooltip>
                    <el-button
                      class="copy-btn"
                      type="primary"
                      link
                      @click="copyToClipboard(row.hash)"
                    >
                      <el-icon><CopyDocument /></el-icon>
                    </el-button>
                  </div>
                </template>
              </el-table-column>

              <!-- 方法 -->
              <el-table-column prop="method" label="方法" width="120">
                <template #default="{ row }">
                  <el-tag
                    size="small"
                    :type="getMethodTagType(row.method)"
                    v-if="getMethodTagType(row.method)"
                  >
                  {{ row.method || 'Transfer' }}
                  </el-tag>
                  <el-tag
                    size="small"
                    v-else
                  >
                    {{ row.method || 'Transfer' }}
                  </el-tag>
                </template>
              </el-table-column>

              <!-- 区块 -->
              <el-table-column prop="block" label="区块" width="100">
                <template #default="{ row }">
                  <el-link type="primary">{{ row.block }}</el-link>
                </template>
              </el-table-column>

              <!-- 时间 -->
              <el-table-column prop="timestamp" label="时间" width="150">
                <template #default="{ row }">
                  <el-tooltip :content="formatFullTime(row.timestamp)" placement="top">
                    <span>{{ formatRelativeTime(row.timestamp) }}</span>
                  </el-tooltip>
                </template>
              </el-table-column>

              <!-- 发送方 -->
              <el-table-column prop="from" label="发送方" min-width="180">
                <template #default="{ row }">
                  <div class="address-container">
                    <el-tooltip :content="row.from" placement="top" effect="light">
                      <span class="address-text">{{ formatAddress(row.from) }}</span>
                    </el-tooltip>
                    <el-button
                      class="copy-btn"
                      type="primary"
                      link
                      @click="copyToClipboard(row.from)"
                    >
                      <el-icon><CopyDocument /></el-icon>
                    </el-button>
                  </div>
                </template>
              </el-table-column>

              <!-- 箭头 -->
              <el-table-column width="50">
                <template #default>
                  <el-icon color="#67C23A"><ArrowRight /></el-icon>
                </template>
              </el-table-column>

              <!-- 接收方 -->
              <el-table-column prop="to" label="接收方" min-width="180">
                <template #default="{ row }">
                  <div class="address-container">
                    <el-tooltip :content="row.to" placement="top" effect="light">
                      <span class="address-text">{{ formatAddress(row.to) }}</span>
                    </el-tooltip>
                    <el-button
                      class="copy-btn"
                      type="primary"
                      link
                      @click="copyToClipboard(row.to)"
                    >
                      <el-icon><CopyDocument /></el-icon>
                    </el-button>
                  </div>
                </template>
              </el-table-column>

              <!-- 金额 -->
              <el-table-column
                prop="value"
                label="金额"
                width="150"
                align="right"
              >
                <template #default="{ row }">
                  <span>{{ formatValue(row.value) }} Wei</span>
                </template>
              </el-table-column>

              <!-- Gas费用 -->
              <el-table-column
                prop="gas_price"
                label="Gas费用"
                width="120"
                align="right"
              >
                <template #default="{ row }">
                  <span>{{ formatGasPrice(row.gas_price) }}</span>
                </template>
              </el-table-column>

              <el-table-column
                prop="status"
                label="状态"
                width="100"
                fixed="right"
              >
                <template #default="{ row }">
                  <el-tag :type="row.status === 1 ? 'success' : 'danger'">
                    {{ row.status === 1 ? '成功' : '失败' }}
                  </el-tag>
                </template>
              </el-table-column>
            </el-table>

            <div class="pagination-container">
              <el-pagination
                v-model:current-page="currentPage"
                v-model:page-size="pageSize"
                :total="total"
                :page-sizes="[10, 15, 20]"
                @current-change="fetchTransactions"
                @size-change="handleSizeChange"
                layout="total, sizes, prev, pager, next"
              />
            </div>
          </el-card>
        </el-col>
      </el-row>
    </el-main>
  </el-container>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import * as echarts from 'echarts'
import { getStats, getTransactions, subscribeToUpdates } from '@/http/http.dashboard'
import type { StatsResp, Transaction } from '@/model/dashboardModel'
import relativeTime from 'dayjs/plugin/relativeTime'
import dayjs from 'dayjs'
import { ElMessage } from 'element-plus'
import { formatEther } from 'ethers'

// 基础数据
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

// 图表相关
const tpsChart = ref(null)

// 交易列表相关
const loading = ref(false)
const currentPage = ref(1)
const transactionList = ref<Transaction[]>([])
const total = ref(0)

// 设置表格高度
const tableHeight = 350 // 根据实际内容调整
const pageSize = ref(15) // 默认显示15条

// 处理页面大小变化
const handleSizeChange = (val: number) => {
  pageSize.value = val
  fetchTransactions()
}

const formatBalance = (balance: string | number) => {
  if (!balance) return '0'
  try {
    const value = Number(balance)
    if (isNaN(value)) return '0'
    return (value / 1e18).toFixed(4)
  } catch (e) {
    console.error('Invalid balance format:', e)
    return '0'
  }
}


dayjs.extend(relativeTime)
dayjs.locale('zh-cn')

// 复制到剪贴板
const copyToClipboard = (text: string) => {
  navigator.clipboard.writeText(text)
  ElMessage.success('已复制到剪贴板')
}

// 获取方法标签类型
const getMethodTagType = (method: string): string | undefined => {
  const types: { [key: string]: string } = {
    'Transfer': 'primary',  // 设置默认类型为 primary
    'Approve': 'success',
    'Swap': 'warning',
    'Mint': 'danger',
    'Burn': 'info'
  }
  return types[method]  // 如果没找到对应的type，返回 undefined
}


// 格式化相对时间
const formatRelativeTime = (timestamp: string) => {
  return dayjs(timestamp).fromNow()
}

// 格式化完整时间
const formatFullTime = (timestamp: string) => {
  return dayjs(timestamp).format('YYYY-MM-DD HH:mm:ss')
}

// 图表实例引用
let chartInstance: echarts.ECharts | null = null

// 初始化图表
const initCharts = () => {
  if (tpsChart.value) {
    // 保存图表实例
    chartInstance = echarts.init(tpsChart.value)
    chartInstance.setOption({
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
    })
  }
}

// 格式化函数
const formatHash = (hash: string) => {
  if (!hash) return ''
  return `${hash.slice(0, 6)}...${hash.slice(-4)}`
}

const formatAddress = (address: string) => {
  if (!address) return ''
  return `${address.slice(0, 6)}...${address.slice(-4)}`
}

const formatValue = (value: number) => {
  // return (value / 1e18).toFixed(6)
  return value
}

const formatGasPrice = (gasPrice: number) => {
  return (gasPrice / 1e9).toFixed(2)
}

// 更新图表
const updateChart = (tps: number) => {
  if (chartInstance) {
    const option = chartInstance.getOption()
    const seriesData = (option.series as any)[0].data as [number, number][]

    seriesData.push([Date.now(), tps])
    if (seriesData.length > 50) {
      seriesData.shift()
    }

    chartInstance.setOption({
      series: [{
        data: seriesData
      }]
    })
  }
}
// 处理窗口大小变化
const handleResize = () => {
  if (chartInstance) {
    chartInstance.resize()
  }
}

// 数据获取函数
const fetchInitialData = async () => {
  const data = await getStats()
  if (data) stats.value = data
}

const fetchTransactions = async () => {
  loading.value = true
  try {
    const data = await getTransactions(currentPage.value, pageSize.value)
    if (data) {
      transactionList.value = data.list || []
      total.value = data.total
    }
  } finally {
    loading.value = false
  }
}

// WebSocket订阅
const subscribeUpdates = () => {
  subscribeToUpdates((data) => {
    if (data) {
      stats.value = data
      updateChart(data.current_tps)
    }
  })
}

// 生命周期钩子
onMounted(() => {
  initCharts()
  fetchInitialData()
  subscribeUpdates()
  fetchTransactions()

  // 添加窗口大小变化监听
  window.addEventListener('resize', handleResize)
})

// 清理监听器
onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
  // 销毁图表实例
  if (chartInstance) {
    chartInstance.dispose()
  }
})
</script>

<style scoped>
/* Header 样式 */
.el-header {
  padding: 0 20px;
  background-color: white;
  height: auto !important; /* 覆盖 element-plus 默认高度 */
  line-height: normal;
  border-bottom: 1px solid #e6e6e6;
}

.dashboard-title {
  margin: 0;
  padding: 16px 0;
  font-size: 20px;
  color: #303133;
  font-weight: 600;
}

.el-main {
  padding: 20px;
  background-color: #f0f2f5;
  height: calc(100vh - 60px); /* 减去 header 高度 */
  overflow-y: auto;
}

/* 确保整个容器占满视口 */
.el-container {
  height: 100vh;
}

.chain-info {
  margin-bottom: 20px;
  height: 100%;
}

.data-row {
  margin-top: 20px;
}

.chart-card,
.transaction-card {
  height: 100%;
}

.chart {
  height: 400px;
  width: 100%;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.transaction-card :deep(.el-card__body) {
  padding: 0;
}

.transaction-card :deep(.el-table) {
  border-radius: 4px 4px 0 0;
}

.pagination-container {
  padding: 15px 0;
  display: flex;
  justify-content: center;
  background-color: white;
  border-radius: 0 0 4px 4px;
}

/* 表格样式优化 */
:deep(.el-table th) {
  background-color: #f5f7fa;
  color: #606266;
  font-weight: bold;
}

.hash-text,
.address-text {
  color: #409EFF;
  cursor: pointer;
}

.hash-text:hover,
.address-text:hover {
  text-decoration: underline;
}

/* 确保卡片内容区域充满高度 */
.el-card {
  height: 100%;
  display: flex;
  flex-direction: column;
}

:deep(.el-card__body) {
  flex: 1;
  overflow: hidden;
}
</style>
