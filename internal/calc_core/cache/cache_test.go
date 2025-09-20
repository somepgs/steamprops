package cache

import (
	"testing"
	"time"
)

func TestCache_BasicOperations(t *testing.T) {
	cache := NewCache(100, 1*time.Minute)

	key := CacheKey{T: 200.0, P: 1e6, Region: "region1"}
	value := "test_value"

	// Test Set
	cache.Set(key, value)

	// Test Get
	retrieved, exists := cache.Get(key)
	if !exists {
		t.Fatal("Expected value to exist in cache")
	}

	if retrieved != value {
		t.Errorf("Expected %v, got %v", value, retrieved)
	}

	// Test non-existent key
	nonExistentKey := CacheKey{T: 300.0, P: 2e6, Region: "region2"}
	_, exists = cache.Get(nonExistentKey)
	if exists {
		t.Fatal("Expected value to not exist in cache")
	}
}

func TestCache_TTL(t *testing.T) {
	cache := NewCache(100, 100*time.Millisecond)

	key := CacheKey{T: 200.0, P: 1e6, Region: "region1"}
	value := "test_value"

	// Set value
	cache.Set(key, value)

	// Should exist immediately
	_, exists := cache.Get(key)
	if !exists {
		t.Fatal("Expected value to exist immediately after setting")
	}

	// Wait for TTL to expire
	time.Sleep(150 * time.Millisecond)

	// Should not exist after TTL
	_, exists = cache.Get(key)
	if exists {
		t.Fatal("Expected value to be expired after TTL")
	}
}

func TestCache_CustomTTL(t *testing.T) {
	cache := NewCache(100, 1*time.Minute)

	key := CacheKey{T: 200.0, P: 1e6, Region: "region1"}
	value := "test_value"

	// Set with custom TTL
	cache.SetWithTTL(key, value, 100*time.Millisecond)

	// Should exist immediately
	_, exists := cache.Get(key)
	if !exists {
		t.Fatal("Expected value to exist immediately after setting")
	}

	// Wait for custom TTL to expire
	time.Sleep(150 * time.Millisecond)

	// Should not exist after custom TTL
	_, exists = cache.Get(key)
	if exists {
		t.Fatal("Expected value to be expired after custom TTL")
	}
}

func TestCache_MaxSize(t *testing.T) {
	cache := NewCache(2, 1*time.Minute)

	// Add more entries than max size
	key1 := CacheKey{T: 200.0, P: 1e6, Region: "region1"}
	key2 := CacheKey{T: 300.0, P: 2e6, Region: "region2"}
	key3 := CacheKey{T: 400.0, P: 3e6, Region: "region3"}

	cache.Set(key1, "value1")
	cache.Set(key2, "value2")
	cache.Set(key3, "value3")

	// Cache should not exceed max size
	if cache.Size() > 2 {
		t.Errorf("Expected cache size <= 2, got %d", cache.Size())
	}

	// At least one of the first two entries should be evicted
	_, exists1 := cache.Get(key1)
	_, exists2 := cache.Get(key2)

	if exists1 && exists2 {
		t.Fatal("Expected at least one of the first two entries to be evicted")
	}
}

func TestCache_Clear(t *testing.T) {
	cache := NewCache(100, 1*time.Minute)

	key := CacheKey{T: 200.0, P: 1e6, Region: "region1"}
	value := "test_value"

	// Set value
	cache.Set(key, value)

	// Should exist
	_, exists := cache.Get(key)
	if !exists {
		t.Fatal("Expected value to exist before clear")
	}

	// Clear cache
	cache.Clear()

	// Should not exist after clear
	_, exists = cache.Get(key)
	if exists {
		t.Fatal("Expected value to not exist after clear")
	}

	// Size should be 0
	if cache.Size() != 0 {
		t.Errorf("Expected cache size 0 after clear, got %d", cache.Size())
	}
}

