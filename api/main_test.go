package api

import (
	"os"
	db "simplebank/db/sqlc"
	"simplebank/util"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func newTestServer(t *testing.T, store db.Store) *Server {
	t.Helper()
	config := util.Config{
		SymmetricKey:        util.RandomString(32),
		AccessTokenDuration: time.Minute,
	}
	server, err := NewServer(config, store)
	require.NoError(t, err)
	require.NotEmpty(t, server)

	return server
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}
