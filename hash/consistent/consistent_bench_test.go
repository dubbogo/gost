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
	"testing"
)

var benchmarkRingHash *Consistent

func BenchmarkFullBuild(b *testing.B) {
	for _, hostCount := range []int{1, 32, 256, 1024} {
		for _, replicaCount := range []int{10, 100} {
			name := fmt.Sprintf("hosts=%d/replicas=%d", hostCount, replicaCount)
			b.Run(name, func(b *testing.B) {
				// Prepare input outside the timed loop: each operation builds a new ring.
				hosts := make([]string, hostCount)
				for i := range hosts {
					hosts[i] = fmt.Sprintf("127.0.0.1:%d", 8000+i)
				}

				b.Run("Add", func(b *testing.B) {
					b.ReportAllocs()
					for b.Loop() {
						ring := NewConsistentHash(WithReplicaNum(replicaCount))
						for _, host := range hosts {
							ring.Add(host)
						}
						benchmarkRingHash = ring
					}
				})
				for _, variant := range batchVariants {
					b.Run(variant.name, func(b *testing.B) {
						b.ReportAllocs()
						for b.Loop() {
							ring := NewConsistentHash(WithReplicaNum(replicaCount))
							variant.add(ring, hosts)
							benchmarkRingHash = ring
						}
					})
				}
			})
		}
	}
}

// BenchmarkBatchAppend excludes rebuilding the initial ring from time and
// allocation metrics. Every iteration starts with the same logical state.
func BenchmarkBatchAppend(b *testing.B) {
	for _, existing := range []int{256, 1024} {
		for _, replicas := range []int{10, 100} {
			initial := make([]string, existing)
			for i := range initial {
				initial[i] = fmt.Sprintf("existing-%d", i)
			}
			for _, count := range []int{1, 32, 256} {
				for _, duplicates := range []bool{false, true} {
					hosts := make([]string, count)
					for i := range hosts {
						if duplicates {
							hosts[i] = initial[i]
						} else {
							hosts[i] = fmt.Sprintf("new-%d", i)
						}
					}
					name := fmt.Sprintf("existing=%d/replicas=%d/batch=%d/duplicates=%t", existing, replicas, count, duplicates)
					b.Run(name, func(b *testing.B) {
						for _, variant := range batchVariants {
							b.Run(variant.name, func(b *testing.B) {
								b.ReportAllocs()
								for b.Loop() {
									b.StopTimer()
									ring := NewConsistentHash(WithReplicaNum(replicas))
									ring.AddBatch(initial)
									b.StartTimer()
									variant.add(ring, hosts)
									benchmarkRingHash = ring
								}
							})
						}
					})
				}
			}
		}
	}
}

func BenchmarkRemoveBatch(b *testing.B) {
	for _, size := range []int{256, 1024} {
		for _, replicas := range []int{10, 100} {
			hosts := make([]string, size)
			for i := range hosts {
				hosts[i] = fmt.Sprintf("host-%d", i)
			}
			for _, count := range []int{1, 32, size} {
				name := fmt.Sprintf("hosts=%d/replicas=%d/remove=%d", size, replicas, count)
				b.Run(name, func(b *testing.B) {
					for _, batch := range []bool{false, true} {
						label := "Remove"
						if batch {
							label = "RemoveBatch"
						}
						b.Run(label, func(b *testing.B) {
							b.ReportAllocs()
							for b.Loop() {
								b.StopTimer()
								c := NewConsistentHash(WithReplicaNum(replicas))
								c.AddBatch(hosts)
								b.StartTimer()
								if batch {
									c.RemoveBatch(hosts[:count])
								} else {
									for _, h := range hosts[:count] {
										c.Remove(h)
									}
								}
								benchmarkRingHash = c
							}
						})
					}
				})
			}
		}
	}
}
