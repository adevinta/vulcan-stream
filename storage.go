/*
Copyright 2021 Adevinta
*/

package stream

import (
	"context"
	"fmt"
	"sync"
	"time"

	redis "github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
)

const (
	checksKeyPrefix = "check:"
	defTTL          = 7 * 24 // 7 days (hours)
	rfshPeriod      = 1 * 24 // 1 day (hours)
	defScanChunk    = 50     // elements per redis SCAN chunk
)

// RemoteDB represents interface to
// interact with remote DB.
type RemoteDB interface {
	GetChecks(ctx context.Context) ([]string, error)
	SetChecks(ctx context.Context, checks []string) error
}

// RedisConfig specifies the required
// config for RedisStorage.
type RedisConfig struct {
	Host string
	Port int
	Usr  string
	Pwd  string
	DB   int
	TTL  int
}

// RedisDB is the implementation of
// a RemoteDB for a Redis database.
type RedisDB struct {
	rdb *redis.Client
	ttl time.Duration
}

// NewRedisDB builds a new redis DB connector.
func NewRedisDB(c RedisConfig) *RedisDB {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprint(c.Host, ":", c.Port),
		Username: c.Usr,
		Password: c.Pwd,
		DB:       c.DB,
	})

	if c.TTL == 0 {
		c.TTL = defTTL
	}

	return &RedisDB{
		rdb: rdb,
		ttl: time.Duration(c.TTL) * time.Hour,
	}
}

// GetChecks returns checks stored in redis.
func (r *RedisDB) GetChecks(ctx context.Context) ([]string, error) {
	var (
		err    error
		cursor uint64
		checks []string
	)

	checks = []string{}
	match := fmt.Sprint(checksKeyPrefix, "*")
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

// SetChecks sets input checks in redis as a single transaction.
func (r *RedisDB) SetChecks(ctx context.Context, checks []string) error {
	pipe := r.rdb.TxPipeline()
	defer pipe.Close()

	for _, c := range checks {
		key := fmt.Sprint(checksKeyPrefix, c)
		err := pipe.Set(ctx, key, c, r.ttl).Err()
		if err != nil {
			pipe.Discard() // nolint
			return err
		}
	}

	_, err := pipe.Exec(ctx)
	return err
}

// Storage represents the stream storage
// for aborted checks.
type Storage interface {
	GetAbortedChecks(ctx context.Context) ([]string, error)
	AddAbortedChecks(ctx context.Context, checks []string) error
}

// cache is a local cache for
// the storage aborted checks.
type cache []string

type storage struct {
	sync.RWMutex
	db RemoteDB
	// Because a current constraint for stream is that
	// it runs as a single instance, we can mantain a local
	// cache in sync with remote storage to speed up retrivals.
	cache cache
	log   log.FieldLogger
}

// NewStorage builds a new Storage.
func NewStorage(db RemoteDB, logger log.FieldLogger) (Storage, error) {
	storage := &storage{
		db:    db,
		cache: cache{},
		log:   logger,
	}

	var err error
	storage.cache, err = storage.db.GetChecks(context.Background())
	if err != nil {
		return nil, fmt.Errorf("err retrieving remote checks: %w", err)
	}

	go storage.refresh()

	return storage, nil
}

// GetAbortedChecks returns the list of UUIDs for the currently aborted checks.
func (s *storage) GetAbortedChecks(ctx context.Context) ([]string, error) {
	s.RLock()
	defer s.RUnlock()

	// Because we have mantained local cache in sync
	// with remote storage, we can return local copy
	// directly instead of performing requests to redis.
	return s.cache, nil
}

// AddAbortedChecks adds the given checks to the current aborted checks list.
func (s *storage) AddAbortedChecks(ctx context.Context, checks []string) error {
	s.Lock()
	defer s.Unlock()

	// Because we have mantained local cache in sync
	// with remote storage, we can add new checks to
	// local cache and set that value in remote DB
	// instead of performing extra requests to retrieve
	// all remote values.
	err := s.db.SetChecks(ctx, checks)
	if err != nil {
		return err
	}
	s.cache = append(s.cache, checks...)

	return nil
}

// refresh refreshes the storage's local cache
// periodically so checks that have been expired
// remotely due to TTL, are also removed locally.
func (s *storage) refresh() {
	var err error
	ctx := context.Background()

	for {
		time.Sleep(time.Duration(rfshPeriod) * time.Hour)
		s.Lock()
		s.cache, err = s.db.GetChecks(ctx)
		if err != nil {
			s.log.Errorf("error refreshing remote checks: %v", err)
		}
		s.Unlock()
	}
}
