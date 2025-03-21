cd C:\Users\DELL
start bootnode -nodekey boot.key -addr :30305
cd G:\geth\eth
start geth --datadir node1 --port 30306 --bootnodes enode://adf791a6af373829117a9c37bd0ebb63d9162285850f44e91b39c073ea8409fe81984eeee12c8571c00a1e176481777d819f15ec62b4dbcfe48ab6c8ba9fdc61@127.0.0.1:0?discport=30305  --networkid 66889 --unlock 0xaca7f9ec8560e83a7c1893516200b13b0293fac3 --password node1/passwd.txt --authrpc.port 8551 --mine --miner.etherbase 0xaca7f9ec8560e83a7c1893516200b13b0293fac3
start geth --datadir node2 --port 30307 --bootnodes enode://adf791a6af373829117a9c37bd0ebb63d9162285850f44e91b39c073ea8409fe81984eeee12c8571c00a1e176481777d819f15ec62b4dbcfe48ab6c8ba9fdc61@127.0.0.1:0?discport=30305  --networkid 66889 --unlock 0x52d0598566648ce0c5de31a925acf9197bfb5c49 --password node2/passwd.txt --authrpc.port 8552 --ipcdisable --http --http.corsdomain https://remix.ethereum.org --allow-insecure-unlock  console 2> 1.log
cd G:\Program\FPoS\front_end\dashboard
start npm run dev
pause
