package BoundedParallelism

import (
	"context"
	"fmt"
	"runtime"
	"time"
)

// 计算 num 的平方结果,并将结果写入 retChan
func doSqure(num int, ctx context.Context, semaphore chan bool, retChan chan *squreItem) {
	// 通过bufferChan控制并发,某一时刻有len(srcNums)个child goroutine运行中,
	// 同时执行的耗时操作的只有len(semaphore)个child goroutine,其余child goroutine会卡在下面这一步
	// 如果len(srcNums)很大,个人感觉这种方式也不是很好
	semaphore <- true
	defer func() {
		// 值传输完成后,释放资源,以便其他child goroutine能继续
		<-semaphore
	}()

	//fmt.Printf("num:%d doSqure start ....\n",num)

	var retItem *squreItem
	if num%10 == 0 {
		retItem = nil // 无需处理被10的倍数,则直接返回nil(不能直接return)
	} else {
		time.Sleep(1 * time.Second)

		retItem = &squreItem{
			num,
			num * num,
			nil,
		}
	}
	select {
	case retChan <- retItem:
	case <-ctx.Done():
		// 可能超时 或 发生了错误
		fmt.Printf("num:%d doSqure canceled\n", num)
		return
	}
}

// 描述:按照输入顺序去得到每个元素的结果
// 该方案缺陷如下:
// - 一开始就会启动大量child goroutine, 执行的耗时操作的只有len(semaphore)个child goroutine,其余child goroutine会卡在 semaphore <- true;
// - 该方案并非按照输入顺序计算每个元素的平方值,计算平方值操作是乱序的,可能 4*4计算完后,计算 13*13,再计算 12*12 ...
// - 该方案仅仅是在接收计算结果时,根据输入顺序接收计算结果;
// 因此如果num == 1 对应的 child goroutine如果最后才被执行,那么大量retChanList中结果最好才会被接收
func ReceiveRetByInputOrder01() {
	srcNums := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}

	// 为每个srcNum创建一个返回值接收channel,因此需提前知道srcNums的个数
	retChanList := make([]chan *squreItem, len(srcNums))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // 10秒后超时
	//控制并发度为4
	semaphore := make(chan bool, 4)
	for index, srcItem := range srcNums {
		retChanList[index] = make(chan *squreItem, 1)
		go doSqure(srcItem, ctx, semaphore, retChanList[index])
	}

	var ok bool
	for _, retChan := range retChanList {
		var retItem *squreItem
		select {
		case retItem = <-retChan:
			ok = true
			break
		case <-ctx.Done():
			ok = false // 已超时
			break
		}

		if !ok {
			break
		}
		if retItem == nil {
			// 不想处理的数据返回结果
			continue
		}
		if retItem.Err != nil {
			// 发生错误
			fmt.Printf("num:%d err:%s\n", retItem.SrcNum, retItem.Err)
			break
		}
		fmt.Printf("num:%d  squreRet=>%d\n", retItem.SrcNum, retItem.SqureRet)
	}
	cancel()

	fmt.Println("main goroutine exit...")
}

func doSqure02(num int, ctx context.Context, retChan chan *squreItem) {
	var retItem *squreItem
	if num%10 == 0 {
		retItem = nil // 无需处理被10的倍数,则直接返回nil(不能直接return)
	} else {
		time.Sleep(1 * time.Second)

		retItem = &squreItem{
			num,
			num * num,
			nil,
		}
	}
	select {
	case retChan <- retItem:
	case <-ctx.Done():
		// 可能超时 或 发生了错误
		fmt.Printf("num:%d doSqure canceled\n", num)
		return
	}
}

// 描述:按照输入顺序去计算每个元素的结果,接收每个元素的结果
// 该方案将输入元素 第一批 完成平方计算，并将结果获取 再进行下一批 元素平方计算,结果获取...以此类推
func ReceiveRetByInputOrder02() {
	srcNums := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}

	// 为每个srcNum创建一个返回值接收channel,因此需提前知道srcNums的个数
	retChanList := make([]chan *squreItem, len(srcNums))
	for index, _ := range srcNums {
		retChanList[index] = make(chan *squreItem)
	}

	type numItemWithIndex struct {
		Index   int
		NumItem int
	}
	genChan := make(chan *numItemWithIndex)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second) // 10秒后超时
	//控制并发度为4
	limit := 4
	for worker := 0; worker < limit; worker++ {
		// 第二stage:进行平方计算
		go func() {
			for genItem := range genChan {
				doSqure02(genItem.NumItem, ctx, retChanList[genItem.Index])
			}
		}()
	}

	go func() {
		// 第一stage:生产者
		defer close(genChan)

		for index, srcItem := range srcNums {
			temp := &numItemWithIndex{
				Index:   index, // index是必须传入的,index 在child goroutine中retChanList中找到传输结果值的channel
				NumItem: srcItem,
			}
			select {
			case genChan <- temp:
			case <-ctx.Done():
				return
			}
		}
	}()

	// 最后结果接收
	var ok bool
	var retItem *squreItem
	for _, retChan := range retChanList {
		// 第三stage: 消费者
		select {
		case retItem = <-retChan:
			ok = true
			break
		case <-ctx.Done():
			ok = false // 已超时
			break
		}

		if !ok {
			break
		}
		if retItem == nil {
			// 不想处理的数据返回结果
			continue
		}
		if retItem.Err != nil {
			// 发生错误
			fmt.Printf("num:%d err:%s\n", retItem.SrcNum, retItem.Err)
			break
		}
		fmt.Printf("num:%d  squreRet=>%d\n", retItem.SrcNum, retItem.SqureRet)
	}
	cancel()
	time.Sleep(2 * time.Second)
	fmt.Printf("after close cancel numGoroutine:%d\n", runtime.NumGoroutine())

	fmt.Println("main goroutine exit...")
}
