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

package consistent

import (
	"sort"
	"testing"
	"testing/quick"
)

import (
	"github.com/spaolacci/murmur3"
)

func murmurHash(key []byte) uint64 {
	return murmur3.Sum64WithSeed(key, 3238918481)
}

func checkNum(num, expected int, t *testing.T) {
	if num != expected {
		t.Errorf("got %d, expected %d", num, expected)
	}
}

func TestNewConsistentHash(t *testing.T) {
	c := NewConsistentHash(WithReplicaNum(13), WithMaxVnodeNum(1023))
	if c == nil {
		t.Fatal("expected obj")
	}
	checkNum(int(c.replicaFactor), 13, t)
}

func TestAdd(t *testing.T) {
	c := NewConsistentHash(WithReplicaNum(10), WithMaxVnodeNum(1023))

	c.Add("127.0.0.1:8000")
	if len(c.sortedHashes) != replicationFactor {
		t.Fatal("vnodes number is incorrect")
	}

	checkNum(len(c.circle), 10, t)
	checkNum(len(c.sortedHashes), 10, t)
	if sort.IsSorted(c.sortedHashes) == false {
		t.Error("expected sorted hashes to be sorted")
	}

	c.Add("qwer")
	checkNum(len(c.circle), 20, t)
	checkNum(len(c.sortedHashes), 20, t)
	if sort.IsSorted(c.sortedHashes) == false {
		t.Error("expected sorted hashes to be sorted")
	}
}

func TestGet(t *testing.T) {
	c := NewConsistentHash(WithReplicaNum(10), WithMaxVnodeNum(1023))

	c.Add("127.0.0.1:8000")
	host, err := c.Get("127.0.0.1:8000")
	if err != nil {
		t.Fatal(err)
	}

	if host != "127.0.0.1:8000" {
		t.Fatal("returned host is not what expected")
	}
}

func TestGetHash(t *testing.T) {
	c := NewConsistentHash(WithReplicaNum(10), WithMaxVnodeNum(1023))

	c.Add("127.0.0.1:8000")
	host, err := c.GetHash(123)
	if err != nil {
		t.Fatal(err)
	}

	if host != "127.0.0.1:8000" {
		t.Fatal("returned host is not what expected")
	}
}

func TestGetEmpty(t *testing.T) {
	c := NewConsistentHash(WithReplicaNum(13), WithMaxVnodeNum(1023))
	_, err := c.Get("asdfsadfsadf")
	if err == nil {
		t.Error("expected error")
	}
	if err != ErrNoHosts {
		t.Error("expected empty circle error")
	}
}

func TestGetSingle(t *testing.T) {
	c := NewConsistentHash(WithReplicaNum(13), WithMaxVnodeNum(1023))
	c.Add("abcdefg")
	f := func(s string) bool {
		y, err := c.Get(s)
		if err != nil {
			t.Log("error: ", err)
			return false
		}
		// t.Logf("s = %q, y = %q", s, y)
		return y == "abcdefg"
	}
	if err := quick.Check(f, nil); err != nil {
		t.Fatal(err)
	}
}

type gtest struct {
	in  string
	out string
}

var gmtests = []gtest{
	{"ggg", "opqrstu"},
	{"hhh", "abcdefg"},
	{"iiiii", "hijklmn"},
}

func TestGetMultiple(t *testing.T) {
	c := NewConsistentHash(WithReplicaNum(10), WithMaxVnodeNum(1023))
	c.Add("abcdefg")
	c.Add("hijklmn")
	c.Add("opqrstu")
	for i, v := range gmtests {
		result, err := c.Get(v.in)
		if err != nil {
			t.Fatal(err)
		}
		if result != v.out {
			t.Errorf("%d. got %q, expected %q", i, result, v.out)
		}
	}
}

