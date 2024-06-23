package inmemeorycache

type CacheItem[T any] struct {
	Item       T
	Expiration uint32
}

type InMemoryCache[T any] struct {
	Data            map[string]CacheItem[T]
	CleanupInterval uint16
}

func NewInMemoryCache[T any](initialCapacity, cleanupInterval uint16) *InMemoryCache[T] {
	return &InMemoryCache[T]{
		Data:            make(map[string]CacheItem[T], initialCapacity),
		CleanupInterval: cleanupInterval,
	}
}

func (c *InMemoryCache[T]) Set(key string, value T, expirationSecond uint32) {
	item := CacheItem[T]{
		Item:       value,
		Expiration: expirationSecond,
	}
	c.Data[key] = item
}
