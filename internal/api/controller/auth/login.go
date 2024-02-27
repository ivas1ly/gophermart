package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/render"

	"github.com/ivas1ly/gophermart/internal/api/controller"
	"github.com/ivas1ly/gophermart/internal/entity"
	"github.com/ivas1ly/gophermart/pkg/jwt"
)

func (ah *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var ur UserRequest
	defer r.Body.Close()

	err := json.NewDecoder(r.Body).Decode(&ur)
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

	err = ah.validate.Struct(ur)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, render.M{"message": controller.MsgInvalidRequest})
		return
	}

	user, err := ah.authService.Login(r.Context(), ur.Username, ur.Password)
	if errors.Is(err, entity.ErrUsernameNotFound) {
		w.WriteHeader(http.StatusNotFound)
		render.JSON(w, r, render.M{"message": entity.ErrUsernameNotFound.Error()})
		return
	}
	if errors.Is(err, entity.ErrIncorrectLoginOrPassword) {
		w.WriteHeader(http.StatusUnauthorized)
		render.JSON(w, r, render.M{"message": entity.ErrIncorrectLoginOrPassword.Error()})
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, render.M{"message": controller.MsgInternalServerError})
		return
	}

	response := ToUserResponse(user)

	authToken, err := jwt.NewToken(jwt.SigningKey, user.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, render.M{"message": controller.MsgInternalServerError})
		return
	}

	w.Header().Set(AuthorizationHeader, fmt.Sprintf("%s %s", AuthorizationSchema, authToken))
	w.WriteHeader(http.StatusOK)
	render.JSON(w, r, response)
}
