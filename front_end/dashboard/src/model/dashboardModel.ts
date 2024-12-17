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
