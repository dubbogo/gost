/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package gxruntime

import (
	"testing"
	"time"
)


func TestSysStat(t *testing.T) {
	t.Logf("current os cpu number %d", GetCPUNum())
	total, used, free, usedPercent := GetMemoryStat()
	t.Logf("memory: limit %d bytes, used %d bytes, free %d bytes, usedPercent %f", total, used, free, usedPercent)
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

	if IsCgroup() {
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

