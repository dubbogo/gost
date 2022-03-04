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
	"encoding/binary"
	"math"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
)

import (
	"github.com/pkg/errors"
	"golang.org/x/crypto/blake2b"
)

import (
	"github.com/dubbogo/gost/strings"
)

const (
	replicationFactor = 10
	maxBucketNum      = math.MaxUint32
)

var ErrNoHosts = errors.New("no hosts added")

type Options struct {
	HashFunc    HashFunc
	ReplicaNum  int
	MaxVnodeNum int
}

type Option func(option *Options)

func WithHashFunc(hash HashFunc) Option {
	return func(opts *Options) {
		opts.HashFunc = hash
	}
}

func WithReplicaNum(replicaNum int) Option {
	return func(opts *Options) {
		opts.ReplicaNum = replicaNum
	}
}

func WithMaxVnodeNum(maxVnodeNum int) Option {
	return func(opts *Options) {
		opts.MaxVnodeNum = maxVnodeNum
	}
}

type hashArray []uint32

// Len returns the length of the hashArray
func (h hashArray) Len() int { return len(h) }

// Less returns true if element i is less than element j
func (h hashArray) Less(i, j int) bool { return h[i] < h[j] }

// Swap exchanges elements i and j
func (h hashArray) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

type Host struct {
	Name string
	Load int64
}

type HashFunc func([]byte) uint64

func hash(key []byte) uint64 {
	out := blake2b.Sum512(key)
	return binary.LittleEndian.Uint64(out[:])
}

type Consistent struct {
	circle        map[uint32]string // hash -> node name
	sortedHashes  hashArray         // hash valid in ascending
	loadMap       map[string]*Host  // node name -> struct Host
	totalLoad     int64             // total load
	replicaFactor uint32
	bucketNum     uint32
	hashFunc      HashFunc

	sync.RWMutex
}

func NewConsistentHash(opts ...Option) *Consistent {
	options := Options{
		HashFunc:    hash,
		ReplicaNum:  replicationFactor,
		MaxVnodeNum: maxBucketNum,
	}

	for index := range opts {
		opts[index](&options)
	}

	return &Consistent{
		circle:        map[uint32]string{},
		loadMap:       map[string]*Host{},
		replicaFactor: uint32(options.ReplicaNum),
		bucketNum:     uint32(options.MaxVnodeNum),
		hashFunc:      options.HashFunc,
	}
}

func (c *Consistent) SetHashFunc(f HashFunc) {
	c.hashFunc = f
}

// eltKey generates a string key for an element with an index
func (c *Consistent) eltKey(elt string, idx int) string {
	return strconv.Itoa(idx) + elt
}

func (c *Consistent) hash(key string) uint32 {
	return uint32(c.hashFunc(gxstrings.Slice(key))) % c.bucketNum
}

// updateSortedHashes sort hashes in ascending
func (c *Consistent) updateSortedHashes() {
	hashes := c.sortedHashes[:0]
	// reallocate if we're holding on to too much (1/4th)
	if c.sortedHashes.Len()/int(c.replicaFactor*4) > len(c.circle) {
		hashes = nil
	}
	for k := range c.circle {
		hashes = append(hashes, k)
	}
	sort.Sort(hashes)
	c.sortedHashes = hashes
}

func (c *Consistent) Add(host string) {
	c.Lock()
	defer c.Unlock()

	c.add(host)
}

func (c *Consistent) add(host string) {
	if _, ok := c.loadMap[host]; ok {
		return
	}

	c.loadMap[host] = &Host{Name: host}
	for i := uint32(0); i < c.replicaFactor; i++ {
		h := c.hash(c.eltKey(host, int(i)))
		c.circle[h] = host
		c.sortedHashes = append(c.sortedHashes, h)
	}

	c.updateSortedHashes()
}

