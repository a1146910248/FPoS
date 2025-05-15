# EVM移植遇到的问题

移植的EVM若要支持PUSH0，得是Shanghai之后版本，在`interpreter.go`的NewEVMInterpreter方法中将版本固定为最新的cancun

![image-20240428113137033](https://gitee.com/SuzzHmm/picture/raw/master/img/202404281131087.png)

这样启用的op table中就会包含PUSH0操作码

![image-20240428113447554](https://gitee.com/SuzzHmm/picture/raw/master/img/202404281134580.png)

在测试中：

以下的内容十分重要

使用remix做编译，可以指定编译器版本，EVM版本（Shanghai之后的编译的字节码会包含PUSH0），ABI也需保留，之后打开编译详情。

<img src="https://gitee.com/SuzzHmm/picture/raw/master/img/202404281139190.png" alt="image-20240428113903165" style="zoom: 80%;" />

翻到最下面的ASSEMBLY条目，里面是合约编译后的汇编代码，如默认的Storage合约：

![image-20240428114151781](https://gitee.com/SuzzHmm/picture/raw/master/img/202404281141807.png)

其中.code 部分为合约的部署和初始化相关的内容，如构造方法，这部分代码不会保存到EVM的持久化数据库中，相对的是.data部分，这部分是真正的合约调用代码，会持久化到DB中。

BYTECODE部分是完整的字节码，包含.code部分与.data部分，与汇编为一一对应关系

![image-20240428114555212](https://gitee.com/SuzzHmm/picture/raw/master/img/202404281145232.png)

但是在测试调用时我们仅需要.data部分，这部分可以在**RUNTIME BYTECODE**中的最下面找到

![image-20240428114705665](https://gitee.com/SuzzHmm/picture/raw/master/img/202404281147685.png)

字节码的对应关系可以查询 https://www.evm.codes/

**注意：**在地址的decode中需要去掉复制的0x！否则取不到，无论是账户地址还是合约调用的input，只要是`hex.DecodeString`

![image-20240428114923134](https://gitee.com/SuzzHmm/picture/raw/master/img/202404281149154.png)

