package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go-leaderboard-server/internal/config"
	"go-leaderboard-server/internal/controllers"
	dbprovider "go-leaderboard-server/internal/db"
	log "go-leaderboard-server/internal/logger"
	"go-leaderboard-server/internal/utils"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestServer(t *testing.T) {
	logger := log.GetLogger()
	config := config.GetAppConfig()

	logger.Initialize(config.IsDebug)

	now := time.Now()

	setupTest := func() (func() error, *AppServer, error) {
		server := NewAppServer(nil)
		err := server.Initialize(config)
		return func() error {
			return server.Shutdown()
		}, server, err
	}

	runTest := func(name string, testFunc utils.TestFcn[*AppServer]) {
		utils.RunTest(t, name, setupTest, testFunc)
	}

	apiCall := func(server *AppServer, method string, path string, body string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		var bbuf io.Reader
		if body != "" {
			bbuf = bytes.NewBuffer([]byte(body))
		}
		req, _ := http.NewRequest(method, path, bbuf)
		req.Header.Set("Content-Type", "application/json")
		server.router.ServeHTTP(w, req)
		return w
	}

	gameId := "game1"

	user1 := dbprovider.UserData{
		UserId: "user1",
		UserProperties: dbprovider.UserProperties{
			Name:  "user1_name",
			Score: 24,
		},
	}

	user2 := dbprovider.UserData{
		UserId: "user2",
		UserProperties: dbprovider.UserProperties{
			Name:   "user2_name",
			Score:  83,
			Params: "some_payload_2",
		},
	}

	user3 := dbprovider.UserData{
		UserId: "user3",
		UserProperties: dbprovider.UserProperties{
			Score:  44,
			Params: "some_payload_3",
		},
	}

	runTest("correct realtime", func(t *testing.T, server *AppServer) {
		time1 := server.clock.Now().UnixMilli()
		time.Sleep(50 * time.Millisecond)
		time2 := server.clock.Now().UnixMilli()
		require.Greater(t, time2, time1)
	})

	runTest("status => success", func(t *testing.T, server *AppServer) {
		w := apiCall(server, "GET", "/Status", "")
		require.Equal(t, http.StatusOK, w.Code)
		require.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))
		require.JSONEq(t, `{"result": "success"}`, w.Body.String())
	})

	runTest("get top => empty data", func(t *testing.T, server *AppServer) {
		w := apiCall(server, "POST", "/leaderboard/GetTop",
			fmt.Sprintf(`{ "gameId": "%s", "nTop": 10 }`, gameId),
		)
		require.Equal(t, http.StatusOK, w.Code)
		require.JSONEq(t, `{"result": []}`, w.Body.String())
	})

	runTest("send score => success", func(t *testing.T, server *AppServer) {
		w := apiCall(server, "POST", "/leaderboard/SendScore",
			fmt.Sprintf(`{ "gameId": "%s", "userId": "%s", "name": "%s", "score": %f, "params": "%s" }`,
				gameId, user1.UserId, user1.Name, user1.Score, user1.Params),
		)
		require.Equal(t, http.StatusOK, w.Code)
		require.JSONEq(t, `{"result": "success"}`, w.Body.String())
	})

	runTest("get score => success", func(t *testing.T, server *AppServer) {
		w := apiCall(server, "POST", "/leaderboard/SendScore",
			fmt.Sprintf(`{ "gameId": "%s", "userId": "%s", "name": "%s", "score": %f, "params": "%s" }`,
				gameId, user2.UserId, user2.Name, user2.Score, user2.Params),
		)
		require.Equal(t, http.StatusOK, w.Code)
		require.JSONEq(t, `{"result": "success"}`, w.Body.String())

		w = apiCall(server, "POST", "/leaderboard/GetScore",
			fmt.Sprintf(`{ "gameId": "%s", "userId": "%s" }`, gameId, user2.UserId),
		)
		require.Equal(t, http.StatusOK, w.Code)
		require.JSONEq(t, fmt.Sprintf(`{"result": { "name": "%s", "score": %f, "params": "%s" } }`,
			user2.Name, user2.Score, user2.Params), w.Body.String())
	})

	runTest("delete score => success", func(t *testing.T, server *AppServer) {
		var w *httptest.ResponseRecorder

		w = apiCall(server, "POST", "/leaderboard/SendScore",
			fmt.Sprintf(`{ "gameId": "%s", "userId": "%s", "name": "%s", "score": %f, "params": "%s" }`,
				gameId, user3.UserId, user3.Name, user3.Score, user3.Params),
		)
		require.Equal(t, http.StatusOK, w.Code)
		require.JSONEq(t, `{"result": "success"}`, w.Body.String())

		w = apiCall(server, "POST", "/leaderboard/DeleteScore",
			fmt.Sprintf(`{ "gameId": "%s", "userId": "%s" }`, gameId, user3.UserId),
		)
		require.Equal(t, http.StatusOK, w.Code)
		require.JSONEq(t, `{"result": "success"}`, w.Body.String())

		w = apiCall(server, "POST", "/leaderboard/GetScore",
			fmt.Sprintf(`{ "gameId": "%s", "userId": "%s" }`, gameId, user3.UserId),
		)
		require.Equal(t, http.StatusOK, w.Code)
		require.JSONEq(t, `{"result": {}}`, w.Body.String())
	})

	runTest("get top => success", func(t *testing.T, server *AppServer) {
		var w *httptest.ResponseRecorder

		server.clock = &utils.MockClock{}

		w = apiCall(server, "POST", "/leaderboard/SendScore",
			fmt.Sprintf(`{ "gameId": "%s", "userId": "%s", "name": "%s", "score": %f, "params": "%s" }`,
				gameId, user1.UserId, user1.Name, user1.Score, user1.Params),
		)
		require.Equal(t, http.StatusOK, w.Code)
		require.JSONEq(t, `{"result": "success"}`, w.Body.String())

		w = apiCall(server, "POST", "/leaderboard/SendScore",
			fmt.Sprintf(`{ "gameId": "%s", "userId": "%s", "name": "%s", "score": %f, "params": "%s" }`,
				gameId, user2.UserId, user2.Name, user2.Score, user2.Params),
		)
		require.Equal(t, http.StatusOK, w.Code)
		require.JSONEq(t, `{"result": "success"}`, w.Body.String())

		w = apiCall(server, "POST", "/leaderboard/SendScore",
			fmt.Sprintf(`{ "gameId": "%s", "userId": "%s", "name": "%s", "score": %f, "params": "%s" }`,
				gameId, user3.UserId, user3.Name, user3.Score, user3.Params),
		)
		require.Equal(t, http.StatusOK, w.Code)
		require.JSONEq(t, `{"result": "success"}`, w.Body.String())

		server.clock.(*utils.MockClock).SetTime(now.Add(time.Duration(config.Cache.Config.GetBaseConfig().Ttl) * time.Millisecond))

		jtop, _ := json.Marshal(dbprovider.TopData{user2, user3, user1})
		for i := 0; i < 2; i++ {
			w := apiCall(server, "POST", "/leaderboard/GetTop",
				fmt.Sprintf(`{ "gameId": "%s", "nTop": 10 }`, gameId),
			)
			require.Equal(t, http.StatusOK, w.Code)
			require.JSONEq(t, fmt.Sprintf(`{"result": %s}`, string(jtop)), w.Body.String())
		}
	})

	runTest("not found => error 404", func(t *testing.T, server *AppServer) {
		w := apiCall(server, "POST", "/fake_path", "")
		require.Equal(t, http.StatusNotFound, w.Code)
		require.JSONEq(t, `{"error": "Not found"}`, w.Body.String())
	})

	runTest("wrong request params => error 400", func(t *testing.T, server *AppServer) {
		var res controllers.ResultError
		w := apiCall(server, "POST", "/leaderboard/SendScore", fmt.Sprintf(`{ gameId: "%s", fake_param: "abc" }`, gameId))
		require.Equal(t, http.StatusBadRequest, w.Code)
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &res))
	})

	runTest("unexpected exception => error 500", func(t *testing.T, server *AppServer) {
		server.clock = &utils.BrokenClock{}
		w := apiCall(server, "POST", "/leaderboard/GetTop",
			fmt.Sprintf(`{ "gameId": "%s", "nTop": 10 }`, gameId),
		)
		require.Equal(t, http.StatusInternalServerError, w.Code)
		require.JSONEq(t, `{"error": "Internal server error"}`, w.Body.String())
	})
}