// Set sets all the elements in the hash. If there are existing elements not
// present in elts, they will be removed.
func (c *Consistent) Set(elts []string) {
	c.Lock()
	defer c.Unlock()

	for k := range c.loadMap {
		found := true
		for _, elt := range elts {
			if k == elt {
				found = false
				break
			}
		}

		if found {
			c.remove(k)
		}

		for _, elt := range elts {
			if _, ok := c.loadMap[elt]; !ok {
				c.add(elt)
			}
		}
	}
}

func (c *Consistent) Members() []string {
	c.RLock()
	defer c.RUnlock()

	m := make([]string, 0, len(c.loadMap))
	for k := range c.loadMap {
		m = append(m, k)
	}
	return m
}

// Get It returns ErrNoHosts if the ring has no hosts in it
func (c *Consistent) Get(key string) (string, error) {
	c.RLock()
	defer c.RUnlock()

	if len(c.circle) == 0 {
		return "", ErrNoHosts
	}
	return c.circle[c.sortedHashes[c.search(c.hash(key))]], nil
}

// GetHash It returns ErrNoHosts if the ring has no hosts in it
func (c *Consistent) GetHash(hashKey uint32) (string, error) {
	c.RLock()
	defer c.RUnlock()

	if len(c.circle) == 0 {
		return "", ErrNoHosts
	}
	return c.circle[c.sortedHashes[c.search(hashKey)]], nil
}

// GetTwo returns the two closest distinct elements to the name input in the circle
func (c *Consistent) GetTwo(name string) (string, string, error) {
	c.RLock()
	defer c.RUnlock()

	if len(c.circle) == 0 {
		return "", "", ErrNoHosts
	}

	i := c.search(c.hash(name))
	a := c.circle[c.sortedHashes[i]]

	if len(c.loadMap) == 1 {
		return a, "", nil
	}

	start := i
	var b string

	for i = start + 1; i != start; i++ {
		if i >= len(c.sortedHashes) {
			i = 0
		}
		b = c.circle[c.sortedHashes[i]]
		if b != a {
			break
		}
	}
	return a, b, nil
}

func sliceContainsMember(set []string, member string) bool {
	for _, m := range set {
		if m == member {
			return true
		}
	}
	return false
}

// GetN returns the N closest distinct elements to the name input in the circle
func (c *Consistent) GetN(name string, n int) ([]string, error) {
	c.RLock()
	defer c.RUnlock()

	if len(c.circle) == 0 {
		return nil, ErrNoHosts
	}

	if len(c.loadMap) < n {
		n = len(c.loadMap)
	}

	var (
		i     = c.search(c.hash(name))
		start = i
		res   = make([]string, 0, n)
		elem  = c.circle[c.sortedHashes[i]]
	)

	res = append(res, elem)

	if len(res) == n {
		return res, nil
	}

	for i = start + 1; i != start; i++ {
		if i >= len(c.sortedHashes) {
			i = 0
		}
		elem = c.circle[c.sortedHashes[i]]
		if !sliceContainsMember(res, elem) {
			res = append(res, elem)
		}
		if len(res) == n {
			break
		}
	}

	return res, nil
}

// GetLeast It uses Consistent Hashing With Bounded loads
// https://research.googleblog.com/2017/04/consistent-hashing-with-bounded-loads.html
// to pick the least loaded host that can serve the key
// It returns ErrNoHosts if the ring has no hosts in it.
func (c *Consistent) GetLeast(key string) (string, error) {
	c.RLock()
	defer c.RUnlock()

	if len(c.circle) == 0 {
		return "", ErrNoHosts
	}

	idx := c.search(c.hash(key))

	i := idx
	for {
		host := c.circle[c.sortedHashes[i]]
		if c.loadOK(host) {
			return host, nil
		}
		i++
		if i >= len(c.circle) {
			i = 0
		}
	}
}

func (c *Consistent) search(key uint32) int {
	idx := sort.Search(len(c.sortedHashes), func(i int) bool { return c.sortedHashes[i] >= key })
	if idx >= len(c.sortedHashes) {
		return 0
	}
	return idx
}

