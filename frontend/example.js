// 根据环境连接不同的WebSocket地址
const wsUrl = process.env.NODE_ENV === 'production' 
  ? 'ws://your-domain/dashboardApi/dashboard/ws'
  : 'ws://localhost:8080/dashboard/ws';

const ws = new WebSocket(wsUrl);

// 连接建立时的处理
ws.onopen = () => {
  console.log('WebSocket连接已建立');
};

// 接收消息的处理
ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('收到链上数据：', data);
  
  // data的结构如下：
  // {
  //   current_tps: 100,
  //   peak_tps: 150,
  //   total_tx: 1000,
  //   block_height: 500,
  //   active_users: 200,
  //   l1_blocks: 100,
  //   l2_blocks: 400,
  //   l1_balance: "1000000000",
  //   l2_tps: 80
  // }
  
  // 在这里更新您的UI显示
};

// 处理错误
ws.onerror = (error) => {
  console.error('WebSocket错误：', error);
};

// 处理连接关闭
ws.onclose = () => {
  console.log('WebSocket连接已关闭');
  // 可以在这里实现重连逻辑
};

// 在组件卸载时关闭连接
function cleanup() {
  if (ws) {
    ws.close();
  }
} 