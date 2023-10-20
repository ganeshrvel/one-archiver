package onearchiver

import (
	"time"
)

type ProgressStatus string

const (
	ProgressStatusStarting  ProgressStatus = "Starting"
	ProgressStatusRunning   ProgressStatus = "Running"
	ProgressStatusCancelled ProgressStatus = "Cancelled"
	ProgressStatusCompleted ProgressStatus = "Completed"
)

type ProgressCancelReason string

const (
	ProgressCancelReasonNone    ProgressCancelReason = "None"
	ProgressCancelReasonStopped ProgressCancelReason = "Stopped"
	ProgressCancelReasonPaused  ProgressCancelReason = "Paused"
)

type Progress struct {
	TotalFiles                        int64   // Total number of files to be transferred.
	SentFilesCount                    int64   // Number of files that have been transferred.
	SentFilesCountPercentage          float64 // Percentage of files that have been transferred.
	CurrentFilepath                   string  // Path of the current file being transferred.
	TotalSize                         int64   // Total byte size of all files.
	SentSize                          int64   // Total byte size that has been transferred.
	SentSizeProgressPercentage        float64 // Percentage of the total byte size that has been transferred.
	CurrentFileSize                   int64   // Size of the current file being transferred.
	CurrentFileSentSize               int64   // Amount of the current file that has been transferred.
	CurrentFileProgressSizePercentage float64 // Percentage of the current file that has been transferred.

	ProgressStatus       ProgressStatus       // Progress status
	ProgressCancelReason ProgressCancelReason // Reason why progress was cancelled
	CanResumeTransfer    bool                 // Indicates whether the session be resumed

	StartTime                  time.Time // Start time of the transfer.
	LatestSentTime             time.Time // Time when the latest file transfer update was sent.
	ProgressStreamDebounceTime int64     // Rate limit the progress result streaming
}

// newProgress initializes and returns a new Progress object
func newProgress(totalFiles, totalSize, progressStreamDebounceTime int64) *Progress {
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

		CanResumeTransfer:    false,
		ProgressStatus:       ProgressStatusStarting,
		ProgressCancelReason: ProgressCancelReasonNone,

		StartTime:      time.Now(),
		LatestSentTime: time.Time{},

		ProgressStreamDebounceTime: progressStreamDebounceTime,
	}
}

// fileProgressStart initializes the progress data for the start of a new file transfer.
func (progress *Progress) fileProgressStart(absolutePath string, progressFunc *ProgressFunc) {
	// It sets the absolute path of the file being processed and resets the relevant counters. It may emit a progress event.
	progress.CurrentFilepath = absolutePath
	progress.CurrentFileSize = 0
	progress.CurrentFileSentSize = 0
	progress.CurrentFileProgressSizePercentage = 0
	progress.setStatus(ProgressStatusRunning)

	progress.rateLimitProgress(progressFunc)
}

// fileProgressEnd finalizes the progress data once a file transfer is complete.
func (progress *Progress) fileProgressEnd(filesProgressCount int64, progressFunc *ProgressFunc) {
	totalFiles := progress.TotalFiles

	// It updates the number of files processed so far and computes the overall progress percentage. It may emit a progress event.
	progressPercentage := TransferRatePercent(float64(filesProgressCount), float64(totalFiles))
	progress.SentFilesCount = filesProgressCount
	progress.SentFilesCountPercentage = progressPercentage

	progress.rateLimitProgress(progressFunc)
}

// sizeProgress updates the progress data based on the size of files transferred.
func (progress *Progress) sizeProgress(currentFileSize, soFarTransferredSize, lastTransferredSize int64, progressFunc *ProgressFunc) {
	// It calculates the percentage of the overall transfer and for the current file in process. It may emit a progress event.
	progress.SentSize += lastTransferredSize
	sentSizeProgressPercentage := TransferRatePercent(float64(progress.SentSize), float64(progress.TotalSize))
	progress.SentSizeProgressPercentage = sentSizeProgressPercentage

	progress.CurrentFileSize = currentFileSize
	progress.CurrentFileSentSize = soFarTransferredSize
	currentFileProgressSizePercentage := TransferRatePercent(float64(progress.CurrentFileSentSize), float64(currentFileSize))
	progress.CurrentFileProgressSizePercentage = currentFileProgressSizePercentage
	progress.setStatus(ProgressStatusRunning)

	progress.rateLimitProgress(progressFunc)
}

// rateLimitProgress ensures that progress updates are emitted at a rate-limited frequency.
// This helps in avoiding potential CPU overuse. Progress updates are emitted immediately if the
// progress reaches 100% completion.
func (progress *Progress) rateLimitProgress(progressFunc *ProgressFunc) {
	if progress.ProgressStreamDebounceTime > 0 {
		now := time.Now()
		timeDifference := now.Sub(progress.LatestSentTime)
		// debounce time for the progress stream to avoid hogging up the cpu
		// if the progressPercentage is 100% then emit an event
		if int64(timeDifference) <= progress.ProgressStreamDebounceTime {
			return
		}
	}

	progress.LatestSentTime = time.Now()
	progressFunc.OnReceived(progress)
}

// revertSizeProgress subtracts the specified size (in bytes) from the progress for the session.
func (progress *Progress) revertSizeProgress(size int64, progressFunc *ProgressFunc) {
	// Decrement the SentSize by the given size.
	progress.SentSize -= size
	sentSizeProgressPercentage := TransferRatePercent(float64(progress.SentSize), float64(progress.TotalSize))
	progress.SentSizeProgressPercentage = sentSizeProgressPercentage

	// Decrement the current file's sent size by the given size.
	progress.CurrentFileSentSize -= size
	progress.CurrentFileProgressSizePercentage = 0
	progress.setStatus(ProgressStatusRunning)

	progress.LatestSentTime = time.Now()
	progressFunc.OnReceived(progress)
}

// totalSizeCorrection adjusts the total size of the progress based on the provided size.
// This is particularly useful in scenarios where the actual file size (e.g., after symlink resolution)
// may differ from the initially computed size. The provided size can be positive or negative,
// indicating an increase or decrease in the total size, respectively.
func (progress *Progress) totalSizeCorrection(size int64, progressFunc *ProgressFunc) {
	progress.TotalSize += size
	progress.setStatus(ProgressStatusRunning)
	progressFunc.OnReceived(progress)
}

// setStatus sets progress status
func (progress *Progress) setStatus(status ProgressStatus) {
	progress.ProgressStatus = status
}

// setStatus sets progress status
func (progress *Progress) setCancelReason(reason ProgressCancelReason) {
	progress.ProgressCancelReason = reason
}

// endProgress signals the end of the file transfer process by invoking the OnEnded function
// of the provided progressFunc with the final progress state.
func (progress *Progress) endProgress(progressFunc *ProgressFunc, status ProgressStatus) {
	progress.LatestSentTime = time.Now()
	progress.setStatus(status)

	progressFunc.OnEnded(progress)
}
