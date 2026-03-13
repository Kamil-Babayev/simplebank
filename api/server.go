package api

import (
	"fmt"
	db "simplebank/db/sqlc"
	"simplebank/token"
	"simplebank/util"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

type Server struct {
	store      db.Store
	router     *gin.Engine
	tokenMaker token.Maker
	config     util.Config
}

func NewServer(config util.Config, store db.Store) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.SymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token tokenMaker: %w", err)
	}

	server := &Server{
		store:      store,
		tokenMaker: tokenMaker,
		config:     config,
	}

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		err := v.RegisterValidation("currency", validCurrency)
		if err != nil {
			return nil, fmt.Errorf("cannot create currency validator: %w", err)
		}
	}

	server.setupRouter()

	return server, nil
}

func (s *Server) setupRouter() {
	router := gin.Default()

	// Endpoints for users
	router.POST("/users", s.createUser)
	// Endpoints for authentification
	router.POST("/login", s.loginUser)

	// Secure endpoints
	authRoutes := router.Group("/").Use(authMiddleware(s.tokenMaker))

	// Endpoints for Account
	authRoutes.POST("/accounts", s.createAccount)
	authRoutes.GET("/accounts", s.listAccounts)
	authRoutes.GET("/accounts/:id", s.getAccount)
	authRoutes.PATCH("/accounts/:id", s.updateAccount)
	authRoutes.DELETE("/accounts/:id", s.deleteAccount)

	// Endpoints for transfers
	authRoutes.POST("/transfers", s.createTransfer)

	// Endpoints for users
	authRoutes.GET("/users/:username", s.getUser)

	s.router = router
}

func (s *Server) Start(address string) error {
	return s.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
