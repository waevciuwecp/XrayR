package limiter

import (
	"context"
	"io"

	"github.com/xtls/xray-core/common"
	"github.com/xtls/xray-core/common/buf"
	"golang.org/x/time/rate"
)

type Writer struct {
	writer  buf.Writer
	limiter *rate.Limiter
	ctx     context.Context
}

func (l *Limiter) RateWriter(ctx context.Context, writer buf.Writer, limiter *rate.Limiter) buf.Writer {
	return &Writer{
		ctx:     ctx,
		writer:  writer,
		limiter: limiter,
	}
}

func (w *Writer) Close() error {
	return common.Close(w.writer)
}

func (w *Writer) WriteMultiBuffer(mb buf.MultiBuffer) error {
	if err := waitBytes(w.ctx, w.limiter, int(mb.Len())); err != nil {
		buf.ReleaseMulti(mb)
		return err
	}
	return w.writer.WriteMultiBuffer(mb)
}

// Reader applies the same shared bucket as Writer to externally supplied links.
type Reader struct {
	reader  buf.Reader
	limiter *rate.Limiter
	ctx     context.Context
}

func (l *Limiter) RateReader(ctx context.Context, reader buf.Reader, bucket *rate.Limiter) buf.Reader {
	return &Reader{reader: reader, limiter: bucket, ctx: ctx}
}

func waitBytes(ctx context.Context, bucket *rate.Limiter, size int) error {
	for size > 0 {
		chunk := min(size, bucket.Burst())
		if chunk <= 0 {
			return io.ErrNoProgress
		}
		if err := bucket.WaitN(ctx, chunk); err != nil {
			return err
		}
		size -= chunk
	}
	return nil
}

func (r *Reader) ReadMultiBuffer() (buf.MultiBuffer, error) {
	mb, err := r.reader.ReadMultiBuffer()
	if waitErr := waitBytes(r.ctx, r.limiter, int(mb.Len())); waitErr != nil {
		buf.ReleaseMulti(mb)
		return nil, waitErr
	}
	return mb, err
}
func (r *Reader) Interrupt()   { common.Interrupt(r.reader) }
func (r *Reader) Close() error { return common.Close(r.reader) }
func (w *Writer) Interrupt()   { common.Interrupt(w.writer) }
