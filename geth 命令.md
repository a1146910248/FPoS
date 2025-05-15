# geth 命令

创世块（1.12以上版本不支持ethash（Pow），用clique代替（poa））：

其中**extradata**是32 个零字节、所有签名者地址和另外 65 个零字节组成

```json
{
  "config": {
    "chainId": 12345,
    "homesteadBlock": 0,
    "eip150Block": 0,
    "eip155Block": 0,
    "eip158Block": 0,
    "byzantiumBlock": 0,
    "constantinopleBlock": 0,
    "petersburgBlock": 0,
    "istanbulBlock": 0,
    "berlinBlock": 0,
    "clique": {
      "period": 5, //目标块时间
      "epoch": 30000
    }		//clique相关配置
  },
  "difficulty": "1",
  "gasLimit": "8000000",
  "extradata": "0x00000000000000000000000000000000000000000000000000000000000000007df9a875a174b3bc565e6424a0050ebc1b2d1d820000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000", // 指定初始出块私钥
  "alloc": {
    "7df9a875a174b3bc565e6424a0050ebc1b2d1d82": { "balance": "300000" },
    "f41c74c9ae680c1aa78f42e5647a62f353b7bdde": { "balance": "400000" }
  }
}
```



开发者模式启动：`geth --datadir ./devdata --networkid 1008 --port 30303 --http --http.addr 0.0.0.0 --http.vhosts "*"  --http.port 8545 --http.api 'db,net,eth,web3,personal' --http.corsdomain "*"  --dev --dev.period 1 console --rpc.enabledeprecatedpersonal --allow-insecure-unlock 2> 1.log`

--rpc.enabledeprecatedpersonal：启动personal命名空间

--allow-insecure-unlock：允许命令行解锁账户

```sh
geth --goerli --datadir goerli-data --signer=goerli-data/clef/clef.ipc
```

--signer=~/.clef/clef.ipc：指定签名者为clef



调用 clef API：https://github.com/ethereum/go-ethereum/tree/master/cmd/clef#external-api-1

clef --keystore ./ --chainid 1008：启动 clef，--keystore ：keystore位置，chainid：chainid

echo '{"id": 1, "jsonrpc": "2.0", "method": "account_list"}' | nc -U ~/.clef/clef.ipc

```json
{
  "id": 0,
  "jsonrpc": "2.0",
  "method": "account_new",
  "params": []
}
```

签署交易并以 RLP 编码和 JSON 形式响应已签名的交易。

#### Arguments

1. transaction object:
   - `from` [address]: account to send the transaction from
   - `to` [address]: receiver account. If omitted or `0x`, will cause contract creation.
   - `gas` [number]: maximum amount of gas to burn
   - `gasPrice` [number]: gas price
   - `value` [number:optional]: amount of Wei to send with the transaction
   - `data` [data:optional]: input data
   - `nonce` [number]: account nonce

```json
{
  "id": 2,
  "jsonrpc": "2.0",
  "method": "account_signTransaction",
  "params": [
    {
      "from": "0x1923f626bb8dc025849e00f99c25fe2b2f7fb0db",
      "gas": "0x55555",
      "gasPrice": "0x1234",
      "input": "0xabcd",
      "nonce": "0x0",
      "to": "0x07a565b7ed7d7a678680a4c162885bedbb695fe0",
      "value": "0x1234"
    }
  ]
}
```

#### 签署数据

对数据块进行签名并返回计算出的签名。

```
{
  "id": 3,
  "jsonrpc": "2.0",
  "method": "account_signData",
  "params": [
    "data/plain",
    "0x1923f626bb8dc025849e00f99c25fe2b2f7fb0db",
    "0xaabbccdd"
  ]
}
```

#### 挖矿导致CPU占用过高

可以显式规定调用线程数`minner.start(1)` 即为使用一个线程挖矿



### 私有链搭建：

官方文档：https://geth.ethereum.org/docs/fundamentals/private-network

引导节点bootnode

密钥生成：

```sh
bootnode -genkey boot.key
```

节点生成：

```sh
bootnode -nodekey boot.key -addr :30305
```

node1：

