package DupDataProcessOnlyOnce

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// 描述: 计算列表中每个元素的平方值,列表中元素有重复,我们希望重复元素平方计算只执行一次。
// - 全部计算结果得到后统一打印;
// - 如果有错误,则直接返回退出;
// - 如果超时,则直接返回退出;
// 缺点: 该方法锁粒度很重,如计算元素4平方值, 其他元素读取 或 写入的child goroutine 都会被卡住
func DupDataProcessOnlyOnceByLock() {
	startTime := time.Now()

	srcData := []int{1, 2, 2, 33, 3, 3, 4, 4, 4, 4, 5, 6, 7, 7, 8, 8, 9, 10, 9, 11, 12, 8, 13, 11, 15, 17, 15, 2}
	retUniq := make(map[int]int) //保证每个元素只执行一次求平方操作

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second) // 10秒后超时
	rmu := &sync.RWMutex{}
	limiter := 5 //控制并发度为3
	genChan := make(chan int)
	errChan := make(chan error)
	finish := make(chan bool)
	wg := sync.WaitGroup{}

	for woker := 0; woker < limiter; woker++ {
		// 消费者
		wg.Add(1)
		go func() {
			defer wg.Done()

			for genItem := range genChan {
				rmu.RLock() // 加读锁
				ret, ok := retUniq[genItem]
				if ok == true {
					// 此时代表已经获取到该元素平方值
					rmu.RUnlock()
					continue
				}
				rmu.RUnlock()

				// 尝试获取该元素平方值
				// rmu.Lock()将把所有child goroutine读取给卡住,比如获取4的平方值,此时其他child goroutine读取1 2 3 ...平方值也会被卡住
				// 直到求平方值这个耗时操作完全结束为止,一句话 锁很重
				rmu.Lock() // 加写锁
				ret, ok = retUniq[genItem]
				if ok == false { // 这一步判断是必不可少的,否则可能多个child goroutine对同一个元素进行求平方操作
					if genItem == 77 {
						// 求7平方值时发生错误
						err := fmt.Errorf("num:%d do squre,unknow error occur", genItem)
						select {
						case errChan <- err:
							break
						case <-ctx.Done():
							break
						}
						rmu.Unlock()
						continue
					}

					time.Sleep(1 * time.Second)
					ret = (genItem) * (genItem)
					retUniq[genItem] = ret
				}
				rmu.Unlock()
			}
		}()
	}
	go func() {
		// 生产者
		defer close(genChan)

		for _, srcItem := range srcData {
			select {
			case genChan <- srcItem:
			case <-ctx.Done():
				return
			}
		}
	}()

	go func() {
		wg.Wait()
		close(finish) // 正常结束
	}()

	var err error
	select {
	case <-finish:
		break
	case err = <-errChan: // 发生错误
		break
	case <-ctx.Done(): // 超时
		err = fmt.Errorf("timeout ")
		break
	}
	cancel()
	//time.Sleep(2 * time.Second)
	//fmt.Printf("after cancel numGoroutine:%d\n", runtime.NumGoroutine())

	fmt.Printf("DupDataProcessOnlyOnceByLock take %.2f seconds \n", time.Now().Sub(startTime).Seconds())

	if err != nil {
		fmt.Printf("err detail:%s\n", err)
	} else {
		//打印结果
		for _, srcItem := range srcData {
			ret, _ := retUniq[srcItem]
			fmt.Printf("num:%d ret=>%d\n", srcItem, ret)
		}
	}
}
