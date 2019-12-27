package BoundedParallelism

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

// 描述: 错误发生时,关闭done channel,以便所有child goroutine(无论是生成者 还是 消费者都能正确退出)
func CanceledErrOccur() {
	srcNums := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}

	wg := sync.WaitGroup{}
	genChan := make(chan int)
	retChan := make(chan *squreItem)
	done := make(chan struct{})
	//控制并发度为4,某一时刻只有4个child goroutine在运行中
	limit := 4
	for worker := 0; worker < limit; worker++ {
		// 第二stage:进行平方计算
		wg.Add(1)
		go func() {
			defer wg.Done()
			for numItem := range genChan {
				if numItem%10 == 0 {
					continue // 如果部分源数据不想处理,用continue 跳过,而不是执行return
				}
				time.Sleep(1 * time.Second)

				var err error
				if numItem == 11 {
					// 模拟在处理数字11时发生了错误
					err = fmt.Errorf("unknown error occured")
				}

				retItem := &squreItem{
					numItem,
					numItem * numItem,
					err,
				}
				select {
				case retChan <- retItem:
				case <-done: // done channel被关闭时,child goroutine退出
					return
				}
			}
		}()
	}
	go func() {
		// 第一stage:生产者
		defer close(genChan)

		for _, srcItem := range srcNums {
			select {
			case genChan <- srcItem:
			case <-done: // done channel被关闭时, 生产者不再产生数据
				return
			}
		}
	}()

	go func() {
		wg.Wait() // 等待所有child goroutine处理完成后关闭retChan,此时main goroutine中接收结果的for循环将退出
		close(retChan)
	}()

	for retItem := range retChan {
		// 第三stage: 消费者
		if retItem.Err != nil {
			// 发生错误后直接退出
			fmt.Printf("num:%d err:%s\n", retItem.SrcNum, retItem.Err)
			close(done)
			time.Sleep(2 * time.Second)
			fmt.Printf("after done numGoroutine:%d\n", runtime.NumGoroutine())
			return
		}
		fmt.Printf("num:%d  squreRet=>%d\n", retItem.SrcNum, retItem.SqureRet)
	}

	fmt.Println("main goroutine exit...")
}
