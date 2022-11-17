package memory




const FakeAddress = 0xbeef0000

const cacheEnabled = true





type MemoryReader interface {
	
	ReadMemory(buf []byte, addr uint64) (n int, err error)
}
