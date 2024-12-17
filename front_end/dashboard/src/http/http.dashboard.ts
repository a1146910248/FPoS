import dashboardRequest from '@/utils/dashboardRequest.ts'
import type { StatsResp } from '@/model/dashboardModel.ts'
import { ElMessage } from 'element-plus'
import type { Base } from '@/model/base.ts'

// 按页获取文章
export const getStats =async():Promise<StatsResp | null> => {
  return  dashboardRequest.get("/dashboard/stats")
    .then(res => {
      const {ok, msg, data} = res.data as Base
      if (ok === 1) {
        return data as StatsResp
      }else {
        ElMessage.error(msg)
        return null
      }
    }).catch(() => {
      ElMessage.error('http请求失败,请刷新页面重试');
      return null;
    })
}

// eslint-disable-next-line
export const subscribeToUpdates = (callback: (data: any) => void) => {
  const ws = new WebSocket(`ws://${window.location.host}/api/dashboard/ws`)
  ws.onmessage = (event) => {
    callback(JSON.parse(event.data))
  }
  return ws
}
