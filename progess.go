package onearchiver

import (
	"time"
)

type Progress struct {
	TotalFiles                        int64     // Total number of files to be transferred.
	SentFilesCount                    int64     // Number of files that have been transferred.
	SentFilesCountPercentage          float32   // Percentage of files that have been transferred.
	CurrentFilepath                   string    // Path of the current file being transferred.
	TotalSize                         int64     // Total byte size of all files.
	SentSize                          int64     // Total byte size that has been transferred.
	SentSizeProgressPercentage        float32   // Percentage of the total byte size that has been transferred.
	CurrentFileSize                   int64     // Size of the current file being transferred.
	CurrentFileSentSize               int64     // Amount of the current file that has been transferred.
	CurrentFileProgressSizePercentage float32   // Percentage of the current file that has been transferred.
	StartTime                         time.Time // Start time of the transfer.
	lastSentTime                      time.Time // Time when the last file transfer update was sent.
}

// newProgress initializes and returns a new Progress object
func newProgress(totalFiles, totalSize int64) *Progress {
	return &Progress{
		TotalFiles:               totalFiles,
		SentFilesCount:           0,
		SentFilesCountPercentage: 0,

		TotalSize:                  totalSize,
		SentSize:                   0,
		SentSizeProgressPercentage: 0,

		CurrentFilepath:                   "",
		CurrentFileSize:                   0,
		CurrentFileSentSize:               0,
		CurrentFileProgressSizePercentage: 0,

		StartTime:    time.Now(),
		lastSentTime: time.Time{},
	}
}

// fileProgress updates the file progress with the current state, such as the absolute path of the file being processed,
// and the number of files processed so far. It emits a progress event if necessary.
func (progress *Progress) fileProgress(absolutePath string, filesProgressCount int64, progressFunc *ProgressFunc) {
	totalFiles := progress.TotalFiles

	progressPercentage := Percent(float32(filesProgressCount), float32(totalFiles))
	progress.SentFilesCount = filesProgressCount
	progress.SentFilesCountPercentage = progressPercentage

	progress.CurrentFilepath = absolutePath
	progress.CurrentFileSize = 0
	progress.CurrentFileSentSize = 0
	progress.CurrentFileProgressSizePercentage = 0

	now := time.Now()
	timeDifference := now.Sub(progress.lastSentTime)
	// debounce time for the progress stream to avoid hogging up the cpu
	// if the progressPercentage is 100% then emit an event
	if timeDifference <= ProgressStreamDebounceTime && progressPercentage < 100 {
		return
	}

	progress.lastSentTime = time.Now()
	progressFunc.OnReceived(progress)
}

// sizeProgress updates the progress based on the size of files transferred. It calculates the
// percentage of the overall transfer as well as the percentage for the current file.
// It emits a progress event if necessary.
func (progress *Progress) sizeProgress(currentFileSize, currentSentFileSize int64, progressFunc *ProgressFunc) {

	progress.SentSize += currentSentFileSize
	sentSizeProgressPercentage := Percent(float32(progress.SentSize), float32(progress.TotalSize))
	progress.SentSizeProgressPercentage = sentSizeProgressPercentage

	progress.CurrentFileSize = currentFileSize
	progress.CurrentFileSentSize += currentSentFileSize
	currentFileProgressSizePercentage := Percent(float32(progress.CurrentFileSentSize), float32(currentFileSize))
	progress.CurrentFileProgressSizePercentage = currentFileProgressSizePercentage

	now := time.Now()
	timeDifference := now.Sub(progress.lastSentTime)
	// debounce time for the progress stream to avoid hogging up the cpu
	// if the progressPercentage is 100% then emit an event
	if timeDifference <= ProgressStreamDebounceTime && currentFileProgressSizePercentage < 100 {
		return
	}

	progress.lastSentTime = time.Now()
	progressFunc.OnReceived(progress)
}

// endProgress signals the end of the file transfer process by invoking the OnCompleted function
// of the provided progressFunc with the final progress state.
func (progress *Progress) endProgress(progressFunc *ProgressFunc) {
	progress.lastSentTime = time.Now()

	progressFunc.OnCompleted(progress)
}
