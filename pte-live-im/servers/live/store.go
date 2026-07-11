package live

import (
	"context"
	"encoding/json"
	"sync"
)

// memStore Redis 不可用时的进程内兜底（单机开发）
type memStore struct {
	mu    sync.RWMutex
	sets  map[string]map[string]struct{}
	lists map[string][]string
	hashes map[string]map[string]string
}

var local = &memStore{
	sets:   make(map[string]map[string]struct{}),
	lists:  make(map[string][]string),
	hashes: make(map[string]map[string]string),
}

func (m *memStore) sadd(key, member string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.sets[key] == nil {
		m.sets[key] = make(map[string]struct{})
	}
	if _, ok := m.sets[key][member]; ok {
		return false
	}
	m.sets[key][member] = struct{}{}
	return true
}

func (m *memStore) srem(key, member string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.sets[key] == nil {
		return false
	}
	if _, ok := m.sets[key][member]; !ok {
		return false
	}
	delete(m.sets[key], member)
	return true
}

func (m *memStore) scard(key string) int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return int64(len(m.sets[key]))
}

func (m *memStore) delSet(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sets, key)
}

func (m *memStore) smembers(key string) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]string, 0, len(m.sets[key]))
	for k := range m.sets[key] {
		out = append(out, k)
	}
	return out
}

func (m *memStore) sismember(key, member string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.sets[key] == nil {
		return false
	}
	_, ok := m.sets[key][member]
	return ok
}

func (m *memStore) lpush(key, val string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lists[key] = append([]string{val}, m.lists[key]...)
}

func (m *memStore) lrange(key string, start, stop int) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	list := m.lists[key]
	if len(list) == 0 {
		return nil
	}
	if start < 0 {
		start = 0
	}
	if stop >= len(list) || stop < 0 {
		stop = len(list) - 1
	}
	if start > stop {
		return nil
	}
	out := make([]string, stop-start+1)
	copy(out, list[start:stop+1])
	return out
}

func (m *memStore) llen(key string) int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return int64(len(m.lists[key]))
}

func (m *memStore) hgetall(key string) map[string]string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	src := m.hashes[key]
	out := make(map[string]string, len(src))
	for k, v := range src {
		out[k] = v
	}
	return out
}

func (m *memStore) hset(key, field, val string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.hashes[key] == nil {
		m.hashes[key] = make(map[string]string)
	}
	m.hashes[key][field] = val
}

func (m *memStore) hget(key, field string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.hashes[key] == nil {
		return "", false
	}
	v, ok := m.hashes[key][field]
	return v, ok
}

func (m *memStore) del(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.lists, key)
}

func encodeJSON(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func decodeJSONList(raw []string, out interface{}) {
	items := make([]json.RawMessage, 0, len(raw))
	for _, s := range raw {
		items = append(items, json.RawMessage(s))
	}
	b, _ := json.Marshal(items)
	_ = json.Unmarshal(b, out)
}

func ctx() context.Context {
	return context.Background()
}