```sh
geth --datadir node1 --port 30306 --bootnodes enode://adf791a6af373829117a9c37bd0ebb63d9162285850f44e91b39c073ea8409fe81984eeee12c8571c00a1e176481777d819f15ec62b4dbcfe48ab6c8ba9fdc61@127.0.0.1:0?discport=30305  --networkid 66889 --unlock 0xaca7f9ec8560e83a7c1893516200b13b0293fac3 --password node1/passwd.txt --authrpc.port 8551 --mine --miner.etherbase 0xaca7f9ec8560e83a7c1893516200b13b0293fac3
```

node2：

```sh
geth --datadir node2 --port 30307 --bootnodes enode://adf791a6af373829117a9c37bd0ebb63d9162285850f44e91b39c073ea8409fe81984eeee12c8571c00a1e176481777d819f15ec62b4dbcfe48ab6c8ba9fdc61@127.0.0.1:0?discport=30305  --networkid 66889 --unlock 0x52d0598566648ce0c5de31a925acf9197bfb5c49 --password node2/passwd.txt --authrpc.port 8552 --ipcdisable --http --http.corsdomain https://remix.ethereum.org --allow-insecure-unlock  console 2> 1.log
```

--ipcdisable：禁用ipc，否则同时起多个node会报错，据说linux系统不会有这样的问题

--allow-insecure-unlock：使其能在http时解锁账户，否则会因为不安全而报错，从而导致连不上remix

### 配置文件

将node1转为配置文件方式：

```sh
geth --datadir node1 --port 30306 --bootnodes enode://421b1d981416ba5095c1e04d6ab7056e79efa672c0eab117e69787f6e1707d0a3279f4b1586413a7371d24d43ee5fafc2ac010679503b7e6c61665100d5c08a7@127.0.0.1:0?discport=30305  --networkid 66889 --unlock 0xaca7f9ec8560e83a7c1893516200b13b0293fac3 --password node1/passwd.txt --authrpc.port 8551 --mine --miner.etherbase 0xaca7f9ec8560e83a7c1893516200b13b0293fac3 dumpconfig > node1-config.toml
```

主要选项为dumpconfig，将其输入到 > 后的文件中，保存的文件应为UTF-8格式（用vscode修改），win中默认的保存格式将导致无法从配置文件中启动

dumpconfig前为需要加入的特性，使用dumpconfig命令会自动生成配置文件，内部可以自己细调

使用配置文件启动不会引入挖矿和解锁账户等配置，需要自己额外增加

使用：

```sh
geth --networkid 66889 --unlock 0xaca7f9ec8560e83a7c1893516200b13b0293fac3 --password node1/passwd.txt --mine --config ./node1-config.toml 
```



![image-20240411170942713](https://gitee.com/SuzzHmm/picture/raw/master/img/202404111709788.png)

默认为快照同步

full sync：--syncmode full

archive node：--syncmode full --gcmode archive

轻节点：--syncmode light

**不连接共识客户端geth无法在pos网络中同步**

**轻量节点很少且不适用与pos网络**

源码由EVM工具，试试能不能改造



### 当节点出错

需要清除数据目录，防止使用到错误的主网数据：

```bash
geth removedb --datadir node2
```

然后重新初始化数据目录：

```bash
geth --datadir node2 init genesis.json
```

转账：

```js
web3.eth.sendTransaction({from:'0xaca7f9ec8560e83a7c1893516200b13b0293fac3', to:'52d0598566648ce0c5de31a925acf9197bfb5c49', value: web3.toWei(1.8 ,'ether')})
```



### crul请求

创建 `request.json` 文件：

```
{
  "jsonrpc": "2.0",
  "method": "eth_getCode",
  "params": ["0xContractAddress", "latest"],
  "id": 1
}
```

执行 curl 命令：

```
curl.exe -X POST -H "Content-Type: application/json" -d "@request.json" http://localhost:8545
```

方法调用json

```json
{
  "jsonrpc": "2.0",
  "method": "eth_sendTransaction",
  "params": [{
    "from": "0x52d0598566648ce0c5de31a925acf9197bfb5c49",  
    "to": "0xba3f2bae5586080a20d3b0667ab129bc597cc315",     
    "gas": "0x30d40",              
    "gasPrice": "0x3b9aca00",     
    "value": "0x0",                
    "data": "0x0f2e5b6c"           
  }],
  "id": 1
}
```

验证结果json

```json
{
  "jsonrpc": "2.0",
  "method": "eth_getTransactionReceipt",
  "params": ["0x4510cb9998864f3b49fd05a40c20158299a4f6173f9e1e52021cd687dc949812"],
  "id": 1
}
```

