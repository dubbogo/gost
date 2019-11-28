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
	"fmt"
	"math/big"
)

// Integer represents a integer value.
type Integer struct {
	bigInt big.Int

	// for hessian
	Value string
}

func (Integer) JavaClassName() string {
	return "java.math.BigInteger"
}

// FromString set data from a 10-bases number
func (i *Integer) FromString(s string) error {
	intPtr, ok := i.bigInt.SetString(s, 10)
	if !ok || intPtr == nil {
		return fmt.Errorf("'%s' is not a 10-based number", s)
	}

	i.bigInt = *intPtr
	return nil
}

// FromBytes set data from a 10-bases number bytes
func (i *Integer) FromBytes(bytes []byte) error {
	i.bigInt = *i.bigInt.SetBytes(bytes)
	return nil
}

// GetBigInt getter
func (i *Integer) GetBigInt() big.Int {
	return i.bigInt
}

// SetBigInt setter
func (i *Integer) SetBigInt(bigInt big.Int) {
	i.bigInt = bigInt
}

func (i *Integer) String() string {
	return i.bigInt.String()
}
