package consistenthash

import (
	"sort"
	"strconv"
	"sync"
)

type ConsistentHash struct {
	keys     []uint32
	replicas int
	hash     func(data []byte) uint32
	sync.RWMutex
	hashMap map[uint32]int64
}

func NewConsistentHash(replicas int, fn func(data []byte) uint32) *ConsistentHash {
	c := &ConsistentHash{
		replicas: replicas,
		hash:     fn,
		hashMap:  make(map[uint32]int64),
	}
	return c
}

func (c *ConsistentHash) Add(keys ...int64) {
	c.Lock()
	for _, key := range keys {
		for i := 0; i < c.replicas; i++ {
			hash := c.hash([]byte(strconv.FormatInt(key, 16) + strconv.Itoa(i)))
			c.hashMap[hash] = key
			c.keys = append(c.keys, hash)
		}
	}
	c.Unlock()
	sort.Slice(c.keys, func(i, j int) bool {
		return c.keys[i] < c.keys[j]
	})
}

func (c *ConsistentHash) Get(key int64) int64 {
	c.RLock()
	defer c.RUnlock()

	if len(c.keys) == 0 {
		return 0
	}
	hash := c.hash([]byte(strconv.FormatInt(key, 16)))
	idx := sort.Search(len(c.keys), func(i int) bool {
		return c.keys[i] > hash
	})
	if idx == len(c.keys) {
		idx = 0
	}
	return c.hashMap[c.keys[idx]]
}

func (c *ConsistentHash) Remove(keys ...int64) {
	c.Lock()
	for _, key := range keys {
		for i := 0; i < c.replicas; i++ {
			hash := c.hash([]byte(strconv.FormatInt(key, 16) + strconv.Itoa(i)))
			delete(c.hashMap, hash)
			for i, k := range c.keys {
				if k == hash {
					c.keys = append(c.keys[:i], c.keys[i+1:]...)
				}
			}
		}
	}
	c.Unlock()
}

func (c *ConsistentHash) Keys() []int64 {
	c.RLock()
	defer c.RUnlock()

	var keys []int64
	for _, key := range c.hashMap {
		keys = append(keys, key)
	}
	return keys
}
