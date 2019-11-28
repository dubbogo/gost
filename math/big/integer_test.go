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
	"reflect"
	"testing"
)

func TestInteger(t *testing.T) {
	tests := []struct {
		name       string
		src        string
		wantString string
		wantErr    bool
	}{
		{`ten`, `10`, `10`, false},
		{`-ten`, `-10`, `-10`, false},
		{`30digits`, `123456789012345678901234567890`, `123456789012345678901234567890`, false},
		{`invalid-x010`, `x010`, ``, true},
		{`invalid-a010`, `a010`, ``, true},
		{`invalid-10x`, `10x`, ``, true},
		{`invalid-010x`, `010x`, ``, true},
		{`special-010`, `010`, `10`, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &Integer{}
			err := i.FromString(tt.src)
			if (err != nil) != tt.wantErr {
				t.Errorf("FromString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil && i.String() != tt.wantString {
				t.Errorf("String() got %v, want %v", i.String(), tt.wantString)
			}
		})
	}
}

func TestInteger_FromSignAndMag(t *testing.T) {
	type args struct {
		signum int32
		mag    []int
	}
	tests := []struct {
		name  string
		digit string
		args  args
	}{
		{`0`, `0`, args{0, []int{}}},
		{`1`, `1`, args{1, []int{1}}},
		{`2147483647`, `2147483647`, args{1, []int{2147483647}}},
		{`4294967295`, `4294967295`, args{1, []int{4294967295}}},
		{`4294967296`, `4294967296`, args{1, []int{1, 0}}},
		{`4294967298`, `4294967298`, args{1, []int{1, 2}}},
		{`1x2x3`, `18446744082299486211`, args{1, []int{1, 2, 3}}},
		{`-1`, `-1`, args{-1, []int{1}}},
		{`-4294967296`, `-4294967296`, args{-1, []int{1, 0}}},
		{`-4294967298`, `-4294967298`, args{-1, []int{1, 2}}},
		{`-1x2x3`, `-18446744082299486211`, args{-1, []int{1, 2, 3}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &Integer{}
			i.FromSignAndMag(tt.args.signum, tt.args.mag)
			if i.String() != tt.digit {
				t.Errorf("digit %s got = %s", tt.digit, i.String())
			}

			s := &Integer{}
			err := s.FromString(tt.digit)
			if err != nil {
				t.Error("FromString error = ", err)
			}

			sign, mag := s.GetSignAndMag()
			if !(sign == tt.args.signum && reflect.DeepEqual(mag, tt.args.mag)) {
				t.Error("want ", tt.args.signum, tt.args.mag,
					"got", sign, mag)
			}
		})
	}
}
