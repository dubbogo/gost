
package gxruntime

import (
    "os"
	"testing"
	"time"
)

// exists returns whether the given file or directory exists
func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil { return true, nil }
	if os.IsNotExist(err) { return false, nil }
	return false, err
}

func t1(t *testing.T) {
	t.Logf("current os cpu number %d, memory limit %d bytes", GetCPUNum(), GetMemoryLimit())
	t.Logf("current prcess thread number %d", GetThreadNum())
	go func() {
		time.Sleep(10e9)
	}()

	grNum := GetGoroutineNum()
	if grNum < 2 {
		t.Errorf("current prcess goroutine number %d", grNum)
	}

	cpu, err := GetProcessCPUStat()
	if err != nil {
		t.Errorf("GetProcessCPUStat() = error %+v", err)
	}
	t.Logf("process cpu stat %v", cpu)

	size := 100 * 1024 * 1024
	arr := make([]byte, size)
	for idx, _ := range arr {
		arr[idx] = byte(idx / 255)
	}
	memoryStat, err := GetProcessMemoryStat()
	if err != nil {
		t.Errorf("GetProcessMemoryStat() = error %+v", err)
	}
	//t.Logf("process memory usage stat %v", memoryStat)
	if memoryStat <= uint64(size) {
		t.Errorf("memory usage stat %d < %d", memoryStat, size)
	}

	memoryUsage, err := GetProcessMemoryPercent()
	if err != nil {
		t.Errorf("GetProcessMemoryPercent() = error %+v", err)
	}
	t.Logf("process memory usage percent %v", memoryUsage)

	if ok, _ := exists(cgroupMemLimitPath); ok {
		memoryLimit, err := GetCgroupMemoryLimit()
		if err != nil {
			t.Errorf("GetCgroupMemoryLimit() = error %+v", err)
		}
		t.Logf("CGroupMemoryLimit() = %d", memoryLimit)

		memoryPercent, err := GetCgroupProcessMemoryPercent()
		if err != nil {
			t.Errorf("GetCgroupProcessMemoryPercent(ps:%d) = error %+v", CurrentPID, err)
		}
		t.Logf("GetCgroupProcessMemoryPercent(ps:%d) = %+v", CurrentPID, memoryPercent)

	}
}

