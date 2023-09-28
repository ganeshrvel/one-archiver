package onearchiver

import (
	"context"
	"sync"
)

// ProgressFunc defines callback functions for updating progress.
type ProgressFunc struct {
	OnReceived func(*Progress) // Called when a progress update is received.
	OnEnded    func(*Progress) // Called upon operation termination. An end doesn't necessarily indicate successful completion. Always inspect ProgressStatus to determine the outcome.
}

// ContextHandler provides a structure to manage context for sessions.
type ContextHandler struct {
	ctx        *context.Context    // The context for the current session.
	cancelFunc *context.CancelFunc // Function to cancel the context.
}

// NewContextHandler initializes and returns a new ContextHandler.
func NewContextHandler() *ContextHandler {
	ctx, cancelFunc := context.WithCancel(context.Background())

	return &ContextHandler{
		ctx:        &ctx,
		cancelFunc: &cancelFunc,
	}
}

// Session represents a transfer or operation session with progress and context management.
type Session struct {
	progress           *Progress       // Current progress of the session.
	ProgressFunc       *ProgressFunc   // Functions to update progress.
	id                 string          // Unique identifier for the session.
	contextHandler     *ContextHandler // Context management for the session.
	isCtxCancelEnabled bool            // Flag to check if context cancellation is enabled.
	cancelMutex        sync.Mutex      // Mutex for protecting concurrent access to isCtxCancelEnabled field.
	cancelFuncMutex    sync.Mutex      // Mutex for protecting concurrent access to the cancelFunc in contextHandler.

	// todo add ;logic for IsResumable     bool
}

// newSession initializes a new session with the given ID
func newSession(sessionId string, progressFunc *ProgressFunc) *Session {
	return &Session{
		progress:           nil,
		id:                 sessionId,
		ProgressFunc:       progressFunc,
		contextHandler:     NewContextHandler(),
		isCtxCancelEnabled: false,
	}
}

// initializeProgress sets up the progress for the session
func (session *Session) initializeProgress(totalFiles, totalSize int64, canResumeTransfer bool) *Progress {
	session.progress = newProgress(totalFiles, totalSize)
	session.progress.CanResumeTransfer = canResumeTransfer

	return session.progress
}

// fileProgress updates the progress of files count for the session.
func (session *Session) fileProgress(absolutePath string, filesProgressCount int64) {
	session.progress.fileProgress(absolutePath, filesProgressCount, session.ProgressFunc)
}

// symlinkSizeProgress updates the size (in bytes) progress, of symlink, for the session.
func (session *Session) symlinkSizeProgress(originalTargetPath, targetPathToWrite string) {
	targetPathSizeToWrite := int64(len(targetPathToWrite))
	correctionSize := targetPathSizeToWrite - int64(len(originalTargetPath))

	// Symlinks within an archive can undergo modifications to fit system-specific symlink creation criteria.
	// For example, the target OS might sanitize a symlink by adjusting slashes or appending additional characters.
	// We need to account for these modifications in our size calculations.
	// To do this, we determine the difference between the original archived symlink size and the processed symlink size.
	// We then adjust the total file size by this difference.
	session.totalSizeCorrection(correctionSize)
	session.sizeProgress(targetPathSizeToWrite, targetPathSizeToWrite, targetPathSizeToWrite)
}

// sizeProgress updates the size (in bytes) progress for the session.
func (session *Session) sizeProgress(currentFileSize, soFarTransferredSize, lastTransferredSize int64) {
	session.progress.sizeProgress(currentFileSize, soFarTransferredSize, lastTransferredSize, session.ProgressFunc)
}

// revertSizeProgress subtracts the specified size (in bytes) from the progress for the session.
func (session *Session) revertSizeProgress(size int64) {
	session.progress.revertSizeProgress(size, session.ProgressFunc)
}

// totalSizeCorrection adjusts the total size of the session's progress based on the provided size.
// This is particularly useful in scenarios where the actual file size (e.g., after symlink resolution)
// may differ from the initially computed size. The provided size can be positive or negative,
// indicating an increase or decrease in the total size, respectively.
func (session *Session) totalSizeCorrection(size int64) {
	session.progress.totalSizeCorrection(size, session.ProgressFunc)
}

// endProgress completes the progress for the session, it marks the transfer as ProgressStatusCompleted.
func (session *Session) endProgress(status ProgressStatus) {
	session.setCancelFunc(nil)

	session.disableCtxCancel()
	session.progress.endProgress(session.ProgressFunc, status)
}

// todo
// Pause the session.
func (session *Session) Pause() {
	session.cancel()
}

// todo
// Stop the session.
func (session *Session) Stop() {
	session.cancel()
}

// enableCtxCancel enables the ability to interrupt progress.
func (session *Session) enableCtxCancel() {
	session.setIsCtxCancelEnabled(true)
}

// disableCtxCancel disables the ability to interrupt progress.
func (session *Session) disableCtxCancel() {
	session.setIsCtxCancelEnabled(false)
}

// getIsCtxCancelEnabled safely retrieves the value of isCtxCancelEnabled using a mutex lock.
// This ensures that the field is accessed in a thread-safe manner.
func (session *Session) getIsCtxCancelEnabled() bool {
	session.cancelMutex.Lock()
	defer session.cancelMutex.Unlock()

	return session.isCtxCancelEnabled
}

// setIsCtxCancelEnabled safely sets the value of isCtxCancelEnabled using a mutex lock.
// This ensures that the field is updated in a thread-safe manner.
func (session *Session) setIsCtxCancelEnabled(val bool) {
	session.cancelMutex.Lock()
	defer session.cancelMutex.Unlock()

	session.isCtxCancelEnabled = val
}

// getCancelFunc safely retrieves the cancelFunc from contextHandler using a mutex lock.
// This ensures that the field is accessed in a thread-safe manner.
func (session *Session) getCancelFunc() *context.CancelFunc {
	session.cancelFuncMutex.Lock()
	defer session.cancelFuncMutex.Unlock()

	return session.contextHandler.cancelFunc
}

// setCancelFunc safely sets the cancelFunc in contextHandler using a mutex lock.
// This ensures that the field is updated in a thread-safe manner.
func (session *Session) setCancelFunc(cf *context.CancelFunc) {
	session.cancelFuncMutex.Lock()
	defer session.cancelFuncMutex.Unlock()

	session.contextHandler.cancelFunc = cf
}

// canCancel checks if the session's context can be cancelled.
func (session *Session) canCancel() bool {
	return session.getCancelFunc() != nil && session.getIsCtxCancelEnabled()
}

// cancel attempts to cancel the session's context.
func (session *Session) cancel() {
	if !session.canCancel() {
		return
	}

	(*session.getCancelFunc())()
}

// isDone returns a channel that will be closed when the context is done.
func (session *Session) isDone() <-chan struct{} {
	return (*session.contextHandler.ctx).Done()
}

// ctxError checks for errors in the session's context.
func (session *Session) ctxError() error {
	return CtxError(session.contextHandler.ctx)
}
