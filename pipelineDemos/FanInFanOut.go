package pipelineDemos

import (
	"fmt"
	"sync"
	"time"
)

// Multiple functions can read from the same channel until that channel is closed; this is called fan-out.
//
// A function can read from multiple inputs and proceed until all are closed by multiplexing the input channels onto
// a single channel that's closed when all the inputs are closed. This is called fan-in.

// 数据生产者
// 将nums中整数发送给指定的out channel
// 当全部数据发送结束后,channel 关闭
func gen02(nums ...int) <-chan int {
	out := make(chan int)

	go func() {
		for _, n := range nums {
			out <- n
		}
		close(out)
	}()
	return out
}

//从input channel中读取数据并计算平方结果, 输出到 out channel中
func sql02(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		for n := range in {
			time.Sleep(1*time.Second) // 用计算平方模拟一个比较消耗资源的操作
			out <- n * n
		}
		close(out)
	}()
	return out
}

// 为每个input channel启动一个goroutine
// 将所有input channel 结果合并到一个 out channel 中
func merge(cs ...<-chan int) <-chan int {
	var wg sync.WaitGroup
	out := make(chan int)

	for _, c := range cs {
		// 为每个input channel启动一个goroutine
		// 该goroutine将input channel中值拷贝输出到out channel,直到input channel关闭
		wg.Add(1)
		go func(c <-chan int) {
			defer wg.Done()
			for n := range c {
				out <- n
			}
		}(c)
	}
	// 启动一个goroutine 在所有goroutine结束后(也就是所有input channel关闭后) 关闭out channel
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}
func FanInFanOut() {
	genOut := gen01(2, 3, 4, 5, 6, 7, 8)
	//3个并发处理来自于genOut channel中的数据
	sqlOut01 := sql02(genOut)
	sqlOut02 := sql02(genOut)
	sqlOut03 := sql02(genOut)

	//从channel sqlOut01 sqlOut02 sqlOut03 合并后的 channel 中接收数据
	for ret := range merge(sqlOut01, sqlOut02, sqlOut03) {
		fmt.Println(ret)
	}
}


