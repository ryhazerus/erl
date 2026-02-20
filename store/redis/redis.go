package redis

import (
	"context"
	"fmt"
	"strconv"

	"github.com/redis/go-redis/v9"
	"github.com/ryhazerus/erl/store"
)

// Compile-time interface check.
var _ store.Store = (*RedisStore)(nil)

// RedisStore is a Store backed by Redis. Each rate limit key is stored as a
// Redis hash with fields "count" and "bucket_key". A TTL equal to the window
// duration is set on each key for automatic expiry.
type RedisStore struct {
	client *redis.Client
}

// NewRedisStore creates a new Redis-backed store.
func NewRedisStore(client *redis.Client) *RedisStore {
	return &RedisStore{client: client}
}

// incrementScript atomically increments a counter, resetting it when the
// bucket key changes. Returns the new count.
//
// KEYS[1] = counter key
// ARGV[1] = bucket_key
// ARGV[2] = window duration in seconds (for TTL)
var incrementScript = redis.NewScript(`
local key = KEYS[1]
local bucket_key = ARGV[1]
local ttl = tonumber(ARGV[2])

local current_bucket = redis.call("HGET", key, "bucket_key")
if current_bucket ~= bucket_key then
    redis.call("HSET", key, "count", "1", "bucket_key", bucket_key)
    if ttl > 0 then
        redis.call("EXPIRE", key, ttl)
    end
    return 1
end

local count = redis.call("HINCRBY", key, "count", 1)
return count
`)

// Increment atomically increments the counter for the given key in the current
// window bucket. If the bucket has rolled over, the counter resets.
func (r *RedisStore) Increment(ctx context.Context, key string, w store.Window) (int64, error) {
	ttl := int64(w.Duration.Seconds())
	result, err := incrementScript.Run(ctx, r.client, []string{redisKey(key)}, w.BucketKey, ttl).Int64()
	if err != nil {
		return 0, fmt.Errorf("erl/store/redis: increment: %w", err)
	}
	return result, nil
}

// Get returns the current counter value for key in the active window bucket.
func (r *RedisStore) Get(ctx context.Context, key string, w store.Window) (int64, error) {
	vals, err := r.client.HGetAll(ctx, redisKey(key)).Result()
	if err != nil {
		return 0, fmt.Errorf("erl/store/redis: get: %w", err)
	}

	if len(vals) == 0 {
		return 0, nil
	}

	if vals["bucket_key"] != w.BucketKey {
		return 0, nil
	}

	count, err := strconv.ParseInt(vals["count"], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("erl/store/redis: parse count: %w", err)
	}

	return count, nil
}

// Reset removes the counter for the given key.
func (r *RedisStore) Reset(ctx context.Context, key string) error {
	return r.client.Del(ctx, redisKey(key)).Err()
}

// Close closes the underlying Redis client.
func (r *RedisStore) Close() error {
	return r.client.Close()
}

func redisKey(key string) string {
	return "erl:" + key
}
