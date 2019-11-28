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
	"encoding/binary"
	"fmt"
	"math/big"
)

// Integer represents a integer value.
type Integer struct {
	bigInt big.Int

	// for hessian
	Signum int32
	Mag    []int

	FirstNonzeroIntNum int
	LowestSetBit       int
	BitLength          int
	BitCount           int
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

// FromMag set data from a array of big-endian unsigned uint32
// @see https://docs.oracle.com/javase/8/docs/api/java/math/BigInteger.html#BigInteger-int-byte:A-
func (i *Integer) FromSignAndMag(signum int32, mag []int) {
	if signum == 0 && len(mag) == 0 {
		return
	}

	i.Signum = signum
	i.Mag = mag

	bytes := make([]byte, 4*len(i.Mag))
	for j := 0; j < len(i.Mag); j++ {
		binary.BigEndian.PutUint32(bytes[j*4:(j+1)*4], uint32(i.Mag[j]))
	}
	i.bigInt = *i.bigInt.SetBytes(bytes)

	if i.Signum == -1 {
		i.bigInt.Neg(&i.bigInt)
	}
}

func (i *Integer) GetSignAndMag() (signum int32, mag []int) {
	signum = int32(i.bigInt.Sign())

	bytes := i.bigInt.Bytes()
	outOf4 := len(bytes) % 4
	if outOf4 > 0 {
		bytes = append(make([]byte, 4-outOf4), bytes...)
	}

	size := len(bytes) / 4

	mag = make([]int, size)

	for i := 0; i < size; i++ {
		mag[i] = int(binary.BigEndian.Uint32(bytes[i*4 : (i+1)*4]))
	}

	return
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

func (i *Integer) Bytes() []byte {
	return i.bigInt.Bytes()
}
