// The MIT License (MIT)

// Copyright (c) 2014 Derek Parker

// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

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
