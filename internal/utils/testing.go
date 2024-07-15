package utils

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

type SetupTestFcn[I any] func() (func() error, I, error)
type TestFcn[I any] func(*testing.T, I)

func RunTest[I any](t *testing.T, name string, setupFunc SetupTestFcn[I], testFunc TestFcn[I]) {
	t.Run(name, func(t *testing.T) {
		t.Log("setup test")
		shutdownTest, param, err := setupFunc()
		defer func() {
			t.Log("shutdown test")
			err := shutdownTest()
			require.NoError(t, err, "test shutdown should not fail")
		}()
		require.NoError(t, err, "test setup should not fail")

		t.Log("run test")
		testFunc(t, param)
	})
}

func GetTestFilePath() string {
	_, fname, _, _ := runtime.Caller(1)
	return filepath.Dir(fname)
}