func TestGetMultipleQuick(t *testing.T) {
	c := NewConsistentHash(WithReplicaNum(13), WithMaxVnodeNum(1023))
	c.Add("abcdefg")
	c.Add("hijklmn")
	c.Add("opqrstu")
	f := func(s string) bool {
		y, err := c.Get(s)
		if err != nil {
			t.Log("error: ", err)
			return false
		}
		// t.Logf("s = %q, y = %q", s, y)
		return y == "abcdefg" || y == "hijklmn" || y == "opqrstu"
	}
	if err := quick.Check(f, nil); err != nil {
		t.Fatal(err)
	}
}

var rtestsBefore = []gtest{
	{"ggg", "abcdefg"},
	{"hhh", "abcdefg"},
	{"iiiii", "opqrstu"},
}

var rtestsAfter = []gtest{
	{"ggg", "abcdefg"},
	{"hhh", "abcdefg"},
	{"iiiii", "opqrstu"},
}

func TestGetMultipleRemove(t *testing.T) {
	c := NewConsistentHash(WithReplicaNum(20), WithMaxVnodeNum(1023), WithHashFunc(murmurHash))
	c.Add("abcdefg")
	c.Add("hijklmn")
	c.Add("opqrstu")
	for i, v := range rtestsBefore {
		result, err := c.Get(v.in)
		if err != nil {
			t.Fatal(err)
		}
		if result != v.out {
			t.Errorf("%d. got %q, expected %q before rm", i, result, v.out)
		}
	}
	c.Remove("hijklmn")
	for i, v := range rtestsAfter {
		result, err := c.Get(v.in)
		if err != nil {
			t.Fatal(err)
		}
		if result != v.out {
			t.Errorf("%d. got %q, expected %q after rm", i, result, v.out)
		}
	}
}

func TestGetMultipleRemoveQuick(t *testing.T) {
	c := NewConsistentHash(WithReplicaNum(20), WithMaxVnodeNum(1023), WithHashFunc(murmurHash))
	c.Add("abcdefg")
	c.Add("hijklmn")
	c.Add("opqrstu")
	c.Remove("opqrstu")
	f := func(s string) bool {
		y, err := c.Get(s)
		if err != nil {
			t.Log("error: ", err)
			return false
		}
		// t.Logf("s = %q, y = %q", s, y)
		return y == "abcdefg" || y == "hijklmn"
	}
	if err := quick.Check(f, nil); err != nil {
		t.Fatal(err)
	}
}
func TestGetTwo(t *testing.T) {
	c := NewConsistentHash(WithReplicaNum(13), WithMaxVnodeNum(10230))
	c.Add("abcdefg")
	c.Add("hijklmn")
	c.Add("opqrstu")
	a, b, err := c.GetTwo("99999999")
	if err != nil {
		t.Fatal(err)
	}
	if a == b {
		t.Error("a shouldn't equal b")
	}
	if a != "opqrstu" {
		t.Error("wrong a: ", a)
	}
	if b != "abcdefg" {
		t.Error("wrong b: ", b)
	}
}

func TestGetTwoQuick(t *testing.T) {
	c := NewConsistentHash(WithReplicaNum(20), WithMaxVnodeNum(1023), WithHashFunc(murmurHash))
	c.Add("abcdefg")
	c.Add("hijklmn")
	c.Add("opqrstu")
	f := func(s string) bool {
		a, b, err := c.GetTwo(s)
		if err != nil {
			t.Log("error: ", err)
			return false
		}
		if a == b {
			t.Log("a == b")
			return false
		}
		if a != "abcdefg" && a != "hijklmn" && a != "opqrstu" {
			t.Log("invalid a: ", a)
			return false
		}

		if b != "abcdefg" && b != "hijklmn" && b != "opqrstu" {
			t.Log("invalid b: ", b)
			return false
		}
		return true
	}
	if err := quick.Check(f, nil); err != nil {
		t.Fatal(err)
	}
}

