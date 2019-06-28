package gxstrings

import (
	"testing"
)

// go test  -v slice_test.go  slice.go

// slice转string之后，如果slice的值有变化，string也会跟着改变
func TestString(t *testing.T) {
	b := []byte("hello world")
	a := String(b)
	b[0] = 'a'
	println(a) //output  aello world
}

func TestSlice(t *testing.T) {
	// 编译器会把 "hello, world" 这个字符串常量的字节数组分配在没有写权限的 memory segment
	a := "hello world"
	b := Slice(a)

	// !!! 上面这个崩溃在defer里面是recover不回来的，真的就崩溃了，原因可能就跟c的非法内存访问一样，os不跟你玩了
	// b[0] = 'a' //这里就等着崩溃吧

	//但是可以这样，因为go又重新给b分配了内存
	b = append(b, "hello world"...)
	println(String(b)) // output: hello worldhello world
}

func TestSlice2(t *testing.T) {
	// 你可以动态生成一个字符串，使其分配在可写的区域，比如 gc heap，那么就不会崩溃了。
	a := string([]byte("hello world"))
	b := Slice(a)
	b = append(b, "hello world"...)
	println(String(b)) // output: hello worldhello world
}

