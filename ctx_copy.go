package onearchiver

import (
	"context"
	"errors"
	"fmt"
	"io"
)

func CtxError(ctx *context.Context) error {
	e := (*ctx).Err()
	if errors.Is(e, context.Canceled) {
		return fmt.Errorf(string(ErrorCancelledFileOperation))
	}

	return e
}

type CtxProgressFunc func(bytesTransferred int64)

// Custom reader to check for context cancellation and call the progress function
type ctxReaderFunc struct {
	r      io.Reader
	ctx    *context.Context
	onProg CtxProgressFunc
	soFar  int64
}

func (cr *ctxReaderFunc) Read(p []byte) (int, error) {
	select {
	case <-(*cr.ctx).Done():
		return 0, CtxError(cr.ctx)
	default:
		n, err := cr.r.Read(p)
		cr.soFar += int64(n)
		if cr.onProg != nil {
			cr.onProg(cr.soFar)
		}
		return n, err
	}
}

// CtxCopy copies from src to dst with context and progress callback
func CtxCopy(ctx *context.Context, dst io.Writer, src io.Reader, progress CtxProgressFunc) (int64, error) {
	return io.Copy(dst, &ctxReaderFunc{r: src, ctx: ctx, onProg: progress})
}
