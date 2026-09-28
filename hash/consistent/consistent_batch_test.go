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
	"fmt"
	"reflect"
	"slices"
	"sync"
	"testing"
)

// Both variants include the same public-operation locking boundary.
var batchVariants = []struct {
	name string
	add  func(*Consistent, []string)
}{
	{"AddBatch", (*Consistent).AddBatch},
	{"AddBatchMerge", func(c *Consistent, hosts []string) {
		if len(hosts) == 0 {
			return
		}
		c.Lock()
		defer c.Unlock()
		c.addBatchMerge(hosts)
	}},
}

func TestAddBatchMatchesSequentialAdd(t *testing.T) {
	for _, variant := range batchVariants {
		t.Run(variant.name, func(t *testing.T) {
			large := make([]string, 256)
			for i := range large {
				large[i] = fmt.Sprintf("host-%d", i)
			}
			tests := []struct {
				name    string
				opts    []Option
				initial []string
				batches [][]string
			}{
				{name: "empty", batches: [][]string{nil, {}}},
				{name: "single", batches: [][]string{{"a"}}},
				{name: "multiple", batches: [][]string{{"a", "b", "c"}}},
				{name: "duplicates", batches: [][]string{{"a", "b", "a", "c", "b"}}},
				{name: "existing_loads", initial: []string{"a", "b"}, batches: [][]string{nil, {}, {"b", "c", "a", "d"}, {"a", "b", "c", "d"}}},
				{name: "repeated_batch", batches: [][]string{{"a", "b"}, {"a", "b"}}},
				{name: "split_batches", batches: [][]string{{"a", "b"}, {"b", "c"}, {"d", "a"}}},
				{name: "single_replica", opts: []Option{WithReplicaNum(1)}, batches: [][]string{{"a", "b", "c"}}},
				{name: "custom_hash", opts: []Option{WithReplicaNum(13), WithMaxVnodeNum(1023), WithHashFunc(murmurHash)}, batches: [][]string{{"a", "b", "c"}}},
				{name: "small_bucket_space", opts: []Option{WithMaxVnodeNum(7)}, batches: [][]string{{"a", "b", "c"}, {"d", "a"}}},
				{name: "forced_collision", opts: []Option{WithReplicaNum(3), WithHashFunc(func([]byte) uint64 { return 7 })}, batches: [][]string{{"a", "b", "a"}, {"c", "b"}}},
				{name: "large", batches: [][]string{large}},
			}
			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					sequential := NewConsistentHash(tt.opts...)
					batch := NewConsistentHash(tt.opts...)
					for i, host := range tt.initial {
						sequential.Add(host)
						batch.Add(host)
						sequential.UpdateLoad(host, int64(i+1))
						batch.UpdateLoad(host, int64(i+1))
					}
					for i, hosts := range tt.batches {
						t.Run(fmt.Sprintf("batch_%d", i), func(t *testing.T) {
							for _, host := range hosts {
								sequential.Add(host)
							}
							variant.add(batch, hosts)
							assertBatchRingEquivalent(t, sequential, batch)
						})
					}
				})
			}

		})
	}
}

// These assertions run after mutations finish; neither ring is accessed concurrently.
func assertBatchRingEquivalent(t *testing.T, want, got *Consistent) {
	t.Helper()
	if !reflect.DeepEqual(want.circle, got.circle) {
		t.Fatalf("circle mismatch: want %v, got %v", want.circle, got.circle)
	}
	if !slices.Equal(want.sortedHashes, got.sortedHashes) {
		t.Fatalf("sortedHashes mismatch: want %v, got %v", want.sortedHashes, got.sortedHashes)
	}
	if !reflect.DeepEqual(want.loadMap, got.loadMap) {
		t.Fatalf("loadMap mismatch: want %v, got %v", want.GetLoads(), got.GetLoads())
	}
	if want.totalLoad != got.totalLoad {
		t.Fatalf("totalLoad: want %d, got %d", want.totalLoad, got.totalLoad)
	}
	if len(got.sortedHashes) != len(got.circle) {
		t.Fatal("hash index length differs from circle size")
	}
	for i, h := range got.sortedHashes {
		if i > 0 && got.sortedHashes[i-1] >= h {
			t.Fatalf("hash index is not strictly increasing at %d", i)
		}
		host, ok := got.circle[h]
		if !ok {
			t.Fatalf("hash %d missing from circle", h)
		}
		if _, ok := got.loadMap[host]; !ok {
			t.Fatalf("host %q missing from loadMap", host)
		}
	}
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("request-%d", i)
		wh, we := want.Get(key)
		gh, ge := got.Get(key)
		if wh != gh || we != ge {
			t.Fatalf("Get(%q): want (%q, %v), got (%q, %v)", key, wh, we, gh, ge)
		}
		if len(got.circle) == 0 && ge != ErrNoHosts {
			t.Fatal("empty ring must return ErrNoHosts")
		}
	}
	const maxHash uint32 = ^uint32(0)
	probes := []uint32{0, maxHash}
	for _, h := range want.sortedHashes {
		probes = append(probes, h)
		if h > 0 {
			probes = append(probes, h-1)
		}
		if h < maxHash {
			probes = append(probes, h+1)
		}
	}
	for _, h := range probes {
		wh, we := want.GetHash(h)
		gh, ge := got.GetHash(h)
		if wh != gh || we != ge {
			t.Fatalf("GetHash(%d): want (%q, %v), got (%q, %v)", h, wh, we, gh, ge)
		}
	}
}

