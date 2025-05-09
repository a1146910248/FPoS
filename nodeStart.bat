@echo off
title 节点集群并行启动器
color 0A

:: 检查必要文件
if not exist "bootNode.exe" (
    echo [错误] 未找到 bootNode.exe
    pause
    exit /b 1
)

if not exist "dacNode.exe" (
    echo [错误] 未找到 dacNode.exe
    pause
    exit /b 1
)

if not exist "sequencerNode.exe" (
    echo [错误] 未找到 sequencerNode.exe
    pause
    exit /b 1
)

:: 并行启动所有节点（不等待）
echo 正在并行启动所有节点...
start "BOOT_NODE" bootNode.exe
timeout /t 1 >nul  :: 微小延迟避免窗口重叠

for /l %%i in (1,1,3) do (
    start "DAC_NODE_%%i" dacNode.exe
    timeout /t 1 >nul
)

for /l %%i in (1,1,3) do (
    start "SEQ_NODE_%%i" sequencerNode.exe
    timeout /t 1 >nul
)

echo 所有节点已并行启动！
echo 注意：
echo 1. 每个节点运行在独立窗口
echo 2. 关闭本窗口不会终止节点进程
pause