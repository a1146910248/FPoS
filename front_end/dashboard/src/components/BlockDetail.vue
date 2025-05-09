<template>
  <div class="transaction-container">
    <el-card class="glassmorphic-card">
      <template #header>
        <div class="card-header">
          <h3 class="gradient-title">区块详情</h3>
          <el-button
            type="primary"
            @click="navigateToDashboard"
            class="floating-btn"
          >
            <el-icon><ArrowLeft /></el-icon>
            返回仪表板
          </el-button>
        </div>
      </template>
      <transition name="el-zoom-in-top">
        <el-card v-if="blockDetail" class="receipt-card">
          <el-descriptions :column="1" border>
            <!-- 基础信息 -->
            <el-descriptions-item label="区块哈希">
              <div class="hash-container">
                <span class="hash-value">{{ blockDetail.hash }}</span>
                <el-tooltip content="复制完整哈希">
                  <el-button
                    type="primary"
                    link
                    @click="copyToClipboard(blockDetail.hash)"
                  >
                    <el-icon><CopyDocument /></el-icon>
                  </el-button>
                </el-tooltip>
              </div>
            </el-descriptions-item>

            <el-descriptions-item label="上一个区块哈希">
              <div class="hash-container">
                <span class="hash-value">{{ blockDetail.previousHash }}</span>
                <el-tooltip content="复制完整哈希">
                  <el-button
                    type="primary"
                    link
                    @click="copyToClipboard(blockDetail.previousHash)"
                  >
                    <el-icon><CopyDocument /></el-icon>
                  </el-button>
                </el-tooltip>
              </div>
            </el-descriptions-item>

            <el-descriptions-item label="高度">
              {{ blockDetail.height }}
            </el-descriptions-item>

            <el-descriptions-item label="时间">
              {{ formatTime(blockDetail.timestamp) }}
            </el-descriptions-item>

            <el-descriptions-item label="状态根">
              {{ blockDetail.stateRoot }}
            </el-descriptions-item>

            <el-descriptions-item label="交易根">
              {{ blockDetail.txRoot }}
            </el-descriptions-item>

            <el-descriptions-item label="提出人">
              {{ blockDetail.proposer }}
            </el-descriptions-item>

            <el-descriptions-item label="签名">
              {{ blockDetail.signature }}
            </el-descriptions-item>

            <el-descriptions-item label="KZG承诺">
              {{ blockDetail.kzg_commitment }}
            </el-descriptions-item>

            <el-descriptions-item label="transfer交易证明">
              {{ formatHash(blockDetail.final_proof) }}
            </el-descriptions-item>

            <el-descriptions-item label="是否可疑">
              <el-tag :type="blockDetail.is_sus ? 1 : 2">
                {{ getStatusText(blockDetail.is_sus) }}
              </el-tag>
            </el-descriptions-item>

            <!-- Gas 信息 -->
            <el-descriptions-item label="Gas 消耗">
              <div class="gas-detail">
                <div>限额：{{ blockDetail.gasLimit }} units</div>
                <div>实际使用：{{ blockDetail.gasUsed }} units</div>
                <div>单价：{{ 2100 }} Gwei</div>
                <div class="total-fee">
                  总费用：{{ calculateTotalFee(blockDetail) }} Gwei
                </div>
              </div>
            </el-descriptions-item>

            <el-descriptions-item label="投票信息1" v-for="(vote, index) in blockDetail.votes" :key="index">
              <div class="gas-detail">
                <div>投票人：{{ vote.voter_address }}</div>
                <div>区块哈希：{{ vote.block_hash }} </div>
                <div>区块高度：{{ vote.block_height }} </div>
                <div class="total-fee">同意：{{ vote.approve }} </div>
                <div>时间：{{ formatTime(vote.timestamp) }}</div>
                <div>签名：{{ vote.signature }} </div>
              </div>
            </el-descriptions-item>

            <el-descriptions-item label="包含交易数">
              <el-tag :type="1">
                {{ blockDetail.transactions.length }}
              </el-tag>
            </el-descriptions-item>

              <el-descriptions-item label="交易信息">
                <div class="gas-detail" v-for="(tx, index) in visibleTransactions" :key="index">
                  <div class="address-text" @click="navigateToTransaction(tx.hash)">hash：{{ tx.hash }}</div>
                </div>
                <div class="pagination-control" v-if="hasMore">
                  <el-button type="primary" @click="showMore">显示更多</el-button>
                </div>
              </el-descriptions-item>

          </el-descriptions>
        </el-card>
      </transition>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import dayjs from 'dayjs'
