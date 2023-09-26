package onearchiver

import (
	"context"
)

// ProgressFunc defines callback functions for updating progress.
type ProgressFunc struct {
	OnReceived  func(*Progress) // Function to be called when a progress update is received.
	OnCompleted func(*Progress) // Function to be called when the operation is completed.
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

	// todo add ;logic for IsResumable     bool
	// todo add cancelMutex        sync.Mutex

	// TODO: Add more fields as needed.

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
func (session *Session) initializeProgress(totalFiles, totalSize int64) *Progress {
	session.progress = newProgress(totalFiles, totalSize)

	return session.progress
}

// enableCtxCancel enables the ability to interrupt progress with pause, resume, or stop.
func (session *Session) enableCtxCancel() {
	session.isCtxCancelEnabled = true
}

// disableCtxCancel disables the ability to interrupt progress.
func (session *Session) disableCtxCancel() {
	session.isCtxCancelEnabled = false
}

// isDone returns a channel that will be closed when the context is done.
func (session *Session) isDone() <-chan struct{} {
	return (*session.contextHandler.ctx).Done()
}

// ctxError checks for errors in the session's context.
func (session *Session) ctxError() error {
	return CtxError(session.contextHandler.ctx)
}

// canCancel checks if the session's context can be cancelled.
func (session *Session) canCancel() bool {
	return session.contextHandler.cancelFunc != nil && session.isCtxCancelEnabled
}

// cancel attempts to cancel the session's context.
func (session *Session) cancel() {
	if !session.canCancel() {
		return
	}

	(*session.contextHandler.cancelFunc)()
}

// fileProgress updates the progress of files count for the session.
func (session *Session) fileProgress(absolutePath string, filesProgressCount int64) {
	session.progress.fileProgress(absolutePath, filesProgressCount, session.ProgressFunc)
}

// sizeProgress updates the size (in bytes) progress for the session.
func (session *Session) sizeProgress(currentFileSize, currentSentFileSize int64) {
	session.progress.sizeProgress(currentFileSize, currentSentFileSize, session.ProgressFunc)
}

// endProgress finalizes the progress for the session.
func (session *Session) endProgress() {
	session.contextHandler.cancelFunc = nil

	session.disableCtxCancel()
	session.progress.endProgress(session.ProgressFunc)
}

// hasProgress checks if the session has progress initialized.
func (session *Session) hasProgress() bool {
	return session.progress != nil
}

// todo
// pause attempts to pause the session.
func (session *Session) pause() {
	//if !session.hasProgress() {
	//	return
	//}

	session.cancel()
}

// todo
// stop attempts to stop the session.
func (session *Session) stop() {
	//if !session.hasProgress() {
	//	return
	//}

	session.cancel()
}
