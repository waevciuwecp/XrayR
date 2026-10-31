package mydispatcher

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/xtls/xray-core/common/buf"
	"github.com/xtls/xray-core/common/net"
	"github.com/xtls/xray-core/core"
)

func TestSniffFragmentedHeaders(t *testing.T) {
	request := "GET / HTTP/1.1\r\n" + strings.Repeat("X-Pad: "+strings.Repeat("a", 100)+"\r\n", 40) + "Host: example.com\r\n\r\n"
	reader := &cachedReader{reader: &buf.TimeoutWrapperReader{Reader: buf.NewReader(bytes.NewBufferString(request))}}
	ctx := context.WithValue(context.Background(), core.XrayKey(1), new(core.Instance))
	result, err := sniffer(ctx, reader, false, net.Network_TCP)
	if err != nil || result.Domain() != "example.com" {
		t.Fatalf("fragmented sniff: %v, %v", result, err)
	}
	var got bytes.Buffer
	for {
		mb, err := reader.ReadMultiBuffer()
		for _, b := range mb {
			if b != nil {
				got.Write(b.Bytes())
			}
		}
		buf.ReleaseMulti(mb)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
	}
	if got.String() != request {
		t.Fatal("sniffing lost or duplicated bytes")
	}
}

type failingTimeoutReader struct{}

func (failingTimeoutReader) ReadMultiBuffer() (buf.MultiBuffer, error) {
	return nil, io.ErrUnexpectedEOF
}
func (failingTimeoutReader) ReadMultiBufferTimeout(time.Duration) (buf.MultiBuffer, error) {
	return nil, io.ErrUnexpectedEOF
}
func TestCachedReaderError(t *testing.T) {
	reader := &cachedReader{reader: failingTimeoutReader{}}
	payload := buf.New()
	defer payload.Release()
	if err := reader.Cache(payload, time.Second); err != io.ErrUnexpectedEOF {
		t.Fatalf("read error lost: %v", err)
	}
}
