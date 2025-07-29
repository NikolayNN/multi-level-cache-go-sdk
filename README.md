# AUR Cache Client (Go)

This module provides a Go client for interacting with the
[multi-level cache service](https://github.com/NikolayNN/multi-level-cache-service).
It mirrors the functionality of the Java SDK.

## Building and Testing

Run the unit tests with:

```bash
go test ./...
```

## Example

```go
client := cache.New("http://localhost:8080")
entries := []cache.CacheEntry[any]{
    {CacheName: "users", Key: "1", Value: User{ID: 1, Name: "Alice"}},
}
err := client.PutAll(context.Background(), entries)
```
