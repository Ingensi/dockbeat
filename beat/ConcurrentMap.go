package beat

import (
	"hash/fnv"
	"sync"
)

var SHARD_COUNT = 32

// A "thread" safe map of type string:BlkioData.
// To avoid lock bottlenecks this map is dived to several (SHARD_COUNT) map shards.
type ConcurrentMap []*ConcurrentMapShared
type ConcurrentMapShared struct {
	items        map[string]BlkioData
	sync.RWMutex // Read Write mutex, guards access to internal map.
}

// Creates a new concurrent map.
func NewBlkioMap() ConcurrentMap {
	m := make(ConcurrentMap, SHARD_COUNT)
	for i := 0; i < SHARD_COUNT; i++ {
		m[i] = &ConcurrentMapShared{items: make(map[string]BlkioData)}
	}
	return m
}

// Returns shard under given key
func (m ConcurrentMap) GetShard(key string) *ConcurrentMapShared {
	hasher := fnv.New32()
	hasher.Write([]byte(key))
	return m[int(hasher.Sum32())%SHARD_COUNT]
}

// Sets the given value under the specified key.
func (m *ConcurrentMap) Set(key string, value BlkioData) {
	// Get map shard.
	shard := m.GetShard(key)
	shard.Lock()
	defer shard.Unlock()
	shard.items[key] = value
}

// Retrieves an element from map under given key.
func (m ConcurrentMap) Get(key string) (BlkioData, bool) {
	// Get shard
	shard := m.GetShard(key)
	shard.RLock()
	defer shard.RUnlock()

	// Get item from shard.
	val, ok := shard.items[key]
	return val, ok
}

// Removes an element from the map.
func (m *ConcurrentMap) Remove(key string) {
	// Try to get shard.
	shard := m.GetShard(key)
	shard.Lock()
	defer shard.Unlock()
	delete(shard.items, key)
}

// Used by the Iter & IterBuffered functions to wrap two variables together over a channel,
type Tuple struct {
	Key string
	Val BlkioData
}

// Returns an iterator which could be used in a for range loop.
func (m ConcurrentMap) Iter() <-chan Tuple {
	ch := make(chan Tuple)
	go func() {
		// Foreach shard.
		for _, shard := range m {
			// Foreach key, value pair.
			shard.RLock()
			for key, val := range shard.items {
				ch <- Tuple{key, val}
			}
			shard.RUnlock()
		}
		close(ch)
	}()
	return ch
}
