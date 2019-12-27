package BoundedParallelism

import (
	"fmt"
	"sync"
	"time"
)

type squreItem struct {
	SrcNum   int
	SqureRet int
	Err      error
}

//推荐使用,同时运行的 limit 个 child goroutine, 切main goroutine 会及时处理child goroutine的返回结果
func DealChildRet() {
	srcNums := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}

	wg := sync.WaitGroup{}
	genChan := make(chan int)
	retChan := make(chan *squreItem)
	//控制并发度为4,某一时刻只有4个child goroutine在运行中
	limit := 4
	for worker := 0; worker < limit; worker++ {
		// 消费者
		wg.Add(1)
		go func() {
			defer wg.Done()
			for numItem := range genChan {
				if numItem%10 == 0 {
					continue // 如果部分源数据不想处理,用continue 跳过,而不是执行return
				}
				time.Sleep(1 * time.Second)

				retItem := &squreItem{
					numItem,
					numItem * numItem,
					nil,
				}
				retChan <- retItem
			}
		}()
	}
	go func() {
		// 生产者
		defer close(genChan)

		for _, srcItem := range srcNums {
			genChan <- srcItem
		}
	}()

	go func() {
		wg.Wait() // 等待所有child goroutine处理完成后关闭retChan,此时main goroutine中接收结果的for循环将退出
		close(retChan)
	}()

	for retItem := range retChan {
		fmt.Printf("num:%d  squreRet=>%d\n", retItem.SrcNum, retItem.SqureRet)
	}

	fmt.Println("main goroutine exit...")
}

// len(srcNums)较小时可以使用
// 因为同时发起len(scrNums)个child goroutine,而运行中的child goroutine个数是len(semaphore), 可能会占用较多内存.
// 不过main goroutine 会及时接收child goroutine的结果,不会导致阻塞
func DealChildRetBuffChan01() {
	srcNums := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}

	wg := sync.WaitGroup{}
	retChan := make(chan *squreItem)
	//控制并发度为4
	semaphore := make(chan bool, 4)
	for _, srcItem := range srcNums {
		wg.Add(1)
		go func(numItem int) {
			defer func() {
				wg.Done()
			}()
			// 通过bufferChan控制并发,某一时刻有len(srcNums)个child goroutine运行中,但是很多goroutine会卡在下面这一步
			// 如果len(srcNums)很大,个人感觉这种方式也不是很好
			semaphore <- true
			defer func() {
				<-semaphore
			}()

			if numItem%10 == 0 {
				return // 如果部分源数据不想处理,用return直接退出child goroutine
			}
			// 同时执行耗时操作的只有 4个 child goroutine
			time.Sleep(1 * time.Second)

			retItem := &squreItem{
				numItem,
				numItem * numItem,
				nil,
			}
			retChan <- retItem
		}(srcItem)
	}

	go func() {
		wg.Wait() // 等待所有child goroutine处理完成后关闭retChan,此时main goroutine中接收结果的for循环将退出
		close(retChan)
	}()

	for retItem := range retChan {
		fmt.Printf("num:%d  squreRet=>%d\n", retItem.SrcNum, retItem.SqureRet)
	}

	fmt.Println("main goroutine exit...")
}

// 不推荐使用,因为main goroutine 在最后一批 child goroutine 跑起来后才会接收 child goroutine的结果
// 如果每个 child goroutine 返回结果较大,可能占用大量内存
// 其次对需要及时了解处理情况的场景也不适合
func DealChildRetBuffChan02() {
	srcNums := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}

	wg := sync.WaitGroup{}
	retChan := make(chan *squreItem)
	//控制并发度为4
	semaphore := make(chan bool, 4)
	for _, srcItem := range srcNums {
		wg.Add(1)
		// 控制同时运行的child goroutine数量为length(semaphore),本示例中是4
		// 该方式在通过 semaphore <- true 控制并发度时,也造成了main goroutine长时间卡主在当前for循环中。只有在最后4个goroutine开始时，main goroutine才能跳出当前for循环
		// 可能最高导致 length(srcNums) - length(semaphore)个child goroutine在等待结果被main goroutine接收
		semaphore <- true
		go func(numItem int) {
			defer func() {
				wg.Done()
			}()

			if numItem%10 == 0 {
				return // 如果部分源数据不想处理,用return直接退出child goroutine
			}

			time.Sleep(1 * time.Second)

			retItem := &squreItem{
				numItem,
				numItem * numItem,
				nil,
			}
			// <- semaphore 不能放在 retChan <- retItem 后执行(更不能放在child goroutine的defer中)
			// 因为main goroutine还没跳出当前for循环,不能接收retChan中的结果, 因此当前child goroutine中的 retChan <- retItem 会一直卡住(不会执行 retChan <- retItem 之后的语句)
			// 如果<- limiter 放在 retChan <- retItem 之后执行, child goroutine卡主，不释放 semaphore; main goroutine无法新建 child goroutine, 此时整个程序卡住;
			// 如果<- limiter 放在 retChan <- retItem 之前执行, 当前child goroutine执行完耗时操作,马上释放持有的 semaphore, main goroutine将继续新建 child goroutine处理数据
			// 当然当前child goroutine 中 retChan <- retItem 大概率会被卡主,直到 main goroutine 跳出当前for循环
			<-semaphore
			retChan <- retItem
		}(srcItem)
	}

	go func() {
		wg.Wait() // 等待所有child goroutine处理完成后关闭retChan,此时main goroutine中接收结果的for循环将退出
		close(retChan)
	}()

	for retItem := range retChan {
		fmt.Printf("num:%d  squreRet=>%d\n", retItem.SrcNum, retItem.SqureRet)
	}

	fmt.Println("main goroutine exit...")
}