func TestAddBatchConcurrentReads(t *testing.T) {
	for _, variant := range batchVariants {
		t.Run(variant.name, func(t *testing.T) {
			got := NewConsistentHash()
			want := NewConsistentHash()
			got.Add("seed")
			want.Add("seed")
			got.UpdateLoad("seed", 5)
			want.UpdateLoad("seed", 5)

			const readers = 4
			for round := 0; round < 16; round++ {
				hosts := []string{"seed"}
				for i := 0; i < 8; i++ {
					hosts = append(hosts, fmt.Sprintf("host-%d-%d", round, i))
				}
				for _, host := range hosts {
					want.Add(host)
				}
				// Build an immutable set before starting workers. Queries may return
				// hosts from either side of the batch, but never an unknown host.
				allowed := make(map[string]bool)
				for _, host := range want.Members() {
					allowed[host] = true
				}
				start := make(chan struct{})
				var ready, done sync.WaitGroup
				ready.Add(readers + 1)
				done.Add(readers + 1)
				for reader := 0; reader < readers; reader++ {
					go func(reader int) {
						defer done.Done()
						ready.Done()
						<-start
						for i := 0; i < 100; i++ {
							key := fmt.Sprintf("reader-%d-key-%d", reader, i)
							host, err := got.Get(key)
							if err != nil || !allowed[host] {
								t.Errorf("Get(%q) = (%q, %v)", key, host, err)
								return
							}
							h := got.Hash(key)
							host, err = got.GetHash(h)
							if err != nil || !allowed[host] {
								t.Errorf("GetHash(%d) = (%q, %v)", h, host, err)
								return
							}
						}
					}(reader)
				}
				go func() {
					defer done.Done()
					ready.Done()
					<-start
					variant.add(got, hosts)
				}()
				ready.Wait()
				close(start)
				done.Wait()
				// Internal maps are inspected only after all workers have stopped.
				assertBatchRingEquivalent(t, want, got)
			}

		})
	}
}

func TestAddBatchCollisionOrder(t *testing.T) {
	for _, variant := range batchVariants {
		t.Run(variant.name, func(t *testing.T) {
			c := NewConsistentHash(WithHashFunc(func([]byte) uint64 { return 7 }))
			variant.add(c, []string{"a", "b", "a"})
			if len(c.circle) != 1 || !slices.Equal(c.sortedHashes, hashArray{7}) || c.circle[7] != "b" {
				t.Fatalf("expected one position owned by b, got circle=%v, hashes=%v", c.circle, c.sortedHashes)
			}
			c.UpdateLoad("b", 5)
			variant.add(c, []string{"c", "b"})
			if c.circle[7] != "c" || len(c.loadMap) != 3 || c.loadMap["b"].Load != 5 || c.totalLoad != 5 {
				t.Fatalf("unexpected state after second batch: circle=%v, loads=%v, total=%d", c.circle, c.GetLoads(), c.totalLoad)
			}

		})
	}
}

func TestAddBatchMergePositions(t *testing.T) {
	positions := map[string]uint64{"0a": 20, "0b": 60, "0c": 10, "0d": 40, "0e": 80, "0f": 40}
	opts := []Option{WithReplicaNum(1), WithHashFunc(func(key []byte) uint64 { return positions[string(key)] })}
	for _, variant := range batchVariants {
		t.Run(variant.name, func(t *testing.T) {
			got, want := NewConsistentHash(opts...), NewConsistentHash(opts...)
			for _, h := range []string{"a", "b"} {
				got.Add(h)
				want.Add(h)
			}
			hosts := []string{"c", "d", "e", "f", "d"}
			for _, h := range hosts {
				want.Add(h)
			}
			variant.add(got, hosts)
			assertBatchRingEquivalent(t, want, got)
			if !slices.Equal(got.sortedHashes, hashArray{10, 20, 40, 60, 80}) || got.circle[40] != "f" {
				t.Fatalf("unexpected merged index or ownership: %v, %v", got.sortedHashes, got.circle)
			}
		})
	}
}

