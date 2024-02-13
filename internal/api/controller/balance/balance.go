package controller

import (
	"net/http"

	"github.com/go-chi/jwtauth/v5"
	"github.com/go-chi/render"

	"github.com/ivas1ly/gophermart/internal/api/controller"
)

func (bh *BalanceHandler) Balance(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	token, _, _ := jwtauth.FromContext(r.Context())

	userID := token.Subject()

	currentBalance, err := bh.balanceService.GetCurrentBalance(r.Context(), userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, render.M{"message": controller.MsgInternalServerError})
		return
	}

	response := ToUserBalanceResponse(currentBalance)

	w.WriteHeader(http.StatusOK)
	render.JSON(w, r, response)
}
