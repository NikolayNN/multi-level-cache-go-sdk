package cache

// CacheId represents an identifier of an entry in a named cache.
type CacheId struct {
	CacheName string `json:"c"`
	Key       string `json:"k"`
}

// CacheEntry represents a value stored in a named cache.
type CacheEntry[T any] struct {
	CacheName string `json:"c"`
	Key       string `json:"k"`
	Value     T      `json:"v"`
}

// CacheEntryHit represents the result of a cache fetch operation.
type CacheEntryHit[T any] struct {
	CacheName string `json:"c"`
	Key       string `json:"k"`
	Value     T      `json:"v"`
	Found     bool   `json:"f"`
}