func TestRemoveBatchMatchesSequentialRemove(t *testing.T) {
	for _, collision := range []bool{false, true} {
		t.Run(fmt.Sprintf("collisions=%t", collision), func(t *testing.T) {
			for _, tc := range []struct {
				name    string
				initial []string
				batches [][]string
			}{
				{"empty_ring", nil, [][]string{nil, {}, {"missing"}}},
				{"empty_input", []string{"a", "b"}, [][]string{nil, {}}},
				{"missing", []string{"a", "b"}, [][]string{{"missing"}}},
				{"duplicates", []string{"a", "b", "c"}, [][]string{{"a", "a", "missing"}, {"a"}}},
				{"all", []string{"a", "b", "c"}, [][]string{{"c", "b", "a"}, {"c"}}},
				{"split", []string{"a", "b", "c", "d"}, [][]string{{"a", "c"}, {"d"}, {"b"}}},
			} {
				t.Run(tc.name, func(t *testing.T) {
					opts := []Option{WithReplicaNum(13)}
					if collision {
						opts = append(opts, WithMaxVnodeNum(7))
					}
					want, got := NewConsistentHash(opts...), NewConsistentHash(opts...)
					for i, h := range tc.initial {
						want.Add(h)
						got.Add(h)
						want.UpdateLoad(h, int64(i+1))
						got.UpdateLoad(h, int64(i+1))
					}
					for _, hosts := range tc.batches {
						for _, h := range hosts {
							want.Remove(h)
						}
						got.RemoveBatch(hosts)
						assertBatchRingEquivalent(t, want, got)
					}
					// Verify subsequent additions also work with the compacted index.
					want.Add("new")
					got.AddBatch([]string{"new"})
					assertBatchRingEquivalent(t, want, got)
				})
			}
		})
	}
}

func TestRemoveBatchLegacyCollisionSemantics(t *testing.T) {
	for _, batch := range []bool{false, true} {
		t.Run(fmt.Sprintf("batch=%t", batch), func(t *testing.T) {
			c := NewConsistentHash(WithHashFunc(func([]byte) uint64 { return 7 }))
			c.AddBatch([]string{"a", "b"})
			c.UpdateLoad("a", 3)
			c.UpdateLoad("b", 5)
			remove := func(hosts []string) {
				if batch {
					c.RemoveBatch(hosts)
				} else {
					for _, h := range hosts {
						c.Remove(h)
					}
				}
			}
			// Legacy Remove deletes a colliding position even for an absent host.
			remove([]string{"missing"})
			if len(c.circle) != 0 || len(c.sortedHashes) != 0 || len(c.loadMap) != 2 || c.totalLoad != 8 {
				t.Fatalf("missing-host semantics differ: circle=%v loads=%v total=%d", c.circle, c.GetLoads(), c.totalLoad)
			}
			// Recreate the shared position, then remove a different existing owner.
			c.Add("c")
			remove([]string{"a", "a"})
			if len(c.circle) != 0 || len(c.sortedHashes) != 0 || len(c.loadMap) != 2 || c.totalLoad != 5 {
				t.Fatalf("collision or duplicate semantics differ: circle=%v loads=%v total=%d", c.circle, c.GetLoads(), c.totalLoad)
			}
			remove([]string{"c"})
			remove([]string{"b", "b"})
			if len(c.circle) != 0 || len(c.sortedHashes) != 0 || len(c.loadMap) != 0 || c.totalLoad != 0 {
				t.Fatal("ring not empty after deleting all hosts")
			}
			if _, err := c.Get("key"); err != ErrNoHosts {
				t.Fatalf("want ErrNoHosts, got %v", err)
			}
		})
	}
}

func TestRemoveBatchConcurrentReads(t *testing.T) {
	got, want := NewConsistentHash(), NewConsistentHash()
	hosts := make([]string, 128)
	for i := range hosts {
		hosts[i] = fmt.Sprintf("host-%d", i)
		got.Add(hosts[i])
		want.Add(hosts[i])
	}
	allowed := make(map[string]bool)
	for _, h := range hosts {
		allowed[h] = true
	}
	start := make(chan struct{})
	var wg sync.WaitGroup
	for reader := 0; reader < 4; reader++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for i := 0; i < 500; i++ {
				key := fmt.Sprintf("key-%d", i)
				h, err := got.Get(key)
				if err != nil || !allowed[h] {
					t.Errorf("Get returned %q, %v", h, err)
					return
				}
				h, err = got.GetHash(got.Hash(key))
				if err != nil || !allowed[h] {
					t.Errorf("GetHash returned %q, %v", h, err)
					return
				}
			}
		}()
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-start
		for i := 0; i < 120; i += 8 {
			got.RemoveBatch(hosts[i : i+8])
		}
	}()
	close(start)
	wg.Wait()
	for _, h := range hosts[:120] {
		want.Remove(h)
	}
	assertBatchRingEquivalent(t, want, got)
}
