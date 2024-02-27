package controller

import (
	"errors"
	"net/http"

	"github.com/go-chi/jwtauth/v5"
	"github.com/go-chi/render"

	"github.com/ivas1ly/gophermart/internal/api/controller"
	"github.com/ivas1ly/gophermart/internal/entity"
)

func (oh *OrderHandler) Orders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	token, _, _ := jwtauth.FromContext(r.Context())

	userID := token.Subject()

	orders, err := oh.orderService.GetOrders(r.Context(), userID)
	if errors.Is(err, entity.ErrNoOrdersFound) {
		w.WriteHeader(http.StatusNoContent)
		render.JSON(w, r, render.M{"message": controller.MsgInternalServerError})
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, render.M{"message": controller.MsgInternalServerError})
		return
	}

	response := ToOrdersResponse(orders)

	w.WriteHeader(http.StatusOK)
	render.JSON(w, r, response)
}
