# GoConcurrencyPatterns
**Go语言常用并发模式**

并发是Go语言最重要特性之一,正是因为Go语言并发支持，我们得以高效利用机器上的CPU IO 网络等资源，缩短程序完成处理的时间。然而我们面对的场景、需求如此复杂，针对不同情况我们都需要编写不同的并发程序去处理。

下面我们用"数字求平方"操作模拟一个耗时1s 负载较高的操作行为，整理了日常工作中几种常见场景的并发编写方法，希望能对大家日常工作有所帮助。

#### 管道pipeline
关于管道，我们可以理解为, 它是由一系列通过 chanel 连接起来的 stage 组成，而每个 stage 都是由一组运行着相同函数的 goroutine 组成。每个 stage 的 goroutine 通常会执行如下的一些工作：
- 从上游的输入 channel 中接收数据;
- 对接收到的数据进行一些处理，（通常）并产生新的数据;
- 将数据通过输出 channel 发送给下游;

除了第一个 stage 和最后一个 stage ，每个 stage 都包含一定数量的输入和输出 channel。第一个 stage 只有输出，通常会把它称为 "生产者"，最后一个 stage 只有输入，通常我们会把它称为 "消费者"。

#### 关于Fan-out 和 Fan-in

当多个函数从一个 channel 中读取数据，直到 channel 关闭，这称为 Fan-out; 这为我们提供了 在多个worker间分配任务 提供了一种方法，这一组worker可以并发的利用CPU和I/O。

当一个函数从多个 channel 中读取数据，直到所有 channel 关闭，这称为 Fan-in。Fan-in 是通过将多个channel中的数据合并到一个输出channel实现的，当所有输入channel关闭，输出channel也关闭。


#### pipeline并发案例
- [普通流水线(串行)](https://github.com/lukexwang/GoConcurrencyPatterns/blob/96b3b9420cdd7276491acd62086c2445cef184b3/pipelineDemos/SerialPipeline.go#L36)
- [Fan-out 和 Fan-in](https://github.com/lukexwang/GoConcurrencyPatterns/blob/96b3b9420cdd7276491acd62086c2445cef184b3/pipelineDemos/FanInFanOut.go)
- [中途取消(发生错误后不再继续)](https://github.com/lukexwang/GoConcurrencyPatterns/blob/96b3b9420cdd7276491acd62086c2445cef184b3/pipelineDemos/CanceledWithoutFinish.go#L77)

#### 并发控制案例
- 无需接收child goroutine返回结果
    - [非 bufferChannel 完成并发控制(推荐)](https://github.com/lukexwang/GoConcurrencyPatterns/blob/96b3b9420cdd7276491acd62086c2445cef184b3/BoundedParallelism/notDealChildRet.go#L10)
    - [bufferChannel 完成并发控制方式一](https://github.com/lukexwang/GoConcurrencyPatterns/blob/96b3b9420cdd7276491acd62086c2445cef184b3/BoundedParallelism/notDealChildRet.go#L42)
    - [bufferChannel 完成并发控制方式二](https://github.com/lukexwang/GoConcurrencyPatterns/blob/96b3b9420cdd7276491acd62086c2445cef184b3/BoundedParallelism/notDealChildRet.go#L71)
- 需接收child goroutine返回结果
    - [非 bufferChannel 完成并发控制(推荐)](https://github.com/lukexwang/GoConcurrencyPatterns/blob/96b3b9420cdd7276491acd62086c2445cef184b3/BoundedParallelism/DealChildRet.go#L16)
    - [bufferChannel 完成并发控制方式一](https://github.com/lukexwang/GoConcurrencyPatterns/blob/96b3b9420cdd7276491acd62086c2445cef184b3/BoundedParallelism/DealChildRet.go#L68)
    - [bufferChannel 完成并发控制方式二](https://github.com/lukexwang/GoConcurrencyPatterns/blob/96b3b9420cdd7276491acd62086c2445cef184b3/BoundedParallelism/DealChildRet.go#L118)
    - [中途取消(发生错误后不再继续)](https://github.com/lukexwang/GoConcurrencyPatterns/blob/96b3b9420cdd7276491acd62086c2445cef184b3/BoundedParallelism/CanceledErrOccur.go)
    - 超时控制,超时情况下保证所有child goroutine正常退出
        - [Timer方式](https://github.com/lukexwang/GoConcurrencyPatterns/blob/96b3b9420cdd7276491acd62086c2445cef184b3/BoundedParallelism/TimeOut.go#L12)
        - [单独的child goroutine发送超时信号](https://github.com/lukexwang/GoConcurrencyPatterns/blob/96b3b9420cdd7276491acd62086c2445cef184b3/BoundedParallelism/TimeOut.go#L106)
        - [context方式](https://github.com/lukexwang/GoConcurrencyPatterns/blob/96b3b9420cdd7276491acd62086c2445cef184b3/BoundedParallelism/TimeOut.go#L204)
    - [定期打印处理进度](https://github.com/lukexwang/GoConcurrencyPatterns/blob/96b3b9420cdd7276491acd62086c2445cef184b3/BoundedParallelism/PrintProcessEveryNSeconds.go)
    - 根据输入顺序接收每个元素的结果
        - [bufferChannel 控制并发度](https://github.com/lukexwang/GoConcurrencyPatterns/blob/96b3b9420cdd7276491acd62086c2445cef184b3/BoundedParallelism/ReturnByInputOrder.go#L50)
        - [非bufferChannel 控制并发度(推荐)](https://github.com/lukexwang/GoConcurrencyPatterns/blob/96b3b9420cdd7276491acd62086c2445cef184b3/BoundedParallelism/ReturnByInputOrder.go#L118)
- 保证重复数据只被处理一次
    - [传统lock方式实现](https://github.com/lukexwang/GoConcurrencyPatterns/blob/96b3b9420cdd7276491acd62086c2445cef184b3/DupDataProcessOnlyOnce/ByLock.go#L15)
    - [lock+channel实现方式(推荐)](https://github.com/lukexwang/GoConcurrencyPatterns/blob/96b3b9420cdd7276491acd62086c2445cef184b3/DupDataProcessOnlyOnce/betterWay.go#L113)
    
参考资料:  
[Go 译文之如何构建并发 Pipeline](https://segmentfault.com/a/1190000019984518)  
[Go Concurrency Patterns: Pipelines and cancellation](https://blog.golang.org/pipelines)  
[Golang 中的并发限制与超时控制](https://www.jianshu.com/p/42e89de33065)
