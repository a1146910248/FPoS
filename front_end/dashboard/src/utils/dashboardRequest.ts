import axios from 'axios';
import type { AxiosResponse } from 'axios';
import { ElMessage } from 'element-plus'

const dashBoardRequest = axios.create({
  // 前端vite.config.ts中配置的代理
  baseURL: '/dashboardApi',
  timeout: 30000
})
// request 拦截器
// 可以自请求发送前对请求做一些处理
// 比如统一加token，对请求参数统一加密
dashBoardRequest.interceptors.request.use(
);

// response 拦截器
// 可以在接口响应后统一处理结果
dashBoardRequest.interceptors.response.use(
  (response: AxiosResponse) => {
    let res = response.data;
    // 如果是返回的文件
    if (response.config.responseType === 'blob') {
      return response;
    }
    // 兼容服务端返回的字符串数据
    if (typeof res === 'string') {
      res = res ? JSON.parse(res) : res;
    }
    return response;
  },
  (error: any) => {
    if (error.response) {
      switch (error.response.status) {
        // 401即令牌失效
        case 401:
          ElMessage.error("暂未登录，请登录后再试！")
          // 这里写清除token的代码
          localStorage.removeItem("token")
      }
      console.log('' + error); // for debug
      return Promise.reject(error);
    }
  }
)

export default dashBoardRequest;
