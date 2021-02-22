package stream

import (
	"context"
	"fmt"
	"sync"
	"time"

	redis "github.com/go-redis/redis/v8"
)

const (
	checksKeyPrefix = "check:"
	defTTL          = 7 * 24 // 7 days (hours)
	rfshPeriod      = 1 * 24 // 1 day (hours)
	defScanChunk    = 50     // elements per redis SCAN chunk
)

// Storage represents the stream storage
// for aborted checks.
type Storage interface {
	GetAbortedChecks(ctx context.Context) ([]string, error)
	AddAbortedChecks(ctx context.Context, checks []string) error
}

// cache is a local cache for
// the storage aborted checks.
type cache []string

// RedisStorage is the Redis implementation
// for the stream Storage.
type RedisStorage struct {
	*sync.RWMutex
	rdb *redis.Client
	// Because a current condition for stream is that
	// it runs as a single instance, we can mantain a local
	// cache in sync with remote storage to speed up retrivals.
	cache cache
	ttl   time.Duration
}

// NewRedisStorage builds a new RedisStorage.
func NewRedisStorage(addr, pwd string, db, ttl int) (*RedisStorage, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pwd,
		DB:       db,
	})

	if ttl == 0 {
		ttl = defTTL
	}

	storage := &RedisStorage{
		rdb:   rdb,
		cache: cache{},
		ttl:   time.Duration(ttl) * time.Hour,
	}

	var err error
	storage.cache, err = storage.getRemoteChecks(context.Background())
	if err != nil {
		return nil, fmt.Errorf("Err retrieving checks: %v", err)
	}

	go storage.refresh()

	return storage, nil
}

// GetAbortedChecks returns the list of UUIDs for the currently aborted checks.
func (r *RedisStorage) GetAbortedChecks(ctx context.Context) ([]string, error) {
	r.RLock()
	defer r.RUnlock()

	// Because we have mantained local cache in sync
	// with remote storage, we can return local copy
	// directly instead of performing a request to redis.
	return r.cache, nil
}

// AddAbortedChecks adds the given checks to the current aborted checks list.
func (r *RedisStorage) AddAbortedChecks(ctx context.Context, checks []string) error {
	r.Lock()
	defer r.Unlock()

	// Because we have mantained local cache in sync
	// with remote storage, we can add new checks to
	// local cache and set that value in redis instead
	// of performing extra requests to retrieve all
	// remote values.
	for i, c := range checks {
		key := fmt.Sprint(checksKeyPrefix, c)
		err := r.rdb.Set(ctx, key, c, r.ttl).Err()
		if err != nil {
			return fmt.Errorf("%d checks could not be aborted: %v", len(checks)-i, err)
		}
		r.cache = append(r.cache, c)
	}

	return nil
}

// getRemoteChecks retrieves the checks stored in redis.
// Because we want to set a TTL per check, we had to store one entry per check,
// so to retrieve them we have to perform a SCAN matching 'check:*' and retrieve
// the value per each obtained key.
// SCAN is performed in chunks so we don't block redis when there are many entries.
func (r *RedisStorage) getRemoteChecks(ctx context.Context) ([]string, error) {
	r.RLock()
	defer r.RUnlock()

	var (
		err    error
		cursor uint64
		checks []string
	)

	match := fmt.Sprint(checksKeyPrefix, '*')
	for {
		var keys []string
		keys, cursor, err = r.rdb.Scan(ctx, cursor, match, defScanChunk).Result()
		if err != nil {
			return nil, err
		}
		for _, k := range keys {
			checkID, err := r.rdb.Get(ctx, k).Result()
			if err != nil {
				return nil, err
			}
			checks = append(checks, checkID)
		}
		if cursor == 0 {
			break
		}
	}

	return checks, nil
}

// refresh refreshes the redis storage local cache
// periodically so checks that have been expired
// remotely due to TTL, are also removed locally.
func (r *RedisStorage) refresh() {
	ctx := context.Background()
	for {
		time.Sleep(rfshPeriod)
		r.Lock()
		r.cache, _ = r.getRemoteChecks(ctx)
		r.Unlock()
	}
}
