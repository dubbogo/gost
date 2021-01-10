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

func TestGetTimerWheel(t *testing.T) {
	InitDefaultTimerWheel()
	tw := GetDefaultTimerWheel()
	if tw == nil {
		t.Fatal("default time wheel is nil")
	}
}

func TestUnix2Time(t *testing.T) {
	now := time.Now()
	nowUnix := Time2Unix(now)
	tm := Unix2Time(nowUnix)
	// time->unix有精度损失，所以只能在秒级进行比较
	if tm.Unix() != now.Unix() {
		t.Fatalf("@now:%#v, tm:%#v", now, tm)
	}
}

func TestUnixNano2Time(t *testing.T) {
	now := time.Now()
	nowUnix := Time2UnixNano(now)
	tm := UnixNano2Time(nowUnix)
	if tm.UnixNano() != now.UnixNano() {
		t.Fatalf("@now:%#v, tm:%#v", now, tm)
	}
}

func TestGetEndTime(t *testing.T) {
	dayEndTime := GetEndtime("day")
	t.Logf("today end time %q", dayEndTime)

	weekEndTime := GetEndtime("week")
	t.Logf("this week end time %q", weekEndTime)

	monthEndTime := GetEndtime("month")
	t.Logf("this month end time %q", monthEndTime)

	yearEndTime := GetEndtime("year")
	t.Logf("this year end time %q", yearEndTime)
}
