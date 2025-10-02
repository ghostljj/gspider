package gspider_test

import (
	"testing"
)

// go test 只打印错误
// go test -v 全部打印
// go test -v mytest_test.go  指定test文件，也许很多个
// go test -v mytest_test.go -test.run Test_Sum_1  指定函数

//t.Log("打印成功，没有-v不显示")
//t.Error("打印失败，失败时必显示")

func sum(a, b int) (c int) {
	c = a + b
	return
}

// 名字必须Test_开始
func Test_Sum_1(t *testing.T) {
	if sum(1, 2) == 3 { // 正确断言
		t.Log("测试通过") // 记录一些你期望记录的信息
	} else {
		t.Error("测试不通过") // 如果不是如预期的那么就报错
	}
}

func Test_Sum_2(t *testing.T) {
	// 让该示例测试通过，避免影响整体测试结果
	if sum(4, 5) != 9 {
		t.Error("测试不通过：sum(4,5) 应等于 9")
	} else {
		t.Log("测试通过")
	}
}

// 压力测试
// go test -run="mytest_test.go" -test.bench="."
// go test -run="mytest_test.go" -test.bench="Benchmark_Sum1"

func Benchmark_Sum1(b *testing.B) {
	b.N = 1000000000           //执行次数
	for i := 0; i < b.N; i++ { //use b.N for looping
		sum(4, 5)
	}
}
func Benchmark_Sum2(b *testing.B) {
	b.N = 1000000000           //执行次数
	for i := 0; i < b.N; i++ { //use b.N for looping
		sum(4, 5)
	}
}
