<template>
  <el-container>
    <el-header>
      <div class="header-container">
        <div class="header-left">
          <img src="@/assets/logo.png" alt="Logo" class="logo" />
          <h2 class="dashboard-title">Layer2 监控面板</h2>
        </div>
        <div class="header-right">
          <el-space :size="16" alignment="center">
            <el-tag
              :type="wsConnected ? 'success' : 'danger'"
              size="large"
              :effect="isDark ? 'dark' : 'light'"
            >
              <div class="network-status-wrapper">
                <el-icon><Connection /></el-icon>
                <span>{{ wsConnected ? '网络已连接' : '网络未连接' }}</span>
              </div>
            </el-tag>

            <el-button
              :type="isDark ? 'primary' : 'default'"
              @click="toggleDark"
            >
              <el-icon>
                <Moon v-if="!isDark" />
                <Sunny v-else />
              </el-icon>
              {{ isDark ? '浅色模式' : '深色模式' }}
            </el-button>

            <el-button type="primary">
              <el-icon><Setting /></el-icon>
              设置
            </el-button>
          </el-space>
        </div>
      </div>
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
                <el-button
                  :loading="loading"
                  type="primary"
                  link
                  @click="refreshTransactions"
                >
                  <el-icon><Refresh /></el-icon>
                  刷新
                </el-button>
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
              <el-table-column prop="block" label="区块" width="160">
                <template #default="{ row }">
                  <div class="hash-container">
                    <el-tooltip :content="row.hash" placement="top" effect="light">
                      <span class="hash-text">{{ formatHash(row.block_hash) }}</span>
                    </el-tooltip>
                    <el-button
                      class="copy-btn"
                      type="primary"
                      link
                      @click="copyToClipboard(row.block_hash)"
                    >
                      <el-icon v-if="row.block_hash"><CopyDocument /></el-icon>
                    </el-button>
                  </div>
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
                width="120"
                fixed="right"
              >
                <template #default="{ row }">
                  <el-tooltip
                    :content="row.block_hash ? `区块: ${formatHash(row.block_hash)}` : ''"
                    placement="top"
                    v-if="row.status === TransactionStatus.Confirmed || row.status === TransactionStatus.L1Confirmed"
                  >
                    <el-tag :type="getStatusTagType(row.status)">
                      {{ getStatusText(row.status) }}
                    </el-tag>
                  </el-tooltip>
                  <el-tag v-else :type="getStatusTagType(row.status)">
                    {{ getStatusText(row.status) }}
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
import { TransactionStatus } from '@/model/dashboardModel'
import relativeTime from 'dayjs/plugin/relativeTime'
import { Refresh, Connection, Setting, Moon, Sunny } from '@element-plus/icons-vue'
import dayjs from 'dayjs'
import { ElMessage } from 'element-plus'
import { useDark, useToggle } from '@vueuse/core'

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

// 深色模式
const isDark = useDark()
const toggleDark = useToggle(isDark)

// WebSocket 连接状态
const wsConnected = ref(false)
let wsInstance: WebSocket | null = null

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

// 获取状态标签类型
const getStatusTagType = (status: TransactionStatus): string => {
  const types: { [key in TransactionStatus]: string } = {
    [TransactionStatus.Pending]: 'info',
    [TransactionStatus.Confirmed]: 'success',
    [TransactionStatus.L1Submitting]: 'warning',
    [TransactionStatus.L1Confirmed]: 'success',
    [TransactionStatus.L1Failed]: 'danger'
  }
  return types[status]
}