func TestPropertiesCache(t *testing.T) {
	pc := NewPropertiesCache(100)

	tCelsius := 200.0
	pPascal := 1e6
	region := "region1"
	properties := "test_properties"

	// Test Set
	pc.Set(tCelsius, pPascal, region, properties)

	// Test Get
	retrieved, exists := pc.Get(tCelsius, pPascal, region)
	if !exists {
		t.Fatal("Expected properties to exist in cache")
	}

	if retrieved != properties {
		t.Errorf("Expected %v, got %v", properties, retrieved)
	}
}

func TestTransportPropertiesCache(t *testing.T) {
	tpc := NewTransportPropertiesCache(100)

	Tkelvin := 473.15
	rho := 958.4
	propertyType := "viscosity"
	value := 0.282e-3

	// Test Set
	tpc.Set(Tkelvin, rho, propertyType, value)

	// Test Get
	retrieved, exists := tpc.Get(Tkelvin, rho, propertyType)
	if !exists {
		t.Fatal("Expected transport properties to exist in cache")
	}

	if retrieved != value {
		t.Errorf("Expected %v, got %v", value, retrieved)
	}
}

func TestCacheManager(t *testing.T) {
	cm := NewCacheManager(100, 100)

	// Test properties cache
	pc := cm.GetPropertiesCache()
	pc.Set(200.0, 1e6, "region1", "test_properties")

	retrieved, exists := pc.Get(200.0, 1e6, "region1")
	if !exists {
		t.Fatal("Expected properties to exist in cache manager")
	}

	if retrieved != "test_properties" {
		t.Errorf("Expected %v, got %v", "test_properties", retrieved)
	}

	// Test transport cache
	tc := cm.GetTransportCache()
	tc.Set(473.15, 958.4, "viscosity", 0.282e-3)

	retrieved, exists = tc.Get(473.15, 958.4, "viscosity")
	if !exists {
		t.Fatal("Expected transport properties to exist in cache manager")
	}

	if retrieved != 0.282e-3 {
		t.Errorf("Expected %v, got %v", 0.282e-3, retrieved)
	}

	// Test stats
	stats := cm.GetStats()
	if stats["properties_cache_size"] != 1 {
		t.Errorf("Expected properties cache size 1, got %d", stats["properties_cache_size"])
	}
	if stats["transport_cache_size"] != 1 {
		t.Errorf("Expected transport cache size 1, got %d", stats["transport_cache_size"])
	}

	// Test clear all
	cm.ClearAll()

	stats = cm.GetStats()
	if stats["properties_cache_size"] != 0 {
		t.Errorf("Expected properties cache size 0 after clear, got %d", stats["properties_cache_size"])
	}
	if stats["transport_cache_size"] != 0 {
		t.Errorf("Expected transport cache size 0 after clear, got %d", stats["transport_cache_size"])
	}
}

func TestCacheManager_Cleanup(t *testing.T) {
	cm := NewCacheManager(100, 100)

	// Start cleanup
	cm.StartCleanup()

	// Add some entries
	pc := cm.GetPropertiesCache()
	pc.Set(200.0, 1e6, "region1", "test_properties")

	// Wait a bit
	time.Sleep(100 * time.Millisecond)

	// Stop cleanup
	cm.StopCleanup()

	// Verify entries still exist (cleanup interval is 1 minute)
	_, exists := pc.Get(200.0, 1e6, "region1")
	if !exists {
		t.Fatal("Expected properties to still exist after cleanup stop")
	}
}

func TestCache_HashCollision(t *testing.T) {
	cache := NewCache(100, 1*time.Minute)

	// Create keys that might hash to the same value
	key1 := CacheKey{T: 200.0, P: 1e6, Region: "region1"}
	key2 := CacheKey{T: 200.0, P: 1e6, Region: "region2"}

	value1 := "value1"
	value2 := "value2"

	cache.Set(key1, value1)
	cache.Set(key2, value2)

	// Both should exist
	retrieved1, exists1 := cache.Get(key1)
	retrieved2, exists2 := cache.Get(key2)

	if !exists1 || !exists2 {
		t.Fatal("Expected both values to exist")
	}

	if retrieved1 != value1 || retrieved2 != value2 {
		t.Errorf("Expected %v and %v, got %v and %v", value1, value2, retrieved1, retrieved2)
	}
}
