package postgre_provider

import (
	"context"
	"fmt"
	dbprovider "go-leaderboard-server/internal/db"
	"go-leaderboard-server/internal/utils"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const POSTGRES_USER string = "admin"
const POSTGRES_PASSWORD string = "admpass"
const POSTGRES_DB string = "GoLeaderboard"

var dbEndpoint string

func prepareTest(t *testing.T) {
	t.Log("prepare test env")

	ctx := context.Background()
	dbContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "postgres:16.1",
			ExposedPorts: []string{"5432"},
			Files: []testcontainers.ContainerFile{
				{
					HostFilePath:      filepath.Join(utils.GetTestFilePath(), "postgresql_setup.sql"),
					ContainerFilePath: "/docker-entrypoint-initdb.d/init.sql",
					FileMode:          0o004,
				},
			},
			Env: map[string]string{
				"POSTGRES_USER":     POSTGRES_USER,
				"POSTGRES_PASSWORD": POSTGRES_PASSWORD,
				"POSTGRES_DB":       POSTGRES_DB,
			},
			WaitingFor: wait.ForListeningPort("5432/tcp"),
		},
		Started: true,
	})
	require.NoError(t, err, "container should start successfully")

	t.Cleanup(func() {
		t.Log("terminate test env")

		err := dbContainer.Terminate(ctx)
		require.NoError(t, err, "container should be terminated successfully")
	})

	ep, err := dbContainer.Endpoint(ctx, "")
	require.NoError(t, err, "container endpoint should be obtained successfully")

	dbEndpoint = ep
}

func setupTest() (func() error, *PostgreProvider, error) {
	dbProvider := NewPostgreProvider()
	err := dbProvider.Initialize(context.Background(), &PostgreProviderConfig{
		DBProviderBaseConfig: dbprovider.DBProviderBaseConfig{
			IsDebug: true,
		},
		ConnStr: fmt.Sprintf("postgres://%s:%s@%s/%s", POSTGRES_USER, POSTGRES_PASSWORD, dbEndpoint, POSTGRES_DB),
	})

	return func() error {
		return dbProvider.Shutdown(context.Background())
	}, dbProvider, err
}

func runTest(t *testing.T, name string, testFunc utils.TestFcn[*PostgreProvider]) {
	utils.RunTest(t, name, setupTest, testFunc)
}

func TestPostgreProvider(t *testing.T) {
	prepareTest(t)

	gameId1 := "game1"
	gameId2 := "game2"
	gameId3 := "game3"
	userId1 := "user1"
	userId2 := "user2"

	userProp1 := dbprovider.UserProperties{
		Name:   "Ted",
		Score:  11,
		Params: "some_payload_1",
	}

	userProp2 := dbprovider.UserProperties{
		Name:  "Jane",
		Score: 56,
	}

	topData := dbprovider.TopData{
		{UserId: userId2, UserProperties: userProp2},
		{UserId: userId1, UserProperties: userProp1},
	}

	runTest(t, "get empty data", func(t *testing.T, dbProvider *PostgreProvider) {
		data, err := dbProvider.Get(context.Background(), gameId1, userId1)
		require.NoError(t, err)
		require.Nil(t, data)
	})

	runTest(t, "set, get and delete data", func(t *testing.T, dbProvider *PostgreProvider) {
		var (
			data *dbprovider.UserProperties
			err  error
		)

		err = dbProvider.Put(context.Background(), gameId2, userId1, userProp1)
		require.NoError(t, err)
		data, err = dbProvider.Get(context.Background(), gameId2, userId1)
		require.NoError(t, err)
		require.Equal(t, userProp1, *data)

		var userProp1Mod = userProp1
		userProp1Mod.Score = 33
		err = dbProvider.Put(context.Background(), gameId2, userId1, userProp1Mod)
		require.NoError(t, err)
		data, err = dbProvider.Get(context.Background(), gameId2, userId1)
		require.NoError(t, err)
		require.Equal(t, userProp1Mod, *data)

		err = dbProvider.Delete(context.Background(), gameId2, userId1)
		require.NoError(t, err)
		data, err = dbProvider.Get(context.Background(), gameId2, userId1)
		require.NoError(t, err)
		require.Nil(t, data)
	})

	runTest(t, "get top data", func(t *testing.T, dbProvider *PostgreProvider) {
		var (
			top dbprovider.TopData
			err error
		)

		top, err = dbProvider.Top(context.Background(), gameId3, 10)
		require.NoError(t, err)
		require.Equal(t, dbprovider.TopData{}, top)

		err = dbProvider.Put(context.Background(), gameId3, userId1, userProp1)
		require.NoError(t, err)
		err = dbProvider.Put(context.Background(), gameId3, userId2, userProp2)
		require.NoError(t, err)

		top, err = dbProvider.Top(context.Background(), gameId3, 10)
		require.NoError(t, err)
		require.Equal(t, topData, top)

		top, err = dbProvider.Top(context.Background(), gameId3, 1)
		require.NoError(t, err)
		require.Equal(t, topData[:1], top)
	})

}
