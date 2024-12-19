import React, { useEffect, useState } from 'react';

const ChainStats = () => {
  const [stats, setStats] = useState(null);
  
  useEffect(() => {
    const wsUrl = process.env.NODE_ENV === 'production' 
      ? 'ws://your-domain/dashboardApi/dashboard/ws'
      : 'ws://localhost:8080/dashboard/ws';
      
    const ws = new WebSocket(wsUrl);
    
    ws.onmessage = (event) => {
      const data = JSON.parse(event.data);
      setStats(data);
    };
    
    // 清理函数
    return () => {
      ws.close();
    };
  }, []);
  
  if (!stats) return <div>加载中...</div>;
  
  return (
    <div>
      <h2>链上数据</h2>
      <p>当前TPS: {stats.current_tps}</p>
      <p>峰值TPS: {stats.peak_tps}</p>
      <p>总交易数: {stats.total_tx}</p>
      <p>区块高度: {stats.block_height}</p>
      <p>活跃用户: {stats.active_users}</p>
      <p>L1区块数: {stats.l1_blocks}</p>
      <p>L2区块数: {stats.l2_blocks}</p>
      <p>L1余额: {stats.l1_balance}</p>
      <p>L2 TPS: {stats.l2_tps}</p>
    </div>
  );
};

export default ChainStats; 