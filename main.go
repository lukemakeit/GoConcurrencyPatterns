package main

import "GoConcurrencyPatterns/DupDataProcessOnlyOnce"

func main() {
	//pipelineDemos.SerialPipeline()
	//pipelineDemos.FanInFanOut()
	//pipelineDemos.CanceledWithoutFinish()

	//BoundedParallelism.NotDealChildRet()
	//BoundedParallelism.NotDealChildRetBuffChan01()
	//BoundedParallelism.NotDealChildRetBuffChan02()

	//BoundedParallelism.DealChildRet()
	//BoundedParallelism.DealChildRetBuffChan01()
	//BoundedParallelism.DealChildRetBuffChan02()

	//BoundedParallelism.CanceledErrOccur()
	//BoundedParallelism.TimeOut01()
	//BoundedParallelism.TimeOutWithContext()

	//BoundedParallelism.ReturnByInputOrder01()
	//BoundedParallelism.ReturnByInputOrder02()
	//BoundedParallelism.PrintProcessEveryNSeconds()

	//DupDataProcessOnlyOnce.DupDataProcessOnlyOnceByLock()
	DupDataProcessOnlyOnce.DupDataProcessOnlyOnceBetter01()
	//DupDataProcessOnlyOnce.DupDataProcessOnlyOnceBetter02()
}
