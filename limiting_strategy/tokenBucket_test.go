package ratelimiter

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/Mayank-032/rate-bastion/cache"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestIsRequestAllowed_TokenBucket(t *testing.T) {
	mockCache := new(cache.MockCache)
	cache.CacheInstance = mockCache

	rateLimiter := NewTokenBucketRateLimiter(1, 2) // 1 request allowed in 2 seconds

	// Case-1: fail
	t.Run("Redis GetUser failure", func(t *testing.T) {
		mockCache.On("Get", "user1").Return("", errors.New("invalid response from redis"))

		allowed, err := rateLimiter.IsRequestAllowed("user1")
		assert.Error(t, err)
		assert.False(t, allowed)
		mockCache.AssertExpectations(t)
	})

	// Case-2: fail
	t.Run("Redis SetUser failure", func(t *testing.T) {
		mockCache.On("Get", "user2").Return("", errors.New("invalid key"))
		mockCache.On("Set", "user2", mock.Anything).Return(errors.New("invalid response from redis"))

		allowed, err := rateLimiter.IsRequestAllowed("user2")
		assert.Error(t, err)
		assert.False(t, allowed)
		mockCache.AssertExpectations(t)
	})

	// Case-3: success
	t.Run("User not in cache, first request is allowed", func(t *testing.T) {
		mockCache.On("Get", "user3").Return("", errors.New("invalid key"))
		mockCache.On("Set", "user3", mock.Anything).Return(nil)

		allowed, err := rateLimiter.IsRequestAllowed("user3")
		assert.NoError(t, err)
		assert.True(t, allowed)
		mockCache.AssertExpectations(t)
	})

	// Case-4: success
	t.Run("User in cache, request allowed", func(t *testing.T) {
		user := userBucket{TokensInBucket: 1, LastRefillTime: time.Now()}
		userBytes, _ := json.Marshal(user)

		mockCache.On("Get", "user4").Return(string(userBytes), nil)
		mockCache.On("Set", "user4", mock.Anything).Return(nil)

		allowed, err := rateLimiter.IsRequestAllowed("user4")
		assert.NoError(t, err)
		assert.True(t, allowed)
		mockCache.AssertExpectations(t)
	})

	// Case-5: fail
	t.Run("User in cache, no tokens left", func(t *testing.T) {
		user := userBucket{TokensInBucket: 0, LastRefillTime: time.Now()}
		userBytes, _ := json.Marshal(user)

		mockCache.On("Get", "user5").Return(string(userBytes), nil)
		mockCache.On("Set", "user5", mock.Anything).Return(nil)

		allowed, err := rateLimiter.IsRequestAllowed("user5")
		assert.NoError(t, err)
		assert.False(t, allowed)
		mockCache.AssertExpectations(t)
	})

	// Case-6: success
	t.Run("User in cache, tokens refilled", func(t *testing.T) {
		user := userBucket{TokensInBucket: 0, LastRefillTime: time.Now().Add(-(2 * time.Second))}
		userBytes, _ := json.Marshal(user)

		mockCache.On("Get", "user6").Return(string(userBytes), nil)
		mockCache.On("Set", "user6", mock.Anything).Return(nil)

		allowed, err := rateLimiter.IsRequestAllowed("user6")
		assert.NoError(t, err)
		assert.True(t, allowed)
		mockCache.AssertExpectations(t)
	})
}
