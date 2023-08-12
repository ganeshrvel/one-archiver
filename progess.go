package onearchiver

import (
	"github.com/reactivex/rxgo/v2"
	"time"
)

func initProgress(totalFiles int, ph *ProgressHandler) (*ProgressInfo, *chan rxgo.Item) {
	pInfo := ProgressInfo{
		StartTime:          time.Now(),
		lastSentTime:       time.Now(),
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
	now := time.Now()
	timeDifference := now.Sub(pInfo.lastSentTime)

	// debounce time for the progress stream to avoid hogging up the cpu
	if timeDifference <= ProgressStreamDebounceTime {
		return
	}

	progressPercentage := Percent(float32(progressCount), float32(totalFiles))

	pInfo.TotalFiles = totalFiles
	pInfo.ProgressCount = progressCount
	pInfo.CurrentFilename = absolutePath
	pInfo.ProgressPercentage = progressPercentage
	pInfo.lastSentTime = time.Now()

	*ch <- rxgo.Of(pInfo)
}

func (pInfo *ProgressInfo) endProgress(ch *chan rxgo.Item, totalFiles int) {
	pInfo.TotalFiles = totalFiles
	pInfo.ProgressCount = totalFiles
	pInfo.CurrentFilename = ""
	pInfo.ProgressPercentage = 100.00
	pInfo.lastSentTime = time.Now()

	*ch <- rxgo.Of(pInfo)

	defer close(*ch)
}
