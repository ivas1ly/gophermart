package controller

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/jwtauth/v5"
	"github.com/go-chi/render"

	"github.com/ivas1ly/gophermart/internal/api/controller"
	"github.com/ivas1ly/gophermart/internal/entity"
	"github.com/ivas1ly/gophermart/pkg/lunh"
)

func (oh *OrderHandler) order(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	token, _, _ := jwtauth.FromContext(r.Context())

	userID := token.Subject()

	buf, err := io.ReadAll(r.Body)
	if len(buf) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, render.M{"message": controller.MsgEmptyBody})
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, render.M{"message": "can't read request body"})
		return
	}

	orderNumber := strings.TrimSpace(string(buf))
	ok := lunh.CheckNumber(orderNumber)
	if !ok {
		w.WriteHeader(http.StatusUnprocessableEntity)
		render.JSON(w, r, render.M{"message": controller.MsgIncorrectOrderNumber})
		return
	}

	orderInfo := &entity.OrderInfo{
		UserID: userID,
		Number: orderNumber,
	}

	order, err := oh.orderService.AddOrder(r.Context(), orderInfo)
	if errors.Is(err, entity.ErrUploadedByThisUser) {
		w.WriteHeader(http.StatusOK)
		render.JSON(w, r, render.M{"message": entity.ErrUploadedByThisUser.Error()})
		return
	}
	if errors.Is(err, entity.ErrUploadedByAnotherUser) {
		w.WriteHeader(http.StatusConflict)
		render.JSON(w, r, render.M{"message": entity.ErrUploadedByAnotherUser.Error()})
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, render.M{"message": controller.MsgInternalServerError})
		return
	}

	response := ToOrderResponse(order)

	w.WriteHeader(http.StatusAccepted)
	render.JSON(w, r, response)
}
