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

package gxbig

import (
	"testing"
)

func BenchmarkRound(b *testing.B) {
	b.StopTimer()
	var roundTo Decimal
	tests := []struct {
		input    string
		scale    int
		inputDec Decimal
	}{
		{input: "123456789.987654321", scale: 1},
		{input: "15.1", scale: 0},
		{input: "15.5", scale: 0},
		{input: "15.9", scale: 0},
		{input: "-15.1", scale: 0},
		{input: "-15.5", scale: 0},
		{input: "-15.9", scale: 0},
		{input: "15.1", scale: 1},
		{input: "-15.1", scale: 1},
		{input: "15.17", scale: 1},
		{input: "15.4", scale: -1},
		{input: "-15.4", scale: -1},
		{input: "5.4", scale: -1},
		{input: ".999", scale: 0},
		{input: "999999999", scale: -9},
	}

	for i := 0; i < len(tests); i++ {
		tests[i].inputDec.FromBytes([]byte(tests[i].input))
	}

	b.StartTimer()
	for n := 0; n < b.N; n++ {
		for i := 0; i < len(tests); i++ {
			tests[i].inputDec.Round(&roundTo, tests[i].scale, ModeHalfEven)
		}
		for i := 0; i < len(tests); i++ {
			tests[i].inputDec.Round(&roundTo, tests[i].scale, ModeTruncate)
		}
		for i := 0; i < len(tests); i++ {
			tests[i].inputDec.Round(&roundTo, tests[i].scale, modeCeiling)
		}
	}
}
