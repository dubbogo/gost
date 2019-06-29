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
}
