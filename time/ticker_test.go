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

// Package gxtime encapsulates some golang.time functions
package gxtime

import (
	"testing"
	"time"
)

import (
	gxlog "github.com/dubbogo/gost/log"
)

func TestTickFunc(t *testing.T) {
	var (
		//num     int
		cw CountWatch
		//xassert *assert.Assertions
	)

	InitDefaultTimerWheel()

	f := func() {
		gxlog.CInfo("timer costs:%dms", cw.Count()/1e6)
	}

	//num = 3
	//xassert = assert.New(t)
	cw.Start()
	TickFunc(TimeSecondDuration(0.5), f)
	TickFunc(TimeSecondDuration(1.3), f)
	TickFunc(TimeSecondDuration(61.5), f)
	time.Sleep(62e9)
	//xassert.Equal(defaultTimerWheel.TimerNumber(), num, "") // just equal in this ut
}

func TestTicker_Reset(t *testing.T) {
	//var (
	//	ticker *Ticker
	//	wg     sync.WaitGroup
	//	cw     CountWatch
	//	xassert *assert.Assertions
	//)
	//
	//Init()
	//
	//f := func() {
	//	defer wg.Done()
	//	gxlog.CInfo("timer costs:%dms", cw.Count()/1e6)
	//	gxlog.CInfo("in timer func, timer number:%d", defaultTimerWheel.TimerNumber())
	//}
	//
	//xassert = assert.New(t)
	//wg.Add(1)
	//cw.Start()
	//ticker = TickFunc(TimeSecondDuration(1.5), f)
	//ticker.Reset(TimeSecondDuration(3.5))
	//time.Sleep(TimeSecondDuration(0.001))
	//xassert.Equal(defaultTimerWheel.TimerNumber(), 1, "") // just equal on this ut
	//wg.Wait()
}

func TestTicker_Stop(t *testing.T) {
	var (
		ticker *Ticker
		cw     CountWatch
		//xassert assert.Assertions
	)

	InitDefaultTimerWheel()

	f := func() {
		gxlog.CInfo("timer costs:%dms", cw.Count()/1e6)
	}

	cw.Start()
	ticker = TickFunc(TimeSecondDuration(4.5), f)
	// 添加是异步进行的，所以sleep一段时间再去检测timer number
	time.Sleep(TimeSecondDuration(0.001))
	//timerNumber := defaultTimerWheel.TimerNumber()
	//xassert.Equal(timerNumber, 1, "")
	time.Sleep(TimeSecondDuration(5))
	ticker.Stop()
	// 删除是异步进行的，所以sleep一段时间再去检测timer number
	//time.Sleep(TimeSecondDuration(0.001))
	//timerNumber = defaultTimerWheel.TimerNumber()
	//xassert.Equal(timerNumber, 0, "")
}