func TestGetTwoOnlyTwoQuick(t *testing.T) {
	c := NewConsistentHash(WithReplicaNum(20), WithMaxVnodeNum(1023), WithHashFunc(murmurHash))
	c.Add("abcdefg")
	c.Add("hijklmn")
	f := func(s string) bool {
		a, b, err := c.GetTwo(s)
		if err != nil {
			t.Log("error: ", err)
			return false
		}
		if a == b {
			t.Log("a == b")
			return false
		}
		if a != "abcdefg" && a != "hijklmn" {
			t.Log("invalid a: ", a)
			return false
		}

		if b != "abcdefg" && b != "hijklmn" {
			t.Log("invalid b: ", b)
			return false
		}
		return true
	}
	if err := quick.Check(f, nil); err != nil {
		t.Fatal(err)
	}
}

func TestGetTwoOnlyOneInCircle(t *testing.T) {
	c := NewConsistentHash(WithReplicaNum(20), WithMaxVnodeNum(1023), WithHashFunc(murmurHash))
	c.Add("abcdefg")
	a, b, err := c.GetTwo("99999999")
	if err != nil {
		t.Fatal(err)
	}
	if a == b {
		t.Error("a shouldn't equal b")
	}
	if a != "abcdefg" {
		t.Error("wrong a: ", a)
	}
	if b != "" {
		t.Error("wrong b: ", b)
	}
}

