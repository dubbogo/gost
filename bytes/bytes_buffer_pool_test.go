package gxbytes

import (
	"testing"
)

func TestBytesBufferPool(t *testing.T) {
	buf := GetBytesBuffer()
	bytes := []byte{0x00, 0x01, 0x02, 0x03, 0x04}
	buf.Write(bytes)
	if buf.Len() != len(bytes) {
		t.Error("iobuffer len not match write bytes' size")
	}
	PutBytesBuffer(buf)
	//buf2 := GetBytesBuffer()
	// https://go-review.googlesource.com/c/go/+/162919/
	// before go 1.13, sync.Pool just reserves some objs before every gc and will be cleanup by gc.
	// after Go 1.13, maybe there are many reserved objs after gc.
	//if buf != buf2 {
	//	t.Errorf("buf pointer %p != buf2 pointer %p", buf, buf2)
	//}
}
