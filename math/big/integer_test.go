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
