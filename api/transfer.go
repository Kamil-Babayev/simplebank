package api

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	db "simplebank/db/sqlc"
	"simplebank/token"

	"github.com/gin-gonic/gin"
)

type createTransferRequest struct {
	FromAccountID int64  `json:"from_account_id" binding:"required,min=1"`
	ToAccountID   int64  `json:"to_account_id" binding:"required,min=1"`
	Amount        int64  `json:"amount" binding:"required,gt=0"`
	Currency      string `json:"currency" binding:"required,currency"`
}

func (s *Server) createTransfer(ctx *gin.Context) {
	var req createTransferRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	fromAccount, valid := s.validAccount(ctx, req.FromAccountID, req.Currency)
	if !valid {
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if authPayload.Username != fromAccount.Owner {
		err := errors.New("from account doesnt belong ro the authenticated user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}
	_, valid = s.validAccount(ctx, req.ToAccountID, req.Currency)
	if !valid {
		return
	}

	arg := db.TransferTxParam{
		FromAccountID: req.FromAccountID,
		ToAccountID:   req.ToAccountID,
		Amount:        req.Amount,
	}

	res, err := s.store.TransferTx(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
	}

	ctx.JSON(http.StatusOK, res)
}

func (s *Server) validAccount(ctx *gin.Context, ID int64, currency string) (db.Account, bool) {
	acc, err := s.store.GetAccount(ctx, ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return acc, false
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return acc, false
	}

	if acc.Currency != currency {
		ctx.JSON(http.StatusBadRequest, fmt.Errorf("account [%d] currency mismatch: %s vs %s", ID, acc.Currency, currency))
		return acc, false
	}

	return acc, true
}
