package gxbytes

import (
	"fmt"
	"testing"
)

import (
	"github.com/stretchr/testify/assert"
)

func TestAcquireReleaseBytes(t *testing.T) {
	testBytesPool := NewBytesPool([]int{512, 1 << 10, 4 << 10, 16 << 10, 64 << 10})

	// ---- acquire 1st
	b := testBytesPool.AcquireBytes(64)
	assert.Equal(t, 0, len(b))
	assert.Equal(t, 512, cap(b))
	p := fmt.Sprintf("%p", b)
	t.Logf("1st addr:%p, val: %s", b, b)
	b = append(b, []byte("hello")...)
	assert.Equal(t, 5, len(b))
	t.Logf("1st addr:%p, val: %s", b, b)
	testBytesPool.ReleaseBytes(b)

	// ---- acquire 2nd
	b = testBytesPool.AcquireBytes(64)
	assert.Equal(t, 0, len(b))
	assert.Equal(t, 512, cap(b))
	assert.Equal(t, p, fmt.Sprintf("%p", b))
	assert.Equal(t, "hello", string(b[0:5]))
	t.Logf("2nd addr:%p, val: %s", b, b)
	testBytesPool.ReleaseBytes(b)

	// ---- acquire 3rd
	b = testBytesPool.AcquireBytes(512)
	assert.Equal(t, 0, len(b))
	assert.Equal(t, 512, cap(b))
	t.Logf("3rd addr:%p, val: %s", b, b)
	assert.Equal(t, p, fmt.Sprintf("%p", b))
	assert.Equal(t, "hello", string(b[0:5]))

	b = b[0:512]
	assert.Equal(t, 512, len(b))
	// bytes len exceeding cap causes reallocation
	b = append(b, []byte("hello")...)
	assert.NotEqual(t, 512, len(b))
	// address changed
	assert.NotEqual(t, p, fmt.Sprintf("%p", b))
	t.Logf("3rd addr:%p, val: %s", b, b)
	p = fmt.Sprintf("%p", b)

	// non-equal cap bytes will not be accepted by pool
	testBytesPool.ReleaseBytes(b)

	// ---- acquire 4th
	b = testBytesPool.AcquireBytes(512)
	// address changed for getting a new bytes
	assert.NotEqual(t, p, fmt.Sprintf("%p", b))
	t.Logf("4th addr:%p, val: %s", b, b)
	testBytesPool.ReleaseBytes(b)
}

func Test_findIndex(t *testing.T) {
	bp := NewBytesPool([]int{16, 4 << 10, 16 << 10, 32 << 10, 64 << 10})

	type args struct {
		size int
	}
	tests := []struct {
		args args
		want int
	}{
		{args{1}, 0},
		{args{15}, 0},
		{args{16}, 0},
		{args{17}, 1},
		{args{4095}, 1},
		{args{4096}, 1},
		{args{4097}, 2},
		{args{16 << 10}, 2},
		{args{64 << 10}, 4},
		{args{64 << 11}, 5},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprint(tt.args.size), func(t *testing.T) {
			if got := bp.findIndex(tt.args.size); got != tt.want {
				t.Errorf("[%v] findIndex() = %v, want %v", tt.args.size, got, tt.want)
			}
		})
	}
}

func BenchmarkAcquireBytesSize32(b *testing.B)  { benchmarkAcquireBytes(b, 32) }
func BenchmarkAcquireBytesSize10k(b *testing.B) { benchmarkAcquireBytes(b, 10000) }
func BenchmarkAcquireBytesSize60k(b *testing.B) { benchmarkAcquireBytes(b, 60000) }
func BenchmarkAcquireBytesSize70k(b *testing.B) { benchmarkAcquireBytes(b, 70000) }

func benchmarkAcquireBytes(b *testing.B, size int) {
	for i := 0; i < b.N; i++ {
		bytes := AcquireBytes(size)
		ReleaseBytes(bytes)
	}
}

func BenchmarkFindIndexSize8(b *testing.B)   { benchmarkfindIndex(b, 8) }
func BenchmarkFindIndexSize60k(b *testing.B) { benchmarkfindIndex(b, 60000) }

func benchmarkfindIndex(b *testing.B, size int) {
	for i := 0; i < b.N; i++ {
		defaultBytesPool.findIndex(size)
	}
}
