<!-- front_end/dashboard/src/components/SendTransaction.vue -->
<template>
  <div class="transaction-container">
    <el-card class="glassmorphic-card">
      <template #header>
        <div class="card-header">
          <h3 class="gradient-title">发送交易</h3>
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

      <el-form
        :model="transaction"
        ref="formRef"
        label-width="140px"
        label-position="top"

      >
        <div class="form-grid">
          <!-- 左侧列 -->
          <div class="form-column">
            <el-form-item label="发送方地址" prop="from">
              <el-input
                v-model="transaction.from"
                placeholder="0x..."
                class="address-input"
                disabled
              >
                <template #append>
                  <el-tooltip content="复制地址">
                    <el-button
                      @click="copyToClipboard(transaction.from)"
                      class="icon-btn"
                    >
                      <el-icon><CopyDocument /></el-icon>
                    </el-button>
                  </el-tooltip>
                </template>
              </el-input>
            </el-form-item>

            <el-form-item label="合约地址" prop="contractAddress" v-if="transaction.is_contract">
              <el-input
                v-model="transaction.contract_address"
                placeholder="请输入合约地址 0x..."
                class="animated-input"
              />
            </el-form-item>
            <el-form-item label="接收方地址" prop="to" v-else>
              <el-input
                v-model="transaction.to_address"
                placeholder="请输入接收方地址 0x..."
                class="animated-input"
              />
            </el-form-item>

            <el-form-item label="交易金额 (ETH)" prop="value">
              <el-input-number
                v-model="transaction.value"
                :min="0"
                :step="0.1"
                controls-position="right"
                class="eth-input"
              />
            </el-form-item>
          </div>

          <!-- 右侧列 -->
          <div class="form-column">
            <el-form-item label="合约交互" prop="isContract">
              <el-switch
                v-model="transaction.is_contract"
                active-text="是"
                inactive-text="否"
                @change="handleContractToggle"
              />
            </el-form-item>

            <div class="cont" v-if="transaction.is_contract">
              <el-form-item
                label="函数选择"
                prop="funcString"
                class="left"
              >

                <el-select
                  v-model="transaction.func_string"
                  placeholder="选择合约函数"
                  class="func-select"
                >
                  <el-option
                    v-for="item in functionOptions"
                    :key="item.value"
                    :label="item.label"
                    :value="item.value"
                  />
                </el-select>
              </el-form-item>
              <el-form-item label="参数值" prop="input" class="right">
                <el-input-number
                  v-model="transaction.input"
                  :min="0"
                  :step="1"
                  controls-position="right"
                />
              </el-form-item>
            </div>


            <el-form-item label="Gas Limit" prop="gasLimit">
              <el-input-number
                v-model="transaction.gas_limit"
                :min="21000"
                :step="1000"
                controls-position="right"
              />
            </el-form-item>
          </div>
        </div>

        <div class="action-bar">
          <el-button
            type="primary"
            @click="submit"
            class="send-btn"
          >
            <el-icon class="send-icon"><Promotion /></el-icon>
            确认发送交易
          </el-button>
        </div>
      </el-form>

      <transition name="el-zoom-in-top">
        <el-alert
          v-if="message"
          :title="message"
          :type="messageType"
          show-icon
          class="status-alert"
        />
      </transition>
    </el-card>
  </div>


  <transition name="el-zoom-in-top">
    <el-card v-if="txDetail" class="receipt-card">
      <el-descriptions :column="1" border>
        <!-- 基础信息 -->
        <el-descriptions-item label="交易哈希">
          <div class="hash-container">
            <span class="hash-value">{{ txDetail.hash }}</span>
            <el-tooltip content="复制完整哈希">
              <el-button
                type="primary"
                link
                @click="copyToClipboard(txDetail.hash)"
              >
                <el-icon><CopyDocument /></el-icon>
              </el-button>
            </el-tooltip>
          </div>
        </el-descriptions-item>

        <el-descriptions-item label="Nonce值">
          {{ txDetail.nonce }}
        </el-descriptions-item>

        <el-descriptions-item label="时间">
          {{ formatTime(txDetail.timestamp) }}
        </el-descriptions-item>

        <!-- 转账信息 -->
        <el-descriptions-item label="发送方">
          {{ txDetail.from }}
        </el-descriptions-item>

        <el-descriptions-item label="接收方">
          {{ txDetail.to }}
        </el-descriptions-item>

        <el-descriptions-item label="状态">
          <el-tag :type="getStatusText(txDetail.stat_log.status)">
            {{ getStatusText(txDetail.stat_log.status) }}
          </el-tag>
        </el-descriptions-item>

        <el-descriptions-item label="签名">
          {{ txDetail.signature }}
        </el-descriptions-item>

        <el-descriptions-item label="转账金额">
          {{ (txDetail.value) }} Gwei
        </el-descriptions-item>

        <!-- Gas 信息 -->
        <el-descriptions-item label="Gas 消耗">
          <div class="gas-detail">
            <div>限额：{{ txDetail.gasLimit}} units</div>
            <div>实际使用：{{ txDetail.gasUsed }} units</div>
            <div>单价：{{ txDetail.gasPrice }} Gwei</div>
            <div class="total-fee">
              总费用：{{ calculateTotalFee(txDetail) }} Gwei
            </div>
          </div>
        </el-descriptions-item>

        <!-- 合约交互信息 -->
        <template v-if="txDetail.is_contract">
          <el-descriptions-item label="合约地址">
            <address-display
              :address="txDetail.contract_address"
              :link="true"
            />
          </el-descriptions-item>

          <el-descriptions-item label="调用函数">
            <div class="func-signature">
              {{ txDetail.func_string }}
              <el-tag
                v-if="txDetail.func_string"
                size="small"
                class="ml-2"
              >
                {{ txDetail.func_string }}
              </el-tag>
            </div>
          </el-descriptions-item>

          <el-descriptions-item v-if="txDetail.input != -1" label="输入参数">
            {{ txDetail.input }}
          </el-descriptions-item>

          <el-descriptions-item label="输出结果">
            {{ txDetail.output }}
          </el-descriptions-item>
        </template>
      </el-descriptions>
    </el-card>
  </transition>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import dayjs from 'dayjs'
