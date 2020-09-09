package one_archiver

import (
	rxgo "github.com/ReactiveX/RxGo"
	"time"
)

func initProgress(totalFiles int, ph *ProgressHandler) (*ProgressInfo, *chan rxgo.Item) {
	pInfo := ProgressInfo{
		startTime:          time.Now(),
		totalFiles:         totalFiles,
		progressCount:      0,
		currentFilename:    "",
		progressPercentage: 0,
	}

	ch := make(chan rxgo.Item)

	observable := rxgo.FromChannel(ch)

	observable.ForEach(func(v interface{}) {
		ph.onReceived(&pInfo)
	}, func(err error) {
		ph.onError(err, &pInfo)
	}, func() {
		ph.onCompleted(&pInfo)
	})

	return &pInfo, &ch
}

func (pInfo *ProgressInfo) progress(ch *chan rxgo.Item, totalFiles int, absolutePath string, progressCount int) {
	progressPercentage := Percent(float32(progressCount), float32(totalFiles))

	pInfo.totalFiles = totalFiles
	pInfo.progressCount = progressCount
	pInfo.currentFilename = absolutePath
	pInfo.progressPercentage = progressPercentage

	*ch <- rxgo.Of(pInfo)
}

func (pInfo *ProgressInfo) endProgress(ch *chan rxgo.Item, totalFiles int) {
	pInfo.totalFiles = totalFiles
	pInfo.progressCount = totalFiles
	pInfo.currentFilename = ""
	pInfo.progressPercentage = 100.00

	*ch <- rxgo.Of(pInfo)

	defer close(*ch)
}
