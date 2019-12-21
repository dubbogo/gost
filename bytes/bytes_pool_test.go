package gxbytes

import (
	"testing"
)

func Test_findIndex(t *testing.T) {
	bufPoolSize = []int{16, 4 << 10, 16 << 10, 32 << 10, 64 << 10}
	InitPool(bufPoolSize)

	type args struct {
		size int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{``, args{1}, 0},
		{``, args{15}, 0},
		{``, args{16}, 0},
		{``, args{17}, 1},
		{``, args{4095}, 1},
		{``, args{4096}, 1},
		{``, args{4097}, 2},
		{``, args{16 << 10}, 2},
		{``, args{64 << 10}, 4},
		{``, args{64 << 11}, 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := findIndex(tt.args.size); got != tt.want {
				t.Errorf("[%v] findIndex() = %v, want %v", tt.args.size, got, tt.want)
			}
		})
	}
}

func BenchmarkAcquireBytesSize8(b *testing.B)   { benchmarkAcquireBytes(b, 8) }
func BenchmarkAcquireBytesSize32(b *testing.B)  { benchmarkAcquireBytes(b, 32) }
func BenchmarkAcquireBytesSize10k(b *testing.B) { benchmarkAcquireBytes(b, 10000) }
func BenchmarkAcquireBytesSize60k(b *testing.B) { benchmarkAcquireBytes(b, 60000) }

func benchmarkAcquireBytes(b *testing.B, size int) {
	for i := 0; i < b.N; i++ {
		bytes, _ := AcquireBytes(size)
		ReleaseBytes(bytes)
	}
}
