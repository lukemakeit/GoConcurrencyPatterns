package BoundedParallelism

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"
)

// 超时 或 错误发生时,关闭done channel,以便所有child goroutine(无论是生成者 还是 消费者都能正确退出)
func TimeOut01() {
	srcNums := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}

	wg := sync.WaitGroup{}
	genChan := make(chan int)
	retChan := make(chan *squreItem)
	done := make(chan struct{})
	//控制并发度为2,某一时刻只有2个child goroutine在运行中
	limit := 2
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
				if numItem == 19 {
					// 模拟在处理数字19时发生了错误
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

	//超时处理
	timeoutSemaphore := time.NewTimer(3 * time.Second)

	var retItem *squreItem
	ok := false
	for {
		// 第三stage: 消费者
		select {
		case retItem, ok = <-retChan:
			if !ok {
				// 数据全部处理完了
				fmt.Println("End.")
				break
			}
			if retItem.Err != nil {
				// 某个元素处理发生了错误
				fmt.Printf("num:%d err:%s\n", retItem.SrcNum, retItem.Err)
				ok = false
				break
			}

			fmt.Printf("num:%d squreRet=>%d\n", retItem.SrcNum, retItem.SqureRet)
		case <-timeoutSemaphore.C:
			ok = false
			fmt.Println("Timeout.")
			break
		}
		if !ok {
			break
		}
	}
	close(done)
	time.Sleep(2 * time.Second)
	fmt.Printf("after close done numGoroutine:%d\n", runtime.NumGoroutine())
	fmt.Println("main goroutine exit...")
}

// 超时 或 错误发生时,关闭done channel,以便所有child goroutine(无论是生成者 还是 消费者都能正确退出)
func TimeOut02() {
	srcNums := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}

	wg := sync.WaitGroup{}
	genChan := make(chan int)
	retChan := make(chan *squreItem)
	done := make(chan struct{})
	//控制并发度为2,某一时刻只有2个child goroutine在运行中
	limit := 2
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
				if numItem == 19 {
					// 模拟在处理数字19时发生了错误
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

	//超时处理
	timeoutSemaphore := make(chan bool)
	go func() {
		//3秒后,发送超时信号
		time.Sleep(3 * time.Second)
		timeoutSemaphore <- false
	}()

	var retItem *squreItem
	ok := false
	for {
		// 第三stage: 消费者
		select {
		case retItem, ok = <-retChan:
			if !ok {
				// 数据全部处理完了
				fmt.Println("End.")
				break
			}
			if retItem.Err != nil {
				// 某个元素处理发生了错误
				fmt.Printf("num:%d err:%s\n", retItem.SrcNum, retItem.Err)
				ok = false
				break
			}

			fmt.Printf("num:%d squreRet=>%d\n", retItem.SrcNum, retItem.SqureRet)
		case ok = <-timeoutSemaphore:
			fmt.Println("Timeout.")
			break
		}
		if !ok {
			break
		}
	}
	close(done)
	time.Sleep(2 * time.Second)
	fmt.Printf("after close done numGoroutine:%d\n", runtime.NumGoroutine())
	fmt.Println("main goroutine exit...")
}

// 超时 或 错误发生时,关闭done channel,以便所有child goroutine(无论是生成者 还是 消费者都能正确退出)
func TimeOutWithContext() {
	srcNums := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}

	wg := sync.WaitGroup{}
	genChan := make(chan int)
	retChan := make(chan *squreItem)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // 10秒后超时

	//控制并发度为2,某一时刻只有2个child goroutine在运行中
	limit := 2
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
				if numItem == 7 {
					// 模拟在处理数字7时发生了错误
					err = fmt.Errorf("unknown error occured")
				}

				retItem := &squreItem{
					numItem,
					numItem * numItem,
					err,
				}
				select {
				case retChan <- retItem:
				case <-ctx.Done(): // 发生错误 或 已经超时
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
			case <-ctx.Done(): // 发生错误 或 已经超时
				break
			}
		}
	}()

	go func() {
		wg.Wait() // 等待所有child goroutine处理完成后关闭retChan,此时main goroutine中接收结果的for循环将退出
		close(retChan)
	}()

	var retItem *squreItem
	ok := false
	for {
		// 第三stage: 消费者
		select {
		case retItem, ok = <-retChan:
			if !ok {
				// 数据全部处理完了
				fmt.Println("End.")
				break
			}
			if retItem.Err != nil {
				// 某个元素处理发生了错误
				fmt.Printf("num:%d err:%s\n", retItem.SrcNum, retItem.Err)
				ok = false
				break
			}

			fmt.Printf("num:%d squreRet=>%d\n", retItem.SrcNum, retItem.SqureRet)
		case <-ctx.Done():
			ok = false
			fmt.Println("Timeout.")
			break
		}
		if !ok {
			break
		}
	}
	cancel()
	time.Sleep(2 * time.Second)
	fmt.Printf("after close cancel numGoroutine:%d\n", runtime.NumGoroutine())
	fmt.Println("main goroutine exit...")
}
