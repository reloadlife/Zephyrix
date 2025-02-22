package zephyrix

// import (
// 	"context"
// 	"testing"
// 	"time"

// 	"github.com/latolukasz/beeorm/v3"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/mock"
// 	"golang.org/x/time/rate"
// )

// // MockRedisCache is a mock implementation of beeorm.RedisCache
// type MockRedisCache struct {
// 	mock.Mock
// }

// func (m *MockRedisCache) Get(orm beeorm.ORM, key string) (string, bool) {
// 	args := m.Called(orm, key)
// 	return args.String(0), args.Bool(1)
// }

// func (m *MockRedisCache) Set(orm beeorm.ORM, key string, value string, expiration time.Duration) {
// 	m.Called(orm, key, value, expiration)
// }

// func (m *MockRedisCache) BLMove(orm beeorm.ORM, source, destination string, wherefrom, whereto string, timeout time.Duration) (string, error) {
// 	args := m.Called(orm, source, destination, wherefrom, whereto, timeout)
// 	return args.String(0), args.Error(1)
// }

// // Add other methods required by beeorm.RedisCache interface
// func (m *MockRedisCache) Del(orm beeorm.ORM, keys ...string) int64 {
// 	args := m.Called(orm, keys)
// 	return args.Get(0).(int64)
// }

// func (m *MockRedisCache) SetNX(orm beeorm.ORM, key string, value string, expiration time.Duration) bool {
// 	args := m.Called(orm, key, value, expiration)
// 	return args.Bool(0)
// }

// // Add more methods as needed...

// // MockORM is a mock implementation of beeorm.ORM
// type MockORM struct {
// 	mock.Mock
// }

// // MockEngine is a mock implementation of beeorm.Engine
// type MockEngine struct {
// 	mock.Mock
// }

// func (m *MockEngine) Redis(pool string) beeorm.RedisCache {
// 	args := m.Called(pool)
// 	return args.Get(0).(beeorm.RedisCache)
// }

// func (m *MockEngine) NewORM(ctx context.Context) beeorm.ORM {
// 	args := m.Called(ctx)
// 	return args.Get(0).(beeorm.ORM)
// }

// func TestNewRateLimiter(t *testing.T) {
// 	mockEngine := new(MockEngine)
// 	mockRedis := new(MockRedisCache)
// 	mockORM := new(MockORM)

// 	config := &Config{
// 		RateLimiter: RateLimiterConfig{
// 			RedisPool: "default",
// 			LimitPools: []RateLimiterPool{
// 				{Name: "default", Limit: 10, Burst: 20, ExpireTime: time.Minute},
// 			},
// 		},
// 	}

// 	mockEngine.On("Redis", "default").Return(mockRedis)
// 	mockEngine.On("NewORM", mock.Anything).Return(mockORM)

// 	rl := NewRateLimiter(mockEngine, config)

// 	assert.NotNil(t, rl)
// 	assert.Equal(t, config.RateLimiter, rl.config)
// 	assert.Equal(t, mockRedis, rl.client)
// 	assert.Equal(t, mockORM, rl.orm)

// 	mockEngine.AssertExpectations(t)
// }

// func TestGetPoolConfig(t *testing.T) {
// 	rl := &RateLimiter{
// 		config: RateLimiterConfig{
// 			LimitPools: []RateLimiterPool{
// 				{Name: "default", Limit: 10, Burst: 20, ExpireTime: time.Minute},
// 				{Name: "api", Limit: 5, Burst: 10, ExpireTime: time.Minute * 5},
// 			},
// 		},
// 	}

// 	tests := []struct {
// 		name     string
// 		poolName string
// 		expected *RateLimiterPool
// 	}{
// 		{"Default pool", "default", &rl.config.LimitPools[0]},
// 		{"API pool", "api", &rl.config.LimitPools[1]},
// 		{"Non-existent pool", "nonexistent", nil},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			result := rl.getPoolConfig(tt.poolName)
// 			assert.Equal(t, tt.expected, result)
// 		})
// 	}
// }

