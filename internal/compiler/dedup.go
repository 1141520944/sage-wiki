package compiler

import (
	"sync"

	"github.com/xoai/sage-wiki/internal/embed"
	"github.com/xoai/sage-wiki/internal/log"
	"github.com/xoai/sage-wiki/internal/vectors"
)

const (
	// maxDedupCacheSize caps the cache to prevent unbounded memory growth.
	// At 384-dim embeddings, 50K entries ≈ 75MB. Beyond this, dedup
	// effectiveness plateaus — the most common duplicates are caught early.
	maxDedupCacheSize = 50000
)

// DedupCache uses embedding cosine similarity to detect duplicate concepts.
// Before creating a new concept, embed its name and check against existing
// concept embeddings. If similarity > threshold, merge as alias.
//
// DedupCache is safe for concurrent use.
type DedupCache struct {
	embedder  embed.Embedder
	vecStore  *vectors.Store // optional — load pre-stored embeddings
	threshold float64        // cosine similarity threshold (default 0.85)

	mu    sync.RWMutex // protects cache
	cache map[string][]float32
}

// NewDedupCache 会根据给定的嵌入器和阈值创建一个去重缓存。
// 如果阈值小于或等于 0，则默认值为 0.85。vecStore 是可选的——如果提供了该参数，
// 则 Seed 将从存储中加载现有的嵌入，而不会重新进行嵌入操作。
func NewDedupCache(embedder embed.Embedder, vecStore *vectors.Store, threshold float64) *DedupCache {
	if threshold <= 0 {
		threshold = 0.85
	}
	return &DedupCache{
		embedder:  embedder,
		vecStore:  vecStore,
		threshold: threshold,
		cache:     make(map[string][]float32),
	}
}

// Seed 会将缓存中填充现有的概念名称。
// 首先尝试从向量存储中加载嵌入（每个概念的加载时间均为 O(1)，无需调用 API）。对于存储中未包含的概念，则通过 API 进行嵌入计算。
// 会限制缓存大小至 maxDedupCacheSize，以防止内存无限增长。
func (dc *DedupCache) Seed(names []string) {
	if dc.embedder == nil {
		return
	}

	if len(names) > maxDedupCacheSize {
		log.Warn("dedup cache: capping seed to max size",
			"requested", len(names), "max", maxDedupCacheSize)
		names = names[:maxDedupCacheSize]
	}

	loaded, embedded, failed := 0, 0, 0

	dc.mu.Lock()
	for _, name := range names {
		// Try vector store first (no API call needed)
		if dc.vecStore != nil {
			if vec, err := dc.vecStore.Get(name); err == nil && vec != nil {
				dc.cache[name] = vec
				loaded++
				continue
			}
		}

		// Fall back to embedding API
		vec, err := dc.embedder.Embed(name)
		if err != nil {
			failed++
			continue
		}
		dc.cache[name] = vec
		embedded++
	}
	cacheSize := len(dc.cache)
	dc.mu.Unlock()

	if cacheSize > 0 {
		log.Info("dedup cache seeded",
			"total", cacheSize, "from_store", loaded,
			"embedded", embedded, "failed", failed)
	}
	if failed > 0 {
		log.Warn("dedup cache: some concepts could not be seeded", "failed", failed)
	}
}

// CheckDuplicate “检查重复”函数用于检查一个概念名称是否与现有概念的名称相同。
// 该函数会返回现有概念的名称、相似度得分以及嵌入向量。
// 返回的嵌入向量可供调用者使用，以便将其传递给“添加并使用向量”函数，从而避免重复嵌入操作。
func (dc *DedupCache) CheckDuplicate(name string) (match string, score float64, vec []float32) {
	dc.mu.RLock()
	cacheEmpty := len(dc.cache) == 0
	dc.mu.RUnlock()

	if dc.embedder == nil || cacheEmpty {
		return "", 0, nil
	}

	var err error
	vec, err = dc.embedder.Embed(name)
	if err != nil {
		return "", 0, nil
	}

	bestMatch := ""
	bestScore := 0.0

	dc.mu.RLock()
	for existing, existingVec := range dc.cache {
		if existing == name {
			continue
		}
		// Guard against dimension mismatch (e.g., provider changed)
		if len(vec) != len(existingVec) {
			continue
		}
		s := vectors.CosineSimilarity(vec, existingVec)
		if s > bestScore {
			bestScore = s
			bestMatch = existing
		}
	}
	dc.mu.RUnlock()

	if bestScore >= dc.threshold {
		return bestMatch, bestScore, vec
	}

	return "", 0, vec
}

// AddWithVec registers a new concept with a pre-computed embedding.
// Use the vec returned from CheckDuplicate to avoid double-embedding.
func (dc *DedupCache) AddWithVec(name string, vec []float32) {
	dc.mu.Lock()
	defer dc.mu.Unlock()
	if len(dc.cache) >= maxDedupCacheSize {
		return // at capacity
	}
	if _, exists := dc.cache[name]; exists {
		return
	}
	dc.cache[name] = vec
}

// Add registers a new concept in the cache by embedding its name.
// Prefer AddWithVec when the embedding is already available.
func (dc *DedupCache) Add(name string) {
	if dc.embedder == nil {
		return
	}
	dc.mu.RLock()
	atCap := len(dc.cache) >= maxDedupCacheSize
	_, exists := dc.cache[name]
	dc.mu.RUnlock()
	if atCap || exists {
		return
	}
	vec, err := dc.embedder.Embed(name)
	if err != nil {
		return
	}
	dc.mu.Lock()
	// Re-check under write lock
	if len(dc.cache) < maxDedupCacheSize {
		if _, exists := dc.cache[name]; !exists {
			dc.cache[name] = vec
		}
	}
	dc.mu.Unlock()
}

// Size returns the number of concepts in the cache.
func (dc *DedupCache) Size() int {
	dc.mu.RLock()
	defer dc.mu.RUnlock()
	return len(dc.cache)
}
