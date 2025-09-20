package cache

import (
	"hash/fnv"
	"sync"
	"time"
)

// CacheKey represents a cache key
type CacheKey struct {
	T      float64
	P      float64
	Region string
}

// CacheEntry represents a cached entry
type CacheEntry struct {
	Value     interface{}
	Timestamp time.Time
	TTL       time.Duration
}

// Cache provides thread-safe caching with TTL
type Cache struct {
	entries    map[uint64]*CacheEntry
	mutex      sync.RWMutex
	maxSize    int
	defaultTTL time.Duration
}

// NewCache creates a new cache instance
func NewCache(maxSize int, defaultTTL time.Duration) *Cache {
	return &Cache{
		entries:    make(map[uint64]*CacheEntry),
		maxSize:    maxSize,
		defaultTTL: defaultTTL,
	}
}

// hashKey creates a hash for the cache key
func (c *Cache) hashKey(key CacheKey) uint64 {
	h := fnv.New64a()
	h.Write([]byte(key.Region))
	h.Write([]byte{0}) // separator
	h.Write([]byte(string(rune(key.T))))
	h.Write([]byte{0}) // separator
	h.Write([]byte(string(rune(key.P))))
	return h.Sum64()
}

// Get retrieves a value from cache
func (c *Cache) Get(key CacheKey) (interface{}, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	hash := c.hashKey(key)
	entry, exists := c.entries[hash]
	if !exists {
		return nil, false
	}

	// Check if entry has expired
	if time.Since(entry.Timestamp) > entry.TTL {
		// Entry expired, remove it
		c.mutex.RUnlock()
		c.mutex.Lock()
		delete(c.entries, hash)
		c.mutex.Unlock()
		c.mutex.RLock()
		return nil, false
	}

	return entry.Value, true
}

// Set stores a value in cache
func (c *Cache) Set(key CacheKey, value interface{}) {
	c.SetWithTTL(key, value, c.defaultTTL)
}

// SetWithTTL stores a value in cache with custom TTL
func (c *Cache) SetWithTTL(key CacheKey, value interface{}, ttl time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	hash := c.hashKey(key)

	// Check if we need to evict entries
	if len(c.entries) >= c.maxSize {
		c.evictOldest()
	}

	c.entries[hash] = &CacheEntry{
		Value:     value,
		Timestamp: time.Now(),
		TTL:       ttl,
	}
}

// evictOldest removes the oldest entry from cache
func (c *Cache) evictOldest() {
	var oldestHash uint64
	var oldestTime time.Time

	for hash, entry := range c.entries {
		if oldestTime.IsZero() || entry.Timestamp.Before(oldestTime) {
			oldestTime = entry.Timestamp
			oldestHash = hash
		}
	}

	if oldestHash != 0 {
		delete(c.entries, oldestHash)
	}
}

// Clear removes all entries from cache
func (c *Cache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.entries = make(map[uint64]*CacheEntry)
}

// Size returns the number of entries in cache
func (c *Cache) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return len(c.entries)
}

// Cleanup removes expired entries
func (c *Cache) Cleanup() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := time.Now()
	for hash, entry := range c.entries {
		if now.Sub(entry.Timestamp) > entry.TTL {
			delete(c.entries, hash)
		}
	}
}

// PropertiesCache is a specialized cache for thermodynamic properties
type PropertiesCache struct {
	cache *Cache
}

// NewPropertiesCache creates a new properties cache
func NewPropertiesCache(maxSize int) *PropertiesCache {
	return &PropertiesCache{
		cache: NewCache(maxSize, 5*time.Minute), // 5 minutes default TTL
	}
}

// Get retrieves properties from cache
func (pc *PropertiesCache) Get(tCelsius, pPascal float64, region string) (interface{}, bool) {
	key := CacheKey{T: tCelsius, P: pPascal, Region: region}
	return pc.cache.Get(key)
}

// Set stores properties in cache
func (pc *PropertiesCache) Set(tCelsius, pPascal float64, region string, properties interface{}) {
	key := CacheKey{T: tCelsius, P: pPascal, Region: region}
	pc.cache.Set(key, properties)
}

