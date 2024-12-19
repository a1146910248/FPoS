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
  status: number
}

export interface TransactionList {
  total: number
  list: Transaction[]
}
