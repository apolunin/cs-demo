package storage

import (
  "sync"

  "github.com/apolunin/orderedmap"
)

type ItemStorage struct {
  mu   sync.RWMutex
  data *orderedmap.OrderedMap[string, string]
}

func New() *ItemStorage {
  return &ItemStorage{
    data: orderedmap.New[string, string](),
  }
}

type Item struct {
  Key, Value string
}

func (s *ItemStorage) GetItem(key string) (string, bool) {
  s.mu.RLock()
  defer s.mu.RUnlock()

  return s.data.Get(key)
}

func (s *ItemStorage) GetAllItems() []*Item {
  s.mu.RLock()
  defer s.mu.RUnlock()

  res := make([]*Item, 0)

  next := s.data.Iterator()
  for k, v, ok := next(); ok; k, v, ok = next() {
    res = append(res, &Item{Key: k, Value: v})
  }

  return res
}

func (s *ItemStorage) AddItem(key, value string) (string, bool) {
  s.mu.Lock()
  defer s.mu.Unlock()

  return s.data.Set(key, value)
}

func (s *ItemStorage) RemoveItem(key string) (string, bool) {
  s.mu.Lock()
  defer s.mu.Unlock()

  return s.data.Delete(key)
}
