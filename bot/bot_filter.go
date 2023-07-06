package bot

import (
	"sync"
	"time"
)

type filterManager struct {
	mu      sync.Mutex
	entries map[string]time.Time
}

func (fm *filterManager) add(id string) {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	fm.entries[id] = time.Now()
}

func (fm *filterManager) loop() {
	for {
		time.Sleep(time.Minute * 5)
		fm.mu.Lock()
		for k, v := range fm.entries {
			if time.Since(v) > time.Hour {
				delete(fm.entries, k)
			}
		}
		fm.mu.Unlock()
	}
}

// check if id in entries, if already exists, return true that need filter
func (fm *filterManager) needFilter(id string) bool {
	_, ok := fm.entries[id]
	return ok
}