import { ElMessage } from 'element-plus'
import { type Block } from '@/model/dashboardModel'
import { getBlock } from '@/http/http.dashboard.ts'

// 路由实例
const router = useRouter()

// 获取参数
const props = defineProps<{
  block_hash: number
}>()

// 复制到剪贴板
const copyToClipboard = (text: string) => {
  navigator.clipboard.writeText(text)
  ElMessage.success('已复制到剪贴板')
}

// 导航到仪表板
const navigateToDashboard = () => {
  router.push({ name: 'home' }) // 假设Dashboard的路由名称为'Dashboard'
}

const visibleCount = ref(3) // 初始显示数量

// 是否还有更多数据
const hasMore = computed(() => visibleCount.value < blockDetail.value!.transactions.length)
const visibleTransactions = computed(() => {
  return blockDetail.value!.transactions.slice(0, visibleCount.value)
})
// 显示更多条目
const showMore = () => {
  if (hasMore.value) {
    visibleCount.value += 3
  }
}

// 获取状态文本
const getStatusText = (status: boolean): string => {
  if (status) {
   return '可疑'
  } else {
    return '非可疑'
  }
}

// 接收交易详情数据
const blockDetail = ref<Block | null>(null)

const fetch = async () => {
  await getBlock(props.block_hash).then((res) => {
    if (res != null) {
      blockDetail.value = res
    }
  })
}

// 格式化函数
const formatHash = (hash: string) => {
  if (!hash) return ''
  return `${hash.slice(0, 20)}.............${hash.slice(-20)}`
}
// 单位转换
// const weiToEth = (wei: number) => (wei / 1e18).toFixed(6)
// const weiToGwei = (wei: number) => (wei / 1e9).toFixed(2)

// 导航到交易详情页面
const navigateToTransaction = (hash: string) => {
  router.push({ name: 'transactionDetail', params: { hash } })
}

const formatTime = (timestamp: string) => {
  return dayjs(timestamp).format('YYYY-MM-DD HH:mm:ss')
}

const calculateTotalFee = (b: Block) => {
  return (b.gasUsed) * 2100
}



onMounted(async () => {
  await fetch()
})
</script>


<style scoped lang="scss">
.transaction-container {
  max-width: 1200px;
  margin: 2rem auto;
  padding: 0 1rem;
}

.glassmorphic-card {
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
  border-radius: 20px;
  box-shadow: 0 8px 32px rgba(31, 38, 135, 0.15);
  border: 1px solid rgba(255, 255, 255, 0.18);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 24px;
  height: 30px;  // 降低高度
}

.cont {
  display: flex;
  justify-content: flex-start;
  align-items: center;
  width: 400px;
  gap: 16px;
  .left {
    flex: 3; /* 占3份 */
  }

  .right {
    flex: 1; /* 占1份 */
  }
}

.address-text {
  color: #409EFF;
  cursor: pointer;
}

.gradient-title {
  background: linear-gradient(45deg, #409EFF, #67C23A);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  font-size: 1.8rem;
  font-weight: 600;
}

.form-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 2rem;
  padding: 1.5rem;
}

.form-column {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.animated-input {
  transition: all 0.3s ease;

  &:focus-within {
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(64, 158, 255, 0.2);
  }
}

.send-btn {
  width: 100%;
  padding: 1rem;
  font-size: 1.1rem;
  transition: all 0.3s ease;

  &:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 15px rgba(64, 158, 255, 0.3);
  }
}

.status-alert {
  margin-top: 1.5rem;
  border-radius: 12px;
}

.floating-btn {
  animation: float 3s ease-in-out infinite;
}

@keyframes float {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-5px); }
}

.receipt-card {
  margin-top: 24px;
  border-radius: 12px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
}

.receipt-header {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 20px;
}

.hash-container {
  display: flex;
  align-items: center;
  gap: 8px;
}

.gas-detail {
  display: grid;
  gap: 4px;
}

.total-fee {
  font-weight: 500;
  color: #e6a23c;
}

.func-signature {
  font-family: monospace;
  color: #409eff;
}

:deep(.el-descriptions__body) {
  background-color: #f8f9fa;
}

:deep(.el-descriptions__label) {
  font-weight: 500;
  color: #909399;
}
</style>
