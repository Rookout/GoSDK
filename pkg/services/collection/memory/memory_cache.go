package memory

type memCache struct {
	loaded    bool
	cacheAddr uint64
	cache     []byte
	mem       MemoryReader
}

func (m *memCache) contains(addr uint64, size int) bool {
	return addr >= m.cacheAddr && addr <= (m.cacheAddr+uint64(len(m.cache)-size))
}

func (m *memCache) ReadMemory(data []byte, addr uint64) (n int, err error) {
	if m.contains(addr, len(data)) {
		if !m.loaded {
			_, err := m.mem.ReadMemory(m.cache, m.cacheAddr)
			if err != nil {
				return 0, err
			}
			m.loaded = true
		}
		copy(data, m.cache[addr-m.cacheAddr:])
		return len(data), nil
	}

	return m.mem.ReadMemory(data, addr)
}

func CacheMemory(mem MemoryReader, addr uint64, size int) MemoryReader {
	if !cacheEnabled {
		return mem
	}
	if size <= 0 {
		return mem
	}
	switch cacheMem := mem.(type) {
	case *memCache:
		if cacheMem.contains(addr, size) {
			return mem
		}
	case *CompositeMemory:
		return mem
	}
	return &memCache{false, addr, make([]byte, size), mem}
}
