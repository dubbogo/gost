/*
 * MIT License
 *
 * Copyright (c) 2020 Mahendra Kanani
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

package rwmutex

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

const (
	start = 4
	end   = 44

	goRtnTotal = 100 * 1000
)

func TestGetNil(t *testing.T) {
	val := Get()
	if val != nil {
		panic("invalid scenario")
	}
}

func getSetDel(group *sync.WaitGroup) {
	for i := start; i < end-1; i++ {
		if i&3 == 0 { // multiple of 4
			Set(i)
			continue
		} else {
			val := Get().(int)
			if val >= i {
				panic("invalid")
			}
		}
	}
	Del()
	group.Done()
}

func getSet(group *sync.WaitGroup) {
	for i := start; i < end; i++ {
		if i&3 == 0 { // multiple of 4
			Set(i)
			continue
		} else {
			val := Get().(int)
			if val >= i {
				panic("invalid")
			}
		}
	}
	group.Done()
}

func TestSet(t *testing.T) {
	wg := &sync.WaitGroup{}
	wg.Add(goRtnTotal)
	for i := 0; i < goRtnTotal; i++ {
		go getSetDel(wg)
	}
	wg.Wait()
}

func BenchmarkSetGetDel(b *testing.B) {
	wg := &sync.WaitGroup{}
	goRtns := b.N
	wg.Add(goRtns)
	for i := 0; i < goRtns; i++ {
		go getSetDel(wg)
	}
	wg.Wait()
}

func BenchmarkSetGet(b *testing.B) {
	wg := &sync.WaitGroup{}
	goRtns := b.N
	wg.Add(goRtns)
	for i := 0; i < goRtns; i++ {
		go getSet(wg)
	}
	wg.Wait()
}

func TestGoRtnReUsageStatsWithoutDel(t *testing.T) {
	wg := &sync.WaitGroup{}
	//warm-up map
	wg.Add(goRtnTotal)
	for i := 0; i < goRtnTotal; i++ {
		go getSet(wg)
	}
	wg.Wait()

	iterations := 50
	samples := make([]time.Duration, iterations)
	// test and measure time
	startTime := time.Now()
	for j := 0; j < iterations; j++ {
		st := time.Now()
		wg.Add(goRtnTotal)
		for i := 0; i < goRtnTotal; i++ {
			go getSet(wg)
		}
		wg.Wait()
		samples[j] = time.Now().Sub(st)
	}

	total := time.Now().Sub(startTime)
	min, max, avg := minMaxAvg(samples)
	fmt.Printf("Min:%v, Max:%v, Avg:%v\n", min, max, (time.Duration)(avg))

	fmt.Println("time taken: ", total)
}

func TestGoRtnReUsageStatsWithDel(t *testing.T) {
	wg := &sync.WaitGroup{}
	//warm-up map
	wg.Add(goRtnTotal)
	for i := 0; i < goRtnTotal; i++ {
		go getSetDel(wg)
	}
	wg.Wait()

	iterations := 50
	samples := make([]time.Duration, iterations)
	// test and measure time
	startTime := time.Now()
	for j := 0; j < iterations; j++ {
		st := time.Now()
		wg.Add(goRtnTotal)
		for i := 0; i < goRtnTotal; i++ {
			go getSetDel(wg)
		}
		wg.Wait()
		samples[j] = time.Now().Sub(st)
	}

	total := time.Now().Sub(startTime)
	min, max, avg := minMaxAvg(samples)
	fmt.Printf("Min:%v, Max:%v, Avg:%v\n", min, max, (time.Duration)(avg))

	fmt.Println("time taken: ", total)
}

func minMaxAvg(samples []time.Duration) (time.Duration, time.Duration, int64) {
	min, max, avg := samples[0], samples[0], samples[0]
	for i := 1; i < len(samples); i++ {
		if min > samples[i] {
			min = samples[i]
		}
		if max < samples[i] {
			max = samples[i]
		}
		avg += samples[i]
	}
	return min, max, ((int64)(avg)) / ((int64)(len(samples)))
}
