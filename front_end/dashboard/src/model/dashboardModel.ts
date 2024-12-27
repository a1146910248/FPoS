export interface StatsResp {
  current_tps:number
  peak_tps:number
  total_tx: number
  block_height:number
  active_users:number
  l1_blocks:number
  l2_blocks:number
  l1_balance:string
  l2_tps:number

  validator_count: number;      // 总验证者数量
  active_validator_count: number;  // 活跃验证者数量
  current_sequencer: string;    // 当前排序器地址
  current_proposers: string[];  // 当前提案者列表(包含排序器)
}

export interface Transaction {
  hash: string
  from: string
  to: string
  value: number
  nonce: number
  gas_price: number
  gas_limit: number
  gas_used: number
  timestamp: string
  status: TransactionStatus
  block_hash?: string  // 所属区块hash，可选
}

export enum TransactionStatus {
  Pending = 0,        // 在交易池中等待
  Confirmed = 1,      // 已被区块确认
  L1Submitting = 2,   // 正在提交到L1
  L1Confirmed = 3,    // L1确认成功
  L1Failed = 4,       // L1确认失败
}

export interface TransactionList {
  total: number
  list: Transaction[]
}
