package api

import (
	db "simplebank/db/sqlc"

	"github.com/gin-gonic/gin"
)

type Server struct {
	store  db.Store
	router *gin.Engine
}

func NewServer(store db.Store) *Server {
	router := gin.Default()
	server := &Server{
		store: store,
	}

	// Endpoints for Account
	router.POST("/accounts", server.createAccount)
	router.GET("/accounts", server.listAccounts)
	router.GET("/accounts/:id", server.getAccount)
	router.PATCH("/accounts/:id", server.updateAccount)
	router.DELETE("/accounts/:id", server.deleteAccount)

	server.router = router
	return server
}

func (s *Server) Start(address string) error {
	return s.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
