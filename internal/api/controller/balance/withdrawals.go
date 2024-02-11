package controller

import (
	"errors"
	"net/http"

	"github.com/go-chi/jwtauth/v5"
	"github.com/go-chi/render"

	"github.com/ivas1ly/gophermart/internal/api/controller"
	"github.com/ivas1ly/gophermart/internal/api/controller/dto"
	"github.com/ivas1ly/gophermart/internal/entity"
)

func (bh *BalanceHandler) withdrawals(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	token, _, _ := jwtauth.FromContext(r.Context())

	userID := token.Subject()

	withdrawals, err := bh.balanceService.GetWithdrawals(r.Context(), userID)
	if errors.Is(err, entity.ErrNoWithdrawalsFound) {
		w.WriteHeader(http.StatusNoContent)
		render.JSON(w, r, render.M{"message": controller.MsgInternalServerError})
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, render.M{"message": controller.MsgInternalServerError})
		return
	}

	response := dto.ToWithdrawalsResponse(withdrawals)

	w.WriteHeader(http.StatusOK)
	render.JSON(w, r, response)
}
