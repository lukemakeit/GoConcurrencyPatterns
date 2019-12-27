package BoundedParallelism

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"
)

// 描述: 每隔n秒就打印处理进度,最后再统一打印平方计算结果
func PrintProcessEveryNSeconds() {
	srcNums := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}

	totalCnt := len(srcNums)
	currentIndex := 0                    // 当前处理元素的下标
	currentItem := srcNums[currentIndex] // 当前处理的元素值

	wg := sync.WaitGroup{}
	genChan := make(chan int)
	retChan := make(chan *squreItem)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // 10秒后超时
	ticker := time.NewTicker(3 * time.Second)                                // 每隔3秒打印一次进度

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
				if numItem == 27 {
					// 模拟在处理数字27时发生了错误
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

		for index, srcItem := range srcNums {
			currentIndex = index
			currentItem = srcItem

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
	retMap := make(map[int]int)
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

			retMap[retItem.SrcNum] = retItem.SqureRet
		case <-ticker.C:
			fmt.Printf("....[%d/%d] currentItem:%d\n", currentIndex, totalCnt-1, currentItem)
			break
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

	for k, v := range retMap {
		fmt.Printf("num:%d squreRet=>%d\n", k, v)
	}
	fmt.Println("main goroutine exit...")
}