func TestGetN(t *testing.T) {
	c := NewConsistentHash(WithReplicaNum(20), WithMaxVnodeNum(1023), WithHashFunc(murmurHash))
	c.Add("abcdefg")
	c.Add("hijklmn")
	c.Add("opqrstu")
	members, err := c.GetN("9999999", 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(members) != 3 {
		t.Error("expected 3 members instead of ", len(members))
	}
	if members[0] != "abcdefg" {
		t.Error("wrong members[0]: ", members[0])
	}
	if members[1] != "hijklmn" {
		t.Error("wrong members[1]: ", members[1])
	}
	if members[2] != "opqrstu" {
		t.Error("wrong members[2]: ", members[2])
	}
}

func TestGetNLess(t *testing.T) {
	c := NewConsistentHash(WithReplicaNum(20), WithMaxVnodeNum(1023), WithHashFunc(murmurHash))
	c.Add("abcdefg")
	c.Add("hijklmn")
	c.Add("opqrstu")
	members, err := c.GetN("99999999", 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(members) != 2 {
		t.Error("expected 2 members instead of ", len(members))
	}
	if members[0] != "abcdefg" {
		t.Error("wrong members[0]: ", members[0])
	}
	if members[1] != "hijklmn" {
		t.Error("wrong members[1]: ", members[1])
	}
}

func TestGetNMore(t *testing.T) {
	c := NewConsistentHash(WithReplicaNum(20), WithMaxVnodeNum(1023), WithHashFunc(murmurHash))
	c.Add("abcdefg")
	c.Add("hijklmn")
	c.Add("opqrstu")
	members, err := c.GetN("9999999", 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(members) != 3 {
		t.Error("expected 3 members instead of ", len(members))
	}
	if members[0] != "abcdefg" {
		t.Error("wrong members[0]: ", members[0])
	}
	if members[1] != "hijklmn" {
		t.Error("wrong members[1]: ", members[1])
	}
	if members[2] != "opqrstu" {
		t.Error("wrong members[2]: ", members[2])
	}
}

func TestGetNQuick(t *testing.T) {
	c := NewConsistentHash(WithReplicaNum(20), WithMaxVnodeNum(1023), WithHashFunc(murmurHash))
	c.Add("abcdefg")
	c.Add("hijklmn")
	c.Add("opqrstu")
	f := func(s string) bool {
		members, err := c.GetN(s, 3)
		if err != nil {
			t.Log("error: ", err)
			return false
		}
		if len(members) != 3 {
			t.Log("expected 3 members instead of ", len(members))
			return false
		}
		set := make(map[string]bool, 4)
		for _, member := range members {
			if set[member] {
				t.Log("duplicate error")
				return false
			}
			set[member] = true
			if member != "abcdefg" && member != "hijklmn" && member != "opqrstu" {
				t.Log("invalid member: ", member)
				return false
			}
		}
		return true
	}
	if err := quick.Check(f, nil); err != nil {
		t.Fatal(err)
	}
}

func TestGetNLessQuick(t *testing.T) {
	c := NewConsistentHash(WithReplicaNum(20), WithMaxVnodeNum(1023), WithHashFunc(murmurHash))
	c.Add("abcdefg")
	c.Add("hijklmn")
	c.Add("opqrstu")
	f := func(s string) bool {
		members, err := c.GetN(s, 2)
		if err != nil {
			t.Log("error: ", err)
			return false
		}
		if len(members) != 2 {
			t.Log("expected 2 members instead of ", len(members))
			return false
		}
		set := make(map[string]bool, 4)
		for _, member := range members {
			if set[member] {
				t.Log("duplicate error")
				return false
			}
			set[member] = true
			if member != "abcdefg" && member != "hijklmn" && member != "opqrstu" {
				t.Log("invalid member: ", member)
				return false
			}
		}
		return true
	}
	if err := quick.Check(f, nil); err != nil {
		t.Fatal(err)
	}
}

func TestGetNMoreQuick(t *testing.T) {
	c := NewConsistentHash(WithReplicaNum(20), WithMaxVnodeNum(1023), WithHashFunc(murmurHash))
	c.Add("abcdefg")
	c.Add("hijklmn")
	c.Add("opqrstu")
	f := func(s string) bool {
		members, err := c.GetN(s, 5)
		if err != nil {
			t.Log("error: ", err)
			return false
		}
		if len(members) != 3 {
			t.Log("expected 3 members instead of ", len(members))
			return false
		}
		set := make(map[string]bool, 4)
		for _, member := range members {
			if set[member] {
				t.Log("duplicate error")
				return false
			}
			set[member] = true
			if member != "abcdefg" && member != "hijklmn" && member != "opqrstu" {
				t.Log("invalid member: ", member)
				return false
			}
		}
		return true
	}
	if err := quick.Check(f, nil); err != nil {
		t.Fatal(err)
	}
}

func TestSet(t *testing.T) {
	c := NewConsistentHash(WithReplicaNum(20), WithMaxVnodeNum(1023), WithHashFunc(murmurHash))
	c.Add("abc")
	c.Add("def")
	c.Add("ghi")
	c.Set([]string{"jkl", "mno"})
	if len(c.loadMap) != 2 {
		t.Errorf("expected 2 elts, got %d", len(c.loadMap))
	}
	a, b, err := c.GetTwo("qwerqwerwqer")
	if err != nil {
		t.Fatal(err)
	}
	if a != "jkl" && a != "mno" {
		t.Error("expected jkl or mno, got ", a)
	}
	if b != "jkl" && b != "mno" {
		t.Error("expected jkl or mno, got ", b)
	}
	if a == b {
		t.Error("expected a != b, they were both ", a)
	}
	c.Set([]string{"pqr", "mno"})
	if len(c.loadMap) != 2 {
		t.Error("expected 2 elts, got ", len(c.loadMap))
	}
	a, b, err = c.GetTwo("qwerqwerwqer")
	if err != nil {
		t.Fatal(err)
	}
	if a != "pqr" && a != "mno" {
		t.Error("expected jkl or mno, got ", a)
	}
	if b != "pqr" && b != "mno" {
		t.Error("expected jkl or mno, got ", b)
	}
	if a == b {
		t.Error("expected a != b, they were both ", a)
	}
	c.Set([]string{"pqr", "mno"})
	if len(c.loadMap) != 2 {
		t.Error("expected 2 elts, got ", len(c.loadMap))
	}
	a, b, err = c.GetTwo("qwerqwerwqer")
	if err != nil {
		t.Fatal(err)
	}
	if a != "pqr" && a != "mno" {
		t.Error("expected jkl or mno, got ", a)
	}
	if b != "pqr" && b != "mno" {
		t.Error("expected jkl or mno, got ", b)
	}
	if a == b {
		t.Error("expected a != b, they were both ", a)
	}
}
func TestRemove(t *testing.T) {
	c := NewConsistentHash(WithReplicaNum(10), WithMaxVnodeNum(1023))

	c.Add("127.0.0.1:8000")
	c.Remove("127.0.0.1:8000")

	// if len(c.sortedHashes) != 0 && len(c.circle) != 0 {
	if c.sortedHashes.Len() != 0 && len(c.circle) != 0 && c.totalLoad == 0 {
		t.Fatal("remove is not working")
	}

}

func TestRemoveNonExisting(t *testing.T) {
	c := NewConsistentHash(WithReplicaNum(10), WithMaxVnodeNum(1023))
	c.Add("abcdefg")
	c.Remove("abcdefghijk")
	checkNum(len(c.circle), 10, t)
}

func TestGetLeast(t *testing.T) {
	c := NewConsistentHash(WithReplicaNum(10), WithMaxVnodeNum(1023))

	c.Add("127.0.0.1:8000")
	c.Add("92.0.0.1:8000")

	for i := 0; i < 100; i++ {
		host, err := c.GetLeast("92.0.0.1:80001")
		if err != nil {
			t.Fatal(err)
		}
		c.Inc(host)
	}

	for k, v := range c.GetLoads() {
		if v > c.MaxLoad() {
			t.Fatalf("host %s is overloaded. %d > %d\n", k, v, c.MaxLoad())
		}
	}
	t.Log("Max load per node", c.MaxLoad())
	t.Log(c.GetLoads())
}

func TestIncDone(t *testing.T) {
	c := NewConsistentHash(WithReplicaNum(10), WithMaxVnodeNum(1023))

	c.Add("127.0.0.1:8000")
	c.Add("92.0.0.1:8000")

	host, err := c.GetLeast("92.0.0.1:80001")
	if err != nil {
		t.Fatal(err)
	}

	c.Inc(host)
	if c.loadMap[host].Load != 1 {
		t.Fatalf("host %s load should be 1\n", host)
	}

	c.Done(host)
	if c.loadMap[host].Load != 0 {
		t.Fatalf("host %s load should be 0\n", host)
	}

}

func TestHosts(t *testing.T) {
	hosts := []string{
		"127.0.0.1:8000",
		"92.0.0.1:8000",
	}

	c := NewConsistentHash(WithReplicaNum(10), WithMaxVnodeNum(1023))
	for _, h := range hosts {
		c.Add(h)
	}
	t.Log("hosts in the ring", c.Hosts())

	addedHosts := c.Hosts()
	for _, h := range hosts {
		found := false
		for _, ah := range addedHosts {
			if h == ah {
				found = true
				break
			}
		}
		if !found {
			t.Fatal("missing host", h)
		}
	}
	c.Remove("127.0.0.1:8000")
	t.Log("hosts in the ring", c.Hosts())
}

func TestDelSlice(t *testing.T) {
	items := []uint32{0, 1, 2, 3, 5, 20, 22, 23, 25, 27, 28, 30, 35, 37, 1008, 1009}
	deletes := []uint32{25, 37, 1009, 3, 100000}

	c := &Consistent{}
	c.sortedHashes = append(c.sortedHashes, items...)

	t.Logf("before deletion%+v\n", c.sortedHashes)

	for _, val := range deletes {
		c.delSlice(val)
	}

	for _, val := range deletes {
		for _, item := range c.sortedHashes {
			if item == val {
				t.Fatalf("%d wasn't deleted\n", val)
			}
		}
	}

	t.Logf("after deletions: %+v\n", c.sortedHashes)
}
