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

  dac_total: number
  dac_current: number
  dac_term: number
  dacs: string[]
}

export interface Transaction {
  hash: string
  from: string
  to: string
  value: number
  nonce: number
  gasPrice: number
  gasLimit: number
  gasUsed: number
  timestamp: string
  status: TransactionStatus
  signature: string
  is_contract: boolean
  contract_address: string
  input: number
  output: number
  func_string: string
  block_hash?: string  // 所属区块hash，可选
  block_height?: number  // 所属区块hash，可选
  stat_log : StateLog
}

export interface StateLog {
  status: number
  block_hash: string
  block_height: number
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

export interface CreateTransactionReq {
  from: string
  to_address: string
  value: number
  gas_limit: number
  is_contract: boolean
  contract_address: string
  func_string: string
  input: number
}

export interface Block {
  height: number; // uint64 在 TS 中用 number 表示
  hash: string;
  previousHash: string;
  timestamp: string; // time.Time 对应 JS 的 Date
  transactions: Transaction[];
  stateRoot: string;
  txRoot: string; // 交易默克尔根
  proposer: string;
  gasUsed: number; // uint64
  gasLimit: number; // uint64
  votes: BlockVote[];
  signature: string;
  is_sus: boolean;
  rootCommitment?: string | null; // *bls12381.PointG1 对应可选的 PointG1 或 null
  previousRoot?: string | null; // *bls12381.PointG1 对应可选的 PointG1 或 null
  final_proof: string
  kzg_commitment: string
}

export interface BlockVote {
  block_hash: string;        // 区块哈希
  block_height: number;      // 区块高度 (uint64 在 TS 中用 number 表示)
  approve: boolean;          // 是否赞成 (true/false)
  voter_address: string;     // 投票者地址
  signature: string;     // 签名 ([]byte 对应 Uint8Array)
  timestamp: string;           // 时间戳 (time.Time 对应 JS 的 Date)
}
