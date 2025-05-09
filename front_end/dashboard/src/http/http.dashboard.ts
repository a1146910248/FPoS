import dashboardRequest from '@/utils/dashboardRequest.ts'
import type { CreateTransactionReq, StatsResp, TransactionList, Transaction, Block } from '@/model/dashboardModel.ts'
import { ElMessage } from 'element-plus'
import type { Base } from '@/model/base.ts'

export const getStats = async(): Promise<StatsResp | null> => {
  return dashboardRequest.get("/dashboard/stats")
    .then(res => {
      const {ok, msg, data} = res.data as Base
      if (ok === 1) {
        return data as StatsResp
      } else {
        ElMessage.error(msg)
        return null
      }
    }).catch(() => {
      ElMessage.error('http请求失败,请刷新页面重试');
      return null;
    })
}

export const getTransactions = async (page: number = 1, limit: number = 20): Promise<TransactionList | null> => {
  return dashboardRequest.get("/dashboard/transactions", {
    params: {
      page,
      limit
    }
  })
    .then(res => {
      const {ok, msg, data} = res.data as Base
      if (ok === 1) {
        return data as TransactionList
      } else {
        ElMessage.error(msg)
        return null
      }
    })
    .catch(() => {
      ElMessage.error('获取交易列表失败,请刷新重试')
      return null
    })
}

export const subscribeToUpdates = (callback: (data: any) => void) => {
  const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const wsUrl = import.meta.env.VITE_NODE_ENV === 'development'
    ? `${wsProtocol}//localhost:8080/dashboard/ws`
    : `${wsProtocol}//${window.location.host}/dashboardApi/dashboard/ws`

  const ws = new WebSocket(wsUrl)

  ws.onopen = () => {
    console.log('WebSocket连接已建立')
  }

  ws.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data)
      callback(data)
    } catch (e) {
      console.error('WebSocket数据解析错误:', e)
    }
  }

  ws.onerror = (error) => {
    console.error('WebSocket错误:', error)
  }

  ws.onclose = () => {
    console.log('WebSocket连接已关闭')
    setTimeout(() => {
      subscribeToUpdates(callback)
    }, 5000)
  }

  return ws
}

export const sendTransaction = async (transaction: CreateTransactionReq): Promise<Transaction | null> => {
  return dashboardRequest.post("/dashboard/postTransactions", transaction, {
    headers: {
      'Content-Type': 'application/json' // 明确指定JSON类型
    }
  })
    .then(res => {
      const { ok, msg, data } = res.data as Base
      if (ok === 1) {
        ElMessage.success('发送成功')
        return data as Transaction
      } else {
        ElMessage.error(msg)
        return null
      }
    })
    .catch(() => {
      ElMessage.error('发送交易失败, 请重试')
      return null
    })
}


export const getTransaction =async(hash:string):Promise<Transaction | null> => {
  return  dashboardRequest.post("/dashboard/getTransaction", `tx_hash=${hash}`, {
    headers: {
      'Content-Type': 'application/x-www-form-urlencoded' // 以表单形式提交
    }})
    .then(res => {
      const {ok, msg, data} = res.data as Base
      if (ok === 1) {
        return data as Transaction
      }else {
        ElMessage.error(msg)
        return null
      }
    }).catch(() => {
      // ElMessage.error('http请求失败,请刷新页面重试');
      return null
    })
}


export const getBlock =async(hash:number):Promise<Block | null> => {
  return  dashboardRequest.post("/dashboard/getBlock", `block_hash=${hash}`, {
    headers: {
      'Content-Type': 'application/x-www-form-urlencoded' // 以表单形式提交
    }})
    .then(res => {
      const {ok, msg, data} = res.data as Base
      if (ok === 1) {
        return data as Block
      }else {
        ElMessage.error(msg)
        return null
      }
    }).catch(() => {
      // ElMessage.error('http请求失败,请刷新页面重试');
      return null
    })
}

// export const getTransactionByHash =async(hash:string):Promise<Transaction | null> => {
//   return  dashboardRequest.post("/dashboard/getTransactionByHash", `tx_hash=${hash}`, {
//     headers: {
//       'Content-Type': 'application/x-www-form-urlencoded' // 以表单形式提交
//     }})
//     .then(res => {
//       const {ok, msg, data} = res.data as Base
//       if (ok === 1) {
//         return data as Transaction
//       }else {
//         ElMessage.error(msg)
//         return null
//       }
//     }).catch(() => {
//       // ElMessage.error('http请求失败,请刷新页面重试');
//       return null
//     })
// }
