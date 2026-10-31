package limiter

import (
	"bytes"
	"context"
	"github.com/xtls/xray-core/common/buf"
	"golang.org/x/time/rate"
	"testing"
	"time"
)

func TestRateReader(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	reader := New().RateReader(ctx, buf.NewReader(bytes.NewBufferString("payload")), rate.NewLimiter(1, 2))
	mb, err := reader.ReadMultiBuffer()
	if err == nil || !mb.IsEmpty() {
		buf.ReleaseMulti(mb)
		t.Fatal("cancelled reader delivered data")
	}
	ctx, cancel = context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	if err := waitBytes(ctx, rate.NewLimiter(1, 2), 10); err == nil {
		t.Fatal("oversized batch bypassed rate limit")
	}
	unlimited := New().RateReader(context.Background(), buf.NewReader(bytes.NewBufferString("ok")), rate.NewLimiter(rate.Inf, 1))
	mb, err = unlimited.ReadMultiBuffer()
	defer buf.ReleaseMulti(mb)
	if err != nil || mb.Len() != 2 {
		t.Fatalf("unlimited reader: %v", err)
	}
}
