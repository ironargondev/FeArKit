package keylogger

import "sync"

var KeyloggerStorage = NewMemoryStorage()

type MemoryStorage struct {
	mu   sync.Mutex
	data []string
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		data: make([]string, 0),
	}
}

func (ms *MemoryStorage) Write(entry string) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.data = append(ms.data, entry)
}

func (ms *MemoryStorage) ReadAndClear() []string {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	out := ms.data
	// "Pop" the bytes by clearing them from storage after returning them
	ms.data = []string{}
	return out
}