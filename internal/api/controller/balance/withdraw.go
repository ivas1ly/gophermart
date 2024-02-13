package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/jwtauth/v5"
	"github.com/go-chi/render"
	"github.com/shopspring/decimal"

	"github.com/ivas1ly/gophermart/internal/api/controller"
	"github.com/ivas1ly/gophermart/internal/entity"
	"github.com/ivas1ly/gophermart/pkg/lunh"
)

func (bh *BalanceHandler) withdraw(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	token, _, _ := jwtauth.FromContext(r.Context())

	userID := token.Subject()

	var wr WithdrawRequest
	defer r.Body.Close()

	err := json.NewDecoder(r.Body).Decode(&wr)
	if errors.Is(err, io.EOF) {
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, render.M{"message": controller.MsgEmptyBody})
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, render.M{"message": controller.MsgCantParseBody})
		return
	}

	err = bh.validate.Struct(wr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, render.M{"message": controller.MsgInvalidRequest})
		return
	}

	okOrder := lunh.CheckNumber(strings.TrimSpace(wr.Order))
	if !okOrder {
		w.WriteHeader(http.StatusUnprocessableEntity)
		render.JSON(w, r, render.M{"message": controller.MsgIncorrectOrderNumber})
		return
	}

	okSum := wr.Sum.GreaterThanOrEqual(decimal.NewFromInt(1).Div(decimal.NewFromInt(entity.DecimalPartDiv)))
	if !okSum {
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, render.M{"message": "the amount to be withdrawn is less than the minimum amount"})
		return
	}

	intSum := wr.Sum.Mul(decimal.NewFromInt(entity.DecimalPartDiv)).IntPart()

	err = bh.balanceService.AddWithdrawal(r.Context(), userID, wr.Order, intSum)
	if errors.Is(err, entity.ErrNotEnoughPointsToWithdraw) {
		w.WriteHeader(http.StatusPaymentRequired)
		render.JSON(w, r, render.M{"message": entity.ErrNotEnoughPointsToWithdraw.Error()})
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, render.M{"message": controller.MsgInternalServerError})
		return
	}

	w.WriteHeader(http.StatusOK)
	render.JSON(w, r, render.M{
		"message": fmt.Sprintf("%s loyalty points are withdrawn for order %s", wr.Sum, wr.Order),
	})
}
