package onearchiver

import (
	rxgo "github.com/ReactiveX/RxGo"
	"time"
)

func initProgress(totalFiles int, ph *ProgressHandler) (*ProgressInfo, *chan rxgo.Item) {
	pInfo := ProgressInfo{
		StartTime:          time.Now(),
		TotalFiles:         totalFiles,
		ProgressCount:      0,
		CurrentFilename:    "",
		ProgressPercentage: 0,
	}

	ch := make(chan rxgo.Item)

	observable := rxgo.FromChannel(ch)

	observable.ForEach(func(v interface{}) {
		ph.OnReceived(&pInfo)
	}, func(err error) {
		ph.OnError(err, &pInfo)
	}, func() {
		ph.OnCompleted(&pInfo)
	})

	return &pInfo, &ch
}

func (pInfo *ProgressInfo) progress(ch *chan rxgo.Item, totalFiles int, absolutePath string, progressCount int) {
	progressPercentage := Percent(float32(progressCount), float32(totalFiles))

	pInfo.TotalFiles = totalFiles
	pInfo.ProgressCount = progressCount
	pInfo.CurrentFilename = absolutePath
	pInfo.ProgressPercentage = progressPercentage

	*ch <- rxgo.Of(pInfo)
}

func (pInfo *ProgressInfo) endProgress(ch *chan rxgo.Item, totalFiles int) {
	pInfo.TotalFiles = totalFiles
	pInfo.ProgressCount = totalFiles
	pInfo.CurrentFilename = ""
	pInfo.ProgressPercentage = 100.00

	*ch <- rxgo.Of(pInfo)

	defer close(*ch)
}
