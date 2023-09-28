package onearchiver

import (
	"context"
	"errors"
	"fmt"
	"io"
)

// CtxError returns a formatted error if the context is cancelled; otherwise, it returns the original context error.
func CtxError(ctx *context.Context) error {
	e := (*ctx).Err()
	if errors.Is(e, context.Canceled) {
		return fmt.Errorf(string(ErrorCancelledFileOperation))
	}

	return e
}

type CtxProgressFunc func(soFarTransferredSize int64, lastTransferredSize int64)

// ctxProgressReader is a custom reader that wraps another io.Reader. It checks for context cancellation and invokes
// a progress callback function whenever data is read.
type ctxProgressReader struct {
	r                    io.Reader        // The underlying reader.
	ctx                  *context.Context // The context.
	onProg               CtxProgressFunc  // Progress callback function.
	soFarTransferredSize int64            // Total number of bytes transferred so far.
	lastTransferredSize  int64            // Number of bytes transferred in the last read.
}

// Read reads data from the underlying io.Reader and invokes the progress callback with the updated transfer statistics.
func (cr *ctxProgressReader) Read(p []byte) (int, error) {
	select {
	case <-(*cr.ctx).Done():
		// If the context is done (cancelled or deadline exceeded), return the context error.
		return 0, CtxError(cr.ctx)
	default:
		read, err := cr.r.Read(p)
		cr.lastTransferredSize = int64(read)
		cr.soFarTransferredSize += int64(read)
		if cr.onProg != nil {
			cr.onProg(cr.soFarTransferredSize, cr.lastTransferredSize)
		}
		return read, err
	}
}

// CtxCopy copies data from the src reader to the dst writer.
func CtxCopy(ctx *context.Context, dst io.Writer, src io.Reader, isDir bool, progress CtxProgressFunc) (writtenBytes int64, err error) {
	if isDir {
		_, err := io.Copy(dst, src)

		return 0, err
	}

	return io.Copy(dst, &ctxProgressReader{r: src, ctx: ctx, onProg: progress})
}