// 获取状态文本
const getStatusText = (status: TransactionStatus): string => {
  const texts: { [key in TransactionStatus]: string } = {
    [TransactionStatus.Pending]: '待确认',
    [TransactionStatus.Confirmed]: 'L2已确认',
    [TransactionStatus.L1Submitting]: 'L1提交中',
    [TransactionStatus.L1Confirmed]: 'L1已确认',
    [TransactionStatus.L1Failed]: 'L1失败'
  }
  return texts[status]
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

// 刷新交易列表
const refreshTransactions = async () => {
  if (loading.value) return

  loading.value = true
  try {
    const data = await getTransactions(currentPage.value, pageSize.value)
    if (data) {
      transactionList.value = data.list
      total.value = data.total
      ElMessage.success('刷新成功')
    }
  } catch (error) {
    console.error('刷新失败:', error)
    ElMessage.error('刷新失败，请重试')
  } finally {
    loading.value = false
  }
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

// // WebSocket订阅
// const subscribeUpdates = () => {
//   subscribeToUpdates((data) => {
//     if (data) {
//       wsConnected.value = true
//       stats.value = data
//       updateChart(data.current_tps)
//     } else {
//       wsConnected.value = false
//     }
//   })
// }
// 初始化 WebSocket 并处理连接状态
const subscribeUpdates = () => {
  wsInstance = subscribeToUpdates((data) => {
    // 处理接收到的数据
    if (data) {
      wsConnected.value = true
      stats.value = data
      updateChart(data.current_tps)
    }
  })

  // 添加连接状态监听
  if (wsInstance) {
    wsInstance.onopen = () => {
      console.log('WebSocket连接已建立')
      wsConnected.value = true
    }

    wsInstance.onclose = () => {
      console.log('WebSocket连接已关闭')
      wsConnected.value = false
    }

    wsInstance.onerror = () => {
      console.error('WebSocket连接错误')
      wsConnected.value = false
    }
  }
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
  padding: 0;
  background-color: white;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.08);
  position: relative;
  z-index: 100;
}

.header-container {
  height: 60px;
  padding: 0 24px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.logo {
  height: 32px;
  width: auto;
}

.dashboard-title {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: var(--el-text-color-primary);
  background: linear-gradient(45deg, #409EFF, #67C23A);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}

.header-right {
  display: flex;
  align-items: center;
}

.network-status-wrapper {
  display: flex;
  align-items: center;
  gap: 6px;
  height: 100%;
}

.el-tag {
  display: flex;
  align-items: center;
  padding: 0 12px;
  height: 32px;
  border-radius: 6px;
}

/* 按钮样式优化 */
.el-button {
  border-radius: 6px;
  transition: all 0.3s;
}

.el-button:hover {
  transform: translateY(-1px);
}

.el-button .el-icon {
  margin-right: 4px;
  font-size: 16px;
}

/* 深色模式按钮样式 */
.el-button .el-icon {
  margin-right: 4px;
  font-size: 16px;
  vertical-align: middle;
}

.el-tag .el-icon {
  margin-right: 4px;
  font-size: 16px;
}

.dashboard-title {
  margin: 0;
  padding: 16px 0;
  font-size: 20px;
  color: #303133;
  font-weight: 600;
}

.el-main {
  padding: 24px;
  background-color: #f0f2f5;
  height: calc(100vh - 60px);
  overflow-y: auto;
}

/* 确保整个容器占满视口 */
.el-container {
  height: 100vh;
}

/* 暗色模式支持 */
:root {
  --header-bg-color: #ffffff;
  --header-shadow: 0 2px 4px rgba(0, 0, 0, 0.08);
}

html.dark {
  --header-bg-color: var(--el-bg-color);
  --header-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
}

.el-header {
  background-color: var(--header-bg-color);
  box-shadow: var(--header-shadow);
}

/* 响应式调整 */
@media screen and (max-width: 768px) {
  .header-right .el-button span,
  .network-status-wrapper span {
    display: none;
  }

  .network-status-wrapper {
    gap: 0;
  }

  .el-tag {
    padding: 0 8px;
  }
}

/* 暗色模式支持 */
@media (prefers-color-scheme: dark) {
  .el-header {
    background-color: var(--el-bg-color);
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
  }

  .dashboard-title {
    background: linear-gradient(45deg, #79bbff, #95d475);
    -webkit-background-clip: text;
  }
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

/* 刷新按钮的hover效果 */
.el-button.el-button--primary.is-link:hover {
  opacity: 0.8;
}
/* 刷新图标的旋转动画 */
.el-icon {
  margin-right: 4px;
  transition: transform 0.3s ease;
}
.el-button:not(.is-loading):hover .el-icon {
  transform: rotate(180deg);
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