// func TestLimiter(t *testing.T) {
// 	rl := &RateLimiter{
// 		config: RateLimiterConfig{
// 			LimitPools: []RateLimiterPool{
// 				{Name: "default", Limit: 10, Burst: 20, ExpireTime: time.Minute},
// 				{Name: "api", Limit: 5, Burst: 10, ExpireTime: time.Minute * 5},
// 			},
// 		},
// 	}

// 	tests := []struct {
// 		name     string
// 		poolName string
// 		expected *RateLimiterPool
// 	}{
// 		{"Default pool", "", &rl.config.LimitPools[0]},
// 		{"API pool", "api", &rl.config.LimitPools[1]},
// 		{"Non-existent pool", "nonexistent", &rl.config.LimitPools[0]},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			ctx := context.Background()
// 			limiter := rl.Limiter(ctx, tt.poolName)
// 			assert.NotNil(t, limiter)
// 			assert.Equal(t, tt.expected.Burst, limiter.Burst)
// 			assert.Equal(t, tt.expected.ExpireTime, limiter.ExpireTime)
// 		})
// 	}
// }

// func TestAllow(t *testing.T) {
// 	mockRedis := new(MockRedisCache)
// 	rl := &RateLimiter{
// 		client: mockRedis,
// 		orm:    new(MockORM),
// 	}

// 	limiter := &Limiter{
// 		rl:         rl,
// 		limiter:    rate.NewLimiter(10, 20),
// 		Burst:      20,
// 		ExpireTime: time.Minute,
// 	}

// 	tests := []struct {
// 		name           string
// 		action         string
// 		key            string
// 		redisCount     string
// 		expectedResult bool
// 	}{
// 		{"Allow local", "login", "127.0.0.1", "0", true},
// 		{"Deny Redis", "login", "127.0.0.1", "20", false},
// 		{"Allow Redis", "login", "127.0.0.1", "10", true},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			ctx := context.Background()
// 			mockRedis.On("Get", rl.orm, mock.Anything).Return(tt.redisCount, true).Once()
// 			if tt.expectedResult && tt.redisCount != "20" {
// 				mockRedis.On("Set", rl.orm, mock.Anything, mock.Anything, limiter.ExpireTime).Once()
// 			}

// 			result := limiter.Allow(ctx, tt.action, tt.key)
// 			assert.Equal(t, tt.expectedResult, result)

// 			mockRedis.AssertExpectations(t)
// 		})
// 	}
// }

// func TestAllowLocal(t *testing.T) {
// 	limiter := &Limiter{
// 		limiter: rate.NewLimiter(10, 20),
// 	}

// 	for i := 0; i < 20; i++ {
// 		assert.True(t, limiter.allowLocal("test"))
// 	}

// 	assert.False(t, limiter.allowLocal("test"))
// }

// func TestAllowRedis(t *testing.T) {
// 	mockRedis := new(MockRedisCache)
// 	rl := &RateLimiter{
// 		client: mockRedis,
// 		orm:    new(MockORM),
// 	}

// 	limiter := &Limiter{
// 		rl:         rl,
// 		Burst:      20,
// 		ExpireTime: time.Minute,
// 	}

// 	tests := []struct {
// 		name           string
// 		redisCount     string
// 		expectedResult bool
// 	}{
// 		{"Allow", "10", true},
// 		{"Deny", "20", false},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			ctx := context.Background()
// 			mockRedis.On("Get", rl.orm, mock.Anything).Return(tt.redisCount, true).Once()
// 			if tt.expectedResult {
// 				mockRedis.On("Set", rl.orm, mock.Anything, mock.Anything, limiter.ExpireTime).Once()
// 			}

// 			result := limiter.allowRedis(ctx, "test")
// 			assert.Equal(t, tt.expectedResult, result)

// 			mockRedis.AssertExpectations(t)
// 		})
// 	}
// }
