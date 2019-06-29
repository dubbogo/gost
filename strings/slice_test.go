package gxstrings

import (
	"reflect"
	"testing"
)

// go test  -v slice_test.go  slice.go

func TestString(t *testing.T) {
	b := []byte("hello world")
	// After converting slice to string, the string value will change
	// when the slice got new value.
	a := String(b)
	b[0] = 'a'
	if !reflect.DeepEqual("aello world", a) {
		t.Errorf("a:%+v != `aello world`", a)
	}
}

// BenchmarkString0-8   	2000000000	         0.27 ns/op
func BenchmarkString0(b *testing.B) {
	b.StartTimer()
	for i := 0; i < 100000000; i++ {
		bs := []byte("hello world")
		_ = string(bs)
		bs[0] = 'a'
	}
	b.StopTimer()
	bs := []byte("hello world")
	a := string(bs)
	bs[0] = 'a'
	if reflect.DeepEqual("aello world", a) {
		b.Errorf("a:%+v != `aello world`", a)
	}
}

// BenchmarkString-8   	       1	1722255064 ns/op
func BenchmarkString(b *testing.B) {
	b.StartTimer()
	for i := 0; i < 100000000; i++ {
		bs := []byte("hello world")
		_ = String(bs)
		bs[0] = 'a'
	}
	b.StopTimer()
	bs := []byte("hello world")
	a := String(bs)
	bs[0] = 'a'
	if !reflect.DeepEqual("aello world", a) {
		b.Errorf("a:%+v != `aello world`", a)
	}
}

func TestSlice(t *testing.T) {
	a := string([]byte("hello world"))
	b := Slice(a)
	b = append(b, "hello world"...)
	println(String(b))

	if !reflect.DeepEqual([]byte("hello worldhello world"), b) {
		t.Errorf("a:%+v != `hello worldhello world`", string(b))
	}
}

// BenchmarkSlice0-8   	       1	1187713598 ns/op
func BenchmarkSlice0(b *testing.B) {
	for i := 0; i < 100000000; i++ {
		a := string([]byte("hello world"))
		bs := ([]byte)(a)
		_ = append(bs, "hello world"...)
	}
}

// BenchmarkSlice-8   	       1	4895001383 ns/op
func BenchmarkSlice(b *testing.B) {
	for i := 0; i < 100000000; i++ {
		a := string([]byte("hello world"))
		bs := Slice(a)
		_ = append(bs, "hello world"...)
	}
}
