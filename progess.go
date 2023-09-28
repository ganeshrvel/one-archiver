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

type Progress struct {
	TotalFiles                        int64          // Total number of files to be transferred.
	SentFilesCount                    int64          // Number of files that have been transferred.
	SentFilesCountPercentage          float64        // Percentage of files that have been transferred.
	CurrentFilepath                   string         // Path of the current file being transferred.
	TotalSize                         int64          // Total byte size of all files.
	SentSize                          int64          // Total byte size that has been transferred.
	SentSizeProgressPercentage        float64        // Percentage of the total byte size that has been transferred.
	CurrentFileSize                   int64          // Size of the current file being transferred.
	CurrentFileSentSize               int64          // Amount of the current file that has been transferred.
	CurrentFileProgressSizePercentage float64        // Percentage of the current file that has been transferred.
	StartTime                         time.Time      // Start time of the transfer.
	ProgressStatus                    ProgressStatus // Progress status
	CanResumeTransfer                 bool           // Indicates whether the session be resumed
	lastSentTime                      time.Time      // Time when the last file transfer update was sent.
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

		CanResumeTransfer: true,
		ProgressStatus:    ProgressStatusStarting,

		StartTime:    time.Now(),
		lastSentTime: time.Time{},
	}
}

// fileProgress updates the file progress with the current state, such as the absolute path of the file being processed,
// and the number of files processed so far. It emits a progress event if necessary.
func (progress *Progress) fileProgress(absolutePath string, filesProgressCount int64, progressFunc *ProgressFunc) {
	totalFiles := progress.TotalFiles

	progressPercentage := Percent(float64(filesProgressCount), float64(totalFiles))
	progress.SentFilesCount = filesProgressCount
	progress.SentFilesCountPercentage = progressPercentage

	progress.CurrentFilepath = absolutePath
	progress.CurrentFileSize = 0
	progress.CurrentFileSentSize = 0
	progress.CurrentFileProgressSizePercentage = 0
	progress.setStatus(ProgressStatusRunning)

	now := time.Now()
	timeDifference := now.Sub(progress.lastSentTime)
	// debounce time for the progress stream to avoid hogging up the cpu
	// if the progressPercentage is 100% then emit an event
	if timeDifference <= ProgressStreamDebounceTime {
		return
	}

	progress.lastSentTime = time.Now()
	progressFunc.OnReceived(progress)
}

// sizeProgress updates the progress based on the size of files transferred. It calculates the
// percentage of the overall transfer as well as the percentage for the current file.
// It emits a progress event if necessary.
func (progress *Progress) sizeProgress(currentFileSize, soFarTransferredSize, lastTransferredSize int64, progressFunc *ProgressFunc) {

	progress.SentSize += lastTransferredSize
	sentSizeProgressPercentage := Percent(float64(progress.SentSize), float64(progress.TotalSize))
	progress.SentSizeProgressPercentage = sentSizeProgressPercentage

	progress.CurrentFileSize = currentFileSize
	progress.CurrentFileSentSize = soFarTransferredSize
	currentFileProgressSizePercentage := Percent(float64(progress.CurrentFileSentSize), float64(currentFileSize))
	progress.CurrentFileProgressSizePercentage = currentFileProgressSizePercentage
	progress.setStatus(ProgressStatusRunning)

	now := time.Now()
	timeDifference := now.Sub(progress.lastSentTime)
	// debounce time for the progress stream to avoid hogging up the cpu
	// if the progressPercentage is 100% then emit an event
	if timeDifference <= ProgressStreamDebounceTime {
		return
	}

	progress.lastSentTime = time.Now()
	progressFunc.OnReceived(progress)
}

// revertSizeProgress subtracts the specified size (in bytes) from the progress for the session.
func (progress *Progress) revertSizeProgress(size int64, progressFunc *ProgressFunc) {
	// Decrement the SentSize by the given size.
	progress.SentSize -= size
	sentSizeProgressPercentage := Percent(float64(progress.SentSize), float64(progress.TotalSize))
	progress.SentSizeProgressPercentage = sentSizeProgressPercentage

	// Decrement the current file's sent size by the given size.
	progress.CurrentFileSentSize -= size
	progress.CurrentFileProgressSizePercentage = 0
	progress.setStatus(ProgressStatusRunning)

	progress.lastSentTime = time.Now()
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

// endProgress signals the end of the file transfer process by invoking the OnEnded function
// of the provided progressFunc with the final progress state.
func (progress *Progress) endProgress(progressFunc *ProgressFunc, status ProgressStatus) {
	progress.lastSentTime = time.Now()
	progress.setStatus(status)

	progressFunc.OnEnded(progress)
}
