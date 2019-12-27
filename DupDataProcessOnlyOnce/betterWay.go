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
// 优势: 该方法相比使用lock的方式,上锁到解锁时间很短,执行耗时仅为lock方式的1/4左右
func DupDataProcessOnlyOnceBetter01() {
	startTime := time.Now()

	srcData := []int{1, 2, 2, 33, 3, 3, 4, 4, 4, 4, 5, 6, 7, 7, 8, 8, 9, 10, 9, 11, 12, 8, 13, 11, 15, 17, 15, 2}

	retUniq := make(map[int]*int) //保证每个元素只执行一次求平方操作

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second) // 10秒后超时
	mu := &sync.Mutex{}
	limiter := 5 //控制并发度为5
	genChan := make(chan int)
	errChan := make(chan error)
	wg := &sync.WaitGroup{}
	finish := make(chan bool)

	//启动5个进程来执行求平方操作
	for worker := 0; worker < limiter; worker++ {
		// 第二stage:进行平方计算并保存
		wg.Add(1)
		go func() {
			defer wg.Done()
			for genItem := range genChan {
				mu.Lock() //上锁
				ret := retUniq[genItem]
				if ret == nil {
					// 如果是第一次获取元素平方值
					tmp := 0 // 先用一个空值占位
					retUniq[genItem] = &tmp
					mu.Unlock() //及时释放锁,完全不用等到获取平方值这个耗时操作完成再释放

					if genItem == 77 {
						// 求7平方值时发生错误
						err := fmt.Errorf("num:%d do squre,unknow error occur", genItem)
						select {
						case errChan <- err:
							break
						case <-ctx.Done():
							break
						}
						continue
					}
					time.Sleep(1 * time.Second)
					tmp = genItem * genItem
					// 如 genItem=4, 有A B C 3个child goroutine在尝试获取4的平方值,则只有A在实际执行求4的平方操作, B C卡在 <-ret.Ready, 而A获取4平方值后,close(ret.Ready),B C 继续执行
				} else {
					mu.Unlock() // 及时释放锁,比如当前srcItem=4,那么到这一步至少说明有一个child goroutine在获取4的平方值 或者 4的平方值已经保存到retUniq
				}
			}
		}()
	}
	go func() {
		// 第一stage:生产者
		defer close(genChan)

		for _, srcIem := range srcData {
			select {
			case genChan <- srcIem:
			case <-ctx.Done():
				return
			}
		}
	}()
	go func() {
		wg.Wait() //等待child goroutine执行完成
		close(finish)
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

	fmt.Printf("DupDataProcessOnlyOnceBetter01 take %.2f seconds \n", time.Now().Sub(startTime).Seconds())

	if err != nil {
		fmt.Printf("err detail:%s\n", err)
	} else {
		//统一打印结果
		for _, srcItem := range srcData {
			ret, _ := retUniq[srcItem]
			fmt.Printf("num:%d squreRet=>%d\n", srcItem, *ret)
		}
	}
}

// 描述: 计算列表中每个元素的平方值,列表中元素有重复,我们希望重复元素平方计算只执行一次。
// - 计算结果得到后实时打印;
// - 如果有错误,则直接返回退出;
// - 如果超时,则直接返回退出;
func DupDataProcessOnlyOnceBetter02() {
	startTime := time.Now()

	srcData := []int{1, 2, 2, 33, 3, 3, 4, 4, 4, 4, 5, 6, 7, 7, 8, 8, 9, 10, 9, 11, 12, 8, 13, 11, 15, 17, 15, 2}

	type squreEntry struct {
		SrcNum   int           // 原始数据
		SqureRet int           // 平方值结果
		Err      error         // 错误信息
		Ready    chan struct{} // 同步信号
	}

	retUniq := make(map[int]*squreEntry) //保证每个元素只执行一次求平方操作

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // 10秒后超时
	mu := &sync.Mutex{}
	limiter := 5 //控制并发度为5
	genChan := make(chan int)
	retChan := make(chan *squreEntry) //接收结果的channel
	wg := &sync.WaitGroup{}

	//启动5个进程来执行求平方操作
	for worker := 0; worker < limiter; worker++ {
		// 第二stage:进行平方计算
		wg.Add(1)
		go func() {
			defer wg.Done()
			for genItem := range genChan {
				mu.Lock() //上锁
				ret := retUniq[genItem]
				if ret == nil {
					//如果是第一次获取元素平方值
					ret = &squreEntry{
						SrcNum:   genItem,
						SqureRet: 0,
						Err:      nil,
						Ready:    make(chan struct{}),
					}
					retUniq[genItem] = ret
					mu.Unlock() //及时释放锁,完全不用等到获取平方值这个耗时操作完成再释放

					if genItem == 7 {
						// 求7平方值时发生错误
						ret.Err = fmt.Errorf("num:%d do squre,unknow error occur", genItem)
					} else {
						ret.SqureRet = genItem * genItem
						time.Sleep(1 * time.Second)
					}
					// 如 genItem=4, 有A B C 3个child goroutine在尝试获取4的平方值,则只有A在实际执行求4的平方操作, B C卡在 <-ret.Ready, 而A获取4平方值后,close(ret.Ready),B C 继续执行
					close(ret.Ready)
				} else {
					mu.Unlock() // 及时释放锁,比如当前srcItem=4,那么到这一步至少说明有一个child goroutine在获取4的平方值 或者 4的平方值已经保存到retUniq
					<-ret.Ready // 这里如果没卡住,那就代表4的平方值已获取到;如果卡住就代表有一个child goroutine正在获取4的平方值
				}

				// 及时发送结果数据
				select {
				case retChan <- ret:
				case <-ctx.Done():
					return
				}

			}
		}()
	}
	go func() {
		// 第一stage:生产者
		defer close(genChan)

		for _, srcIem := range srcData {
			select {
			case genChan <- srcIem:
			case <-ctx.Done():
				return
			}
		}
	}()
	go func() {
		wg.Wait() //等待child goroutine执行完成
		close(retChan)
	}()

	var err error
	var ok bool
	var retItem *squreEntry
	for {
		// 第三stage: 消费者
		select {
		case retItem,ok=<-retChan:
			if ok == false{
				//执行完成
				break
			}
			if retItem.Err != nil{
				// 发生错误
				ok = false
				err=retItem.Err
				break
			}

			fmt.Printf("num:%d squreRet:%d\n",retItem.SrcNum,retItem.SqureRet)
			case <- ctx.Done():
				// 超时
				err = fmt.Errorf("timeout ")
				ok=false
				break
		}
		if ok == false{
			break
		}
	}
	cancel()
	if err != nil {
		fmt.Printf("err detail:%s\n", err)
	}
	//time.Sleep(2 * time.Second)
	//fmt.Printf("after cancel numGoroutine:%d\n", runtime.NumGoroutine())

	fmt.Printf("DupDataProcessOnlyOnceBetter02 take %.2f seconds \n", time.Now().Sub(startTime).Seconds())
}