import AddressDisplay from './AddressDisplay.vue'
import { ElMessage } from 'element-plus'
import { type Transaction, type CreateTransactionReq, TransactionStatus } from '@/model/dashboardModel'
import { sendTransaction } from '@/http/http.dashboard.ts'
// import { validateEthAddress } from '@/utils/validators'


// 消息提示
const message = ref('')
const messageType = ref('')

// 路由实例
const router = useRouter()

// 复制到剪贴板
const copyToClipboard = (text: string) => {
  navigator.clipboard.writeText(text)
  ElMessage.success('已复制到剪贴板')
}

// 导航到仪表板
const navigateToDashboard = () => {
  router.push({ name: 'home' }) // 假设Dashboard的路由名称为'Dashboard'
}

// 交易数据模型
const transaction = reactive<CreateTransactionReq>({
  from: '0x8f00527c4f08eb89f9158f9fe14545b868e0498d',
  to_address: '',
  value: 0,
  gas_limit: 21000,
  is_contract: false,
  contract_address: '',
  func_string: '',
  input: 0
})

// 合约函数选项
const functionOptions = [
  { value: 'store', label: '存储' },
  { value: 'retrieve', label: '获取' },
]

// 处理合约切换
const handleContractToggle = (val: boolean) => {
  if (!val) {
    transaction.contract_address = ''
    transaction.func_string = ''
  }
}


// 接收交易详情数据
const txDetail = ref<Transaction | null>(null)

const submit = async () => {
  if (transaction.input == 0) {
    transaction.input = -1
  }
  await sendTransaction(transaction).then((res) => {
    if (res != null) {
      txDetail.value = res
    }
  })

  transaction.to_address = ''
  transaction.value = 0
  transaction.is_contract = false
  transaction.input = 0
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

const formatTime = (timestamp: string) => {
  return dayjs(timestamp).format('YYYY-MM-DD HH:mm:ss')
}

const calculateTotalFee = (tx: Transaction) => {
  return (tx.gasUsed * tx.gasPrice)
}

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
