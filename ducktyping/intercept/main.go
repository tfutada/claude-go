// Package main demonstrates intercepting third-party code
// to add caching without modifying the original implementation.
package main

import (
	"fmt"
	"time"
)

// === "Third-party" code (imagine you can't modify this) ===

type SlowAPI struct{}

func (a *SlowAPI) GetUser(id int) (string, error) {
	// Simulates slow API call
	fmt.Printf("    [API] Fetching user %d from server...\n", id)
	time.Sleep(500 * time.Millisecond)
	return fmt.Sprintf("User-%d", id), nil
}

func (a *SlowAPI) GetProduct(id int) (string, error) {
	fmt.Printf("    [API] Fetching product %d from server...\n", id)
	time.Sleep(500 * time.Millisecond)
	return fmt.Sprintf("Product-%d", id), nil
}

// === Your code: Define interface for what you need ===

type UserFetcher interface {
	GetUser(id int) (string, error)
}

// === Caching interceptor ===

type CachedUserFetcher struct {
	wrapped UserFetcher        // the original implementation
	cache   map[int]string     // simple cache
	hits    int
	misses  int
}

func NewCachedUserFetcher(wrapped UserFetcher) *CachedUserFetcher {
	return &CachedUserFetcher{
		wrapped: wrapped,
		cache:   make(map[int]string),
	}
}

func (c *CachedUserFetcher) GetUser(id int) (string, error) {
	// Check cache first
	if user, ok := c.cache[id]; ok {
		c.hits++
		fmt.Printf("    [CACHE] Hit for user %d\n", id)
		return user, nil
	}

	// Cache miss - call original
	c.misses++
	fmt.Printf("    [CACHE] Miss for user %d\n", id)
	user, err := c.wrapped.GetUser(id)
	if err != nil {
		return "", err
	}

	// Store in cache
	c.cache[id] = user
	return user, nil
}

func (c *CachedUserFetcher) Stats() (hits, misses int) {
	return c.hits, c.misses
}

// === Another interceptor: Logging ===

type LoggedUserFetcher struct {
	wrapped UserFetcher
}

func (l *LoggedUserFetcher) GetUser(id int) (string, error) {
	start := time.Now()
	user, err := l.wrapped.GetUser(id)
	duration := time.Since(start)

	if err != nil {
		fmt.Printf("    [LOG] GetUser(%d) failed: %v [%v]\n", id, err, duration)
	} else {
		fmt.Printf("    [LOG] GetUser(%d) = %s [%v]\n", id, user, duration)
	}
	return user, err
}

// === Business logic that uses the interface ===

type UserService struct {
	fetcher UserFetcher
}

func (s *UserService) GetUserNames(ids []int) []string {
	var names []string
	for _, id := range ids {
		name, err := s.fetcher.GetUser(id)
		if err == nil {
			names = append(names, name)
		}
	}
	return names
}

func main() {
	fmt.Println("=== Intercepting for Caching Demo ===\n")

	// 1. Without caching (slow)
	fmt.Println("1. Without caching:")
	api := &SlowAPI{}
	service1 := &UserService{fetcher: api}

	start := time.Now()
	service1.GetUserNames([]int{1, 2, 1, 2}) // duplicate IDs
	fmt.Printf("   Total time: %v\n", time.Since(start))

	// 2. With caching (fast for repeated calls)
	fmt.Println("\n2. With caching interceptor:")
	cachedAPI := NewCachedUserFetcher(&SlowAPI{})
	service2 := &UserService{fetcher: cachedAPI}

	start = time.Now()
	service2.GetUserNames([]int{1, 2, 1, 2}) // same IDs
	fmt.Printf("   Total time: %v\n", time.Since(start))

	hits, misses := cachedAPI.Stats()
	fmt.Printf("   Cache hits: %d, misses: %d\n", hits, misses)

	// 3. Stacking interceptors: Cache + Logging
	fmt.Println("\n3. Stacked interceptors (Logging -> Cache -> API):")
	stackedAPI := &LoggedUserFetcher{
		wrapped: NewCachedUserFetcher(&SlowAPI{}),
	}
	service3 := &UserService{fetcher: stackedAPI}
	service3.GetUserNames([]int{1, 1, 1})

	// 4. The pattern
	fmt.Println("\n=== The Interception Pattern ===")
	fmt.Println("```go")
	fmt.Println("// 1. Define interface for behavior you need")
	fmt.Println("type UserFetcher interface {")
	fmt.Println("    GetUser(id int) (string, error)")
	fmt.Println("}")
	fmt.Println("")
	fmt.Println("// 2. Wrap original implementation")
	fmt.Println("type CachedUserFetcher struct {")
	fmt.Println("    wrapped UserFetcher  // holds original")
	fmt.Println("    cache   map[int]string")
	fmt.Println("}")
	fmt.Println("")
	fmt.Println("// 3. Implement same interface")
	fmt.Println("func (c *CachedUserFetcher) GetUser(id int) (string, error) {")
	fmt.Println("    if cached, ok := c.cache[id]; ok {")
	fmt.Println("        return cached, nil  // return from cache")
	fmt.Println("    }")
	fmt.Println("    result, err := c.wrapped.GetUser(id)  // call original")
	fmt.Println("    c.cache[id] = result  // store in cache")
	fmt.Println("    return result, err")
	fmt.Println("}")
	fmt.Println("```")

	fmt.Println("\n=== Key Points ===")
	fmt.Println("- SlowAPI never modified")
	fmt.Println("- SlowAPI doesn't know about UserFetcher interface")
	fmt.Println("- Interceptors can be stacked (decorator pattern)")
	fmt.Println("- Business logic (UserService) unchanged")
	fmt.Println("- Easy to add: retry, metrics, circuit breaker, etc.")
}
