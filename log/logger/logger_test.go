/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file distributed with
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

package logger

import (
	"sync"
	"testing"
	"time"
)

type nopLogger struct{}

func (l *nopLogger) Info(args ...interface{})                  {}
func (l *nopLogger) Warn(args ...interface{})                  {}
func (l *nopLogger) Error(args ...interface{})                 {}
func (l *nopLogger) Debug(args ...interface{})                 {}
func (l *nopLogger) Fatal(args ...interface{})                 {}
func (l *nopLogger) Infof(format string, args ...interface{})  {}
func (l *nopLogger) Warnf(format string, args ...interface{})  {}
func (l *nopLogger) Errorf(format string, args ...interface{}) {}
func (l *nopLogger) Debugf(format string, args ...interface{}) {}
func (l *nopLogger) Fatalf(format string, args ...interface{}) {}

// TestSetLoggerConcurrentWithLogging checks that swapping the global logger
// while other goroutines log does not race (apache/dubbo-go#3823).
func TestSetLoggerConcurrentWithLogging(t *testing.T) {
	// Save the logger installed before this test and restore it afterwards,
	// so this test leaves no global state behind (e.g. a custom logger set
	// by an earlier test is not silently replaced by the default one).
	prev := GetLogger()
	t.Cleanup(func() { SetLogger(prev) })

	done := make(chan struct{})
	var wg sync.WaitGroup

	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-done:
					return
				default:
					Infof("info %d", 1)
					Warnf("warn %d", 2)
					Debugf("debug %d", 3)
					Errorf("error %d", 4)
				}
			}
		}()
	}

	loggers := []Logger{&nopLogger{}, &nopLogger{}}
	deadline := time.Now().Add(300 * time.Millisecond)
	for i := 0; time.Now().Before(deadline); i++ {
		SetLogger(loggers[i%len(loggers)])
	}
	close(done)
	wg.Wait()
}