// UpdateLoad Sets the load of `host` to the given `load`
func (c *Consistent) UpdateLoad(host string, load int64) {
	c.Lock()
	defer c.Unlock()

	if _, ok := c.loadMap[host]; !ok {
		return
	}

	c.totalLoad -= c.loadMap[host].Load
	c.loadMap[host].Load = load
	c.totalLoad += load
}

// Inc Increments the load of host by 1
// should only be used with if you obtained a host with GetLeast
func (c *Consistent) Inc(host string) {
	c.Lock()
	defer c.Unlock()

	atomic.AddInt64(&c.loadMap[host].Load, 1)
	atomic.AddInt64(&c.totalLoad, 1)
}

// Done Decrements the load of host by 1
// should only be used with if you obtained a host with GetLeast
func (c *Consistent) Done(host string) {
	c.Lock()
	defer c.Unlock()

	if _, ok := c.loadMap[host]; !ok {
		return
	}

	atomic.AddInt64(&c.loadMap[host].Load, -1)
	atomic.AddInt64(&c.totalLoad, 1)
}

// Remove Deletes host from the ring
func (c *Consistent) Remove(host string) bool {
	c.Lock()
	defer c.Unlock()
	return c.remove(host)
}

func (c *Consistent) remove(host string) bool {
	for i := uint32(0); i < c.replicaFactor; i++ {
		h := c.hash(c.eltKey(host, int(i)))
		delete(c.circle, h)
		c.delSlice(h)
	}

	if _, ok := c.loadMap[host]; ok {
		atomic.AddInt64(&c.totalLoad, -c.loadMap[host].Load)
		delete(c.loadMap, host)
	}
	return true
}

// Hosts Return the list of hosts in the ring
func (c *Consistent) Hosts() []string {
	c.RLock()
	defer c.RUnlock()

	hosts := make([]string, 0, len(c.loadMap))
	for k := range c.loadMap {
		hosts = append(hosts, k)
	}
	return hosts
}

// GetLoads Returns the loads of all the hosts
func (c *Consistent) GetLoads() map[string]int64 {
	loads := make(map[string]int64, len(c.loadMap))

	for k, v := range c.loadMap {
		loads[k] = v.Load
	}
	return loads
}

// MaxLoad Returns the maximum load of the single host
// which is:
// (total_load/number_of_hosts)*1.25
// total_load = is the total number of active requests served by hosts
// for more info:
// https://research.googleblog.com/2017/04/consistent-hashing-with-bounded-loads.html
func (c *Consistent) MaxLoad() int64 {
	if c.totalLoad == 0 {
		c.totalLoad = 1
	}

	avgLoadPerNode := float64(c.totalLoad / int64(len(c.loadMap)))
	if avgLoadPerNode == 0 {
		avgLoadPerNode = 1
	}
	avgLoadPerNode = math.Ceil(avgLoadPerNode * 1.25)
	return int64(avgLoadPerNode)
}

func (c *Consistent) loadOK(host string) bool {
	// a safety check if someone performed c.Done more than needed
	if c.totalLoad < 0 {
		c.totalLoad = 0
	}

	var avgLoadPerNode float64
	avgLoadPerNode = float64((c.totalLoad + 1) / int64(len(c.loadMap)))
	if avgLoadPerNode == 0 {
		avgLoadPerNode = 1
	}
	avgLoadPerNode = math.Ceil(avgLoadPerNode * 1.25)

	bhost, ok := c.loadMap[host]
	if !ok {
		panic("given host(" + bhost.Name + ") not in loadsMap")
	}

	if float64(bhost.Load)+1 <= avgLoadPerNode {
		return true
	}

	return false
}

func (c *Consistent) delSlice(val uint32) {
	idx := -1
	l := 0
	r := len(c.sortedHashes) - 1
	for l <= r {
		m := (l + r) / 2
		if c.sortedHashes[m] == val {
			idx = m
			break
		} else if c.sortedHashes[m] < val {
			l = m + 1
		} else if c.sortedHashes[m] > val {
			r = m - 1
		}
	}
	if idx != -1 {
		c.sortedHashes = append(c.sortedHashes[:idx], c.sortedHashes[idx+1:]...)
	}
}
