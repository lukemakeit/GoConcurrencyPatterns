package pipelineDemos

import (
	"fmt"
	"time"
)

// 数据生产者
// 将nums中整数发送给指定的out channel
// 当全部数据发送结束后,channel 关闭
func gen01(nums ...int) <-chan int {
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
func sql01(in <-chan int) <-chan int {
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

func SerialPipeline() {
	genOut := gen01(2, 3, 4)
	sqlOut := sql01(genOut)

	for ret := range sqlOut {
		fmt.Println(ret)
	}
}
