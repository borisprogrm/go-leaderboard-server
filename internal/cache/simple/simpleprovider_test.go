package cache_simple_provider

import (
	"context"
	cacheprovider "go-leaderboard-server/internal/cache"
	dbprovider "go-leaderboard-server/internal/db"
	"go-leaderboard-server/internal/utils"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockDbProvider struct {
	mock.Mock
}

func (m *MockDbProvider) Initialize(ctx context.Context, config dbprovider.IDBProviderConfig) error {
	args := m.Called(config)
	return args.Error(0)
}

func (m *MockDbProvider) Put(ctx context.Context, gameId string, userId string, userProp dbprovider.UserProperties) error {
	args := m.Called(gameId, userId, userProp)
	return args.Error(0)
}

func (m *MockDbProvider) Delete(ctx context.Context, gameId string, userId string) error {
	args := m.Called(gameId, userId)
	return args.Error(0)
}

func (m *MockDbProvider) Get(ctx context.Context, gameId string, userId string) (*dbprovider.UserProperties, error) {
	args := m.Called(gameId, userId)
	return args.Get(0).(*dbprovider.UserProperties), args.Error(1)
}

func (m *MockDbProvider) Top(ctx context.Context, gameId string, nTop uint32) (dbprovider.TopData, error) {
	args := m.Called(gameId, nTop)
	return args.Get(0).(dbprovider.TopData), args.Error(1)
}

func (m *MockDbProvider) Shutdown(ctx context.Context) error {
	args := m.Called()
	return args.Error(0)
}

const CACHE_TTL uint32 = 1000

func setupTest() (func() error, *CacheSimpleProvider, error) {
	var mockClock utils.IClock = &utils.MockClock{}
	cacheProvider := NewCacheSimpleProvider(&mockClock)
	err := cacheProvider.Initialize(context.Background(), &CacheSimpleProviderConfig{
		CacheProviderBaseConfig: cacheprovider.CacheProviderBaseConfig{
			IsDebug: true,
			Ttl:     CACHE_TTL,
		},
	}, &MockDbProvider{})

	return func() error {
		return cacheProvider.Shutdown(context.Background())
	}, cacheProvider, err
}

func runTest(t *testing.T, name string, testFunc utils.TestFcn[*CacheSimpleProvider]) {
	utils.RunTest(t, name, setupTest, testFunc)
}

func TestCacheSimpleProvider(t *testing.T) {
	gameId1 := "game1"
	gameId2 := "game2"

	runTest(t, "get empty top data", func(t *testing.T, cacheProvider *CacheSimpleProvider) {
		var (
			top            dbprovider.TopData
			err            error
			mockClock      = (*cacheProvider.clock).(*utils.MockClock)
			mockDbProvider = cacheProvider.dbprovider.(*MockDbProvider)
		)

		getTopAsync := func(gameId string, nTop uint32) (chan dbprovider.TopData, chan error) {
			c := make(chan dbprovider.TopData, 1)
			e := make(chan error, 1)
			go func() {
				top, err := cacheProvider.Top(context.Background(), gameId, nTop)
				c <- top
				e <- err
			}()
			return c, e
		}

		mockClock.SetTime(time.UnixMilli(0))

		mockDbProvider.On("Top", gameId1, uint32(10)).Return(dbprovider.TopData{}, nil)
		top, err = cacheProvider.Top(context.Background(), gameId1, 10)
		require.NoError(t, err)
		mockDbProvider.AssertNumberOfCalls(t, "Top", 1)
		require.Equal(t, dbprovider.TopData{}, top)

		mockDbProvider.On("Top", gameId2, uint32(10)).Return(dbprovider.TopData{}, nil)
		mockDbProvider.On("Top", gameId2, uint32(8)).Return(dbprovider.TopData{}, nil)
		_, _ = cacheProvider.Top(context.Background(), gameId2, 10)
		topCh1, errCh1 := getTopAsync(gameId2, 10)
		topCh2, errCh2 := getTopAsync(gameId2, 8)
		require.NoError(t, <-errCh1)
		require.NoError(t, <-errCh2)
		require.Equal(t, dbprovider.TopData{}, <-topCh1)
		require.Equal(t, dbprovider.TopData{}, <-topCh2)
		mockDbProvider.AssertNumberOfCalls(t, "Top", 2)

		mockDbProvider.On("Top", gameId2, uint32(20)).Return(dbprovider.TopData{}, nil)
		mockDbProvider.On("Top", gameId2, uint32(100)).Return(dbprovider.TopData{}, nil)
		top1, err1 := cacheProvider.Top(context.Background(), gameId2, 20)
		top2, err2 := cacheProvider.Top(context.Background(), gameId2, 100)
		require.NoError(t, err1)
		require.NoError(t, err2)
		require.Equal(t, dbprovider.TopData{}, top1)
		require.Equal(t, dbprovider.TopData{}, top2)
		mockDbProvider.AssertNumberOfCalls(t, "Top", 4)
	})

	runTest(t, "get top data", func(t *testing.T, cacheProvider *CacheSimpleProvider) {
		var (
			top            dbprovider.TopData
			err            error
			mockClock      = (*cacheProvider.clock).(*utils.MockClock)
			mockDbProvider = cacheProvider.dbprovider.(*MockDbProvider)
			mockCall       *mock.Call
		)

		userData1 := dbprovider.UserData{
			UserId: "user1",
			UserProperties: dbprovider.UserProperties{
				Name:   "Jack",
				Score:  84,
				Params: "some_payload_1",
			},
		}

		userData2 := dbprovider.UserData{
			UserId: "user2",
			UserProperties: dbprovider.UserProperties{
				Score:  52,
				Params: "some_payload_2",
			},
		}

		userData3 := dbprovider.UserData{
			UserId: "user3",
			UserProperties: dbprovider.UserProperties{
				Name:  "Tom",
				Score: 31,
			},
		}

		now := time.Now()
		mockClock.SetTime(now)

		mockCall = mockDbProvider.On("Top", gameId1, uint32(10)).Return(dbprovider.TopData{userData1, userData2, userData3}, nil)
		top, err = cacheProvider.Top(context.Background(), gameId1, 10)
		require.NoError(t, err)
		require.Equal(t, dbprovider.TopData{userData1, userData2, userData3}, top)
		mockDbProvider.AssertNumberOfCalls(t, "Top", 1)

		mockDbProvider.On("Top", gameId1, uint32(2)).Return(dbprovider.TopData{userData1, userData2}, nil)
		top, err = cacheProvider.Top(context.Background(), gameId1, 2)
		require.NoError(t, err)
		require.Equal(t, dbprovider.TopData{userData1, userData2}, top)
		mockDbProvider.AssertNumberOfCalls(t, "Top", 1)

		mockClock.SetTime(now.Add(time.Duration(CACHE_TTL-1) * time.Millisecond))
		user3Mod := userData3
		user3Mod.Score = 150
		mockCall.Unset()
		mockDbProvider.On("Top", gameId1, uint32(10)).Return(dbprovider.TopData{userData1, userData2, user3Mod}, nil)
		top, err = cacheProvider.Top(context.Background(), gameId1, 10)
		require.NoError(t, err)
		require.Equal(t, dbprovider.TopData{userData1, userData2, userData3}, top)
		mockDbProvider.AssertNumberOfCalls(t, "Top", 1)

		mockClock.SetTime(now.Add(time.Duration(CACHE_TTL) * time.Millisecond))
		top, err = cacheProvider.Top(context.Background(), gameId1, 10)
		require.NoError(t, err)
		require.Equal(t, dbprovider.TopData{userData1, userData2, user3Mod}, top)
		mockDbProvider.AssertNumberOfCalls(t, "Top", 2)
	})

}