// Clear clears the cache
func (pc *PropertiesCache) Clear() {
	pc.cache.Clear()
}

// Size returns cache size
func (pc *PropertiesCache) Size() int {
	return pc.cache.Size()
}

// Cleanup removes expired entries
func (pc *PropertiesCache) Cleanup() {
	pc.cache.Cleanup()
}

// TransportPropertiesCache is a specialized cache for transport properties
type TransportPropertiesCache struct {
	cache *Cache
}

// NewTransportPropertiesCache creates a new transport properties cache
func NewTransportPropertiesCache(maxSize int) *TransportPropertiesCache {
	return &TransportPropertiesCache{
		cache: NewCache(maxSize, 10*time.Minute), // 10 minutes default TTL
	}
}

// Get retrieves transport properties from cache
func (tpc *TransportPropertiesCache) Get(Tkelvin, rho float64, propertyType string) (interface{}, bool) {
	key := CacheKey{T: Tkelvin, P: rho, Region: propertyType}
	return tpc.cache.Get(key)
}

// Set stores transport properties in cache
func (tpc *TransportPropertiesCache) Set(Tkelvin, rho float64, propertyType string, value interface{}) {
	key := CacheKey{T: Tkelvin, P: rho, Region: propertyType}
	tpc.cache.Set(key, value)
}

// Clear clears the cache
func (tpc *TransportPropertiesCache) Clear() {
	tpc.cache.Clear()
}

// Size returns cache size
func (tpc *TransportPropertiesCache) Size() int {
	return tpc.cache.Size()
}

// Cleanup removes expired entries
func (tpc *TransportPropertiesCache) Cleanup() {
	tpc.cache.Cleanup()
}

// CacheManager manages multiple caches
type CacheManager struct {
	propertiesCache *PropertiesCache
	transportCache  *TransportPropertiesCache
	cleanupInterval time.Duration
	stopCleanup     chan bool
	cleanupRunning  bool
	cleanupMutex    sync.Mutex
}

// NewCacheManager creates a new cache manager
func NewCacheManager(propertiesCacheSize, transportCacheSize int) *CacheManager {
	return &CacheManager{
		propertiesCache: NewPropertiesCache(propertiesCacheSize),
		transportCache:  NewTransportPropertiesCache(transportCacheSize),
		cleanupInterval: 1 * time.Minute,
		stopCleanup:     make(chan bool),
	}
}

// GetPropertiesCache returns the properties cache
func (cm *CacheManager) GetPropertiesCache() *PropertiesCache {
	return cm.propertiesCache
}

// GetTransportCache returns the transport properties cache
func (cm *CacheManager) GetTransportCache() *TransportPropertiesCache {
	return cm.transportCache
}

// StartCleanup starts the cleanup goroutine
func (cm *CacheManager) StartCleanup() {
	cm.cleanupMutex.Lock()
	defer cm.cleanupMutex.Unlock()

	if cm.cleanupRunning {
		return
	}

	cm.cleanupRunning = true
	go cm.cleanupLoop()
}

// StopCleanup stops the cleanup goroutine
func (cm *CacheManager) StopCleanup() {
	cm.cleanupMutex.Lock()
	defer cm.cleanupMutex.Unlock()

	if !cm.cleanupRunning {
		return
	}

	cm.stopCleanup <- true
	cm.cleanupRunning = false
}

// cleanupLoop runs the cleanup loop
func (cm *CacheManager) cleanupLoop() {
	ticker := time.NewTicker(cm.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			cm.propertiesCache.Cleanup()
			cm.transportCache.Cleanup()
		case <-cm.stopCleanup:
			return
		}
	}
}

// ClearAll clears all caches
func (cm *CacheManager) ClearAll() {
	cm.propertiesCache.Clear()
	cm.transportCache.Clear()
}

// GetStats returns cache statistics
func (cm *CacheManager) GetStats() map[string]int {
	return map[string]int{
		"properties_cache_size": cm.propertiesCache.Size(),
		"transport_cache_size":  cm.transportCache.Size(),
	}
}
