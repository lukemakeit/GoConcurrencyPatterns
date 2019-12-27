package pipelineDemos

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

// 数据生产者
// 将nums中整数发送给指定的out channel
// 当全部数据发送结束后,channel 关闭
// 如果 done channel被关闭,则goroutine直接退出
func genWithDone(done <-chan struct{}, nums ...int) <-chan int {
	out := make(chan int)

	go func() {
		defer close(out)
		for _, n := range nums {
			select {
			case out <- n:
			case <-done:
				return
			}
		}
	}()
	return out
}

// 从input channel中读取数据并计算平方结果, 输出到 out channel中
// 如果 done channel被关闭,则goroutine直接退出
func sqlWithDone(done <-chan struct{}, in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		for n := range in {
			time.Sleep(1 * time.Second)
			select {
			case out <- n * n:
			case <-done:
				return
			}
		}
		close(out)
	}()
	return out
}

// 为每个input channel启动一个goroutine
// 将所有input channel 结果合并到一个 out channel 中
// 如果 done channel被关闭,则goroutine直接退出
func mergeWithDone(done <-chan struct{}, cs ...<-chan int) <-chan int {
	var wg sync.WaitGroup
	out := make(chan int)

	for _, c := range cs {
		// 为每个input channel启动一个goroutine
		// 该goroutine将input channel中值拷贝输出到out channel,直到input channel关闭 或 done channel 关闭
		wg.Add(1)
		go func(c <-chan int) {
			defer wg.Done()
			for n := range c {
				select {
				case out <- n:
				case <-done:
					return
				}
			}
		}(c)
	}
	// 启动一个goroutine 在所有input channel关闭后,关闭out channel
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}
func CanceledWithoutFinish() {
	fmt.Println("now in PiplineSquareTest03 ...")

	done := make(chan struct{})
	genOut := genWithDone(done, 2, 3, 4, 5, 6, 7, 8)
	//并发处理来自于genOut channel中的数据
	sqlOut01 := sqlWithDone(done, genOut)
	sqlOut02 := sqlWithDone(done, genOut)
	sqlOut03 := sqlWithDone(done, genOut)

	//从channel sqlOut01 sqlOut02 sqlOut03 合并后的 channel 中接收数据
	mergeOut := mergeWithDone(done, sqlOut01, sqlOut02, sqlOut03)
	for item := range mergeOut {
		if item == 9 {
			fmt.Printf("error occur\n")
			break
		}
		fmt.Printf("ret=>%d\n", item)
	}
	close(done)
	time.Sleep(2*time.Second)
	fmt.Printf("after done numGoroutine:%d\n",runtime.NumGoroutine())
}
