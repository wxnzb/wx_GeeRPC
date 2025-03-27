## wx_GeeRPC
GeeRPC从零实现Go语言官方的标准库`net/rpc`，并在此基础上，新增了协议交换(protocol exchange)、注册中心(registry)、服务发现(service discovery)、负载均衡(load balance)、超时处理(timeout processing)等特性.
### 服务端与消息编码
客户端固定采用`JSON`编码 Option，后续的 header 和 body 的编码方式由Option 中的 CodeType 指定，服务端首先使用 JSON 解码 Option，然后通过 Option 的CodeType 解码剩余的内容,CodeType中先在实现的是用`Gob`编码
```

| Option{MagicNumber: xxx, CodecType: xxx} | Header{ServiceMethod ...} | Body interface{} |
| <------      固定 JSON 编码      ------>  | <-------   编码方式由 CodeType 决定   ------->|

```
在一次连接中，Option 固定在报文的最开始，Header 和 Body 可以有多个，即报文可能是这样的
```

| Option | Header1 | Body1 | Header2 | Body2 | ...

```
### 服务端
- 实现了Accept方式，net.Listener作为参数，for循环等待socket连接建立，并开启子协程处理，处理过程交给了ServerConn方法
- 子协程中的ServeConn函数主要先通过JSON解码Option,然后交给serveCodec函数循环接收并处理消息
### 客户端
- 客户端支持异步和并发，实现接收响应和发送请求两个功能，并实现了两个暴露给用户的 `RPC` 服务调用接口，`Go` 是一个异步端口，`Call` 是一个同步端口

### 服务注册
- 通过反射，获取某个结构体的所有方法，并且能够通过方法，获取到该方法所有的参数类型与返回值

### 超时处理
- 都是使用使用`time.After()`结合`select+chan`完成
- 客户端超时，有两个地方可能超时，一个是`NewClient`可能超时，一个是`Client`.call可能超时
- 超时设定放在了Option中，客户端连接超时，为Dial添加一层超时处理的外壳dialTimeout
- 服务端超时，`Server.handleRequest` 超时

### 支持http协议
- 主要通过Go标准库中http.Handle实现

### 负载均衡
- 现在了两种：随机选择，轮询调度

### 服务发现与注册中心
- 具体步骤
- 服务端启动后，向注册中心发送注册消息，注册中心得知该服务已经启动，处于可用状态
- 客户端向注册中心询问，当前哪天服务是可用的，注册中心将可用的服务列表返回客户端
- 客户端根据注册中心得到的服务列表，选择其中一个发起调用
### 详细步骤可看3ago总结.md!!!