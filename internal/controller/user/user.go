package user

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"

	"github.com/ivas1ly/gophermart/internal/controller/dto"
	"github.com/ivas1ly/gophermart/internal/entity"
	"github.com/ivas1ly/gophermart/internal/service"
	"github.com/ivas1ly/gophermart/internal/utils/jwt"
)

const (
	MsgEmptyBody           = "empty request body"
	MsgCantParseBody       = "can't parse request body"
	MsgInvalidRequest      = "invalid request format"
	MsgInternalServerError = "internal server error"
)

type Handler struct {
	userService service.UserService
	log         *zap.Logger
	validate    *validator.Validate
	tokenAuth   *jwtauth.JWTAuth
}

func NewHandler(userService service.UserService, log *zap.Logger, validate *validator.Validate) *Handler {
	tokenAuth := jwtauth.New("HS256", jwt.SigningKey, nil)

	return &Handler{
		userService: userService,
		log:         log,
		validate:    validate,
		tokenAuth:   tokenAuth,
	}
}

func (h *Handler) Register(router *chi.Mux) {
	router.Route("/api/user", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Post("/login", h.login)
			r.Post("/register", h.register)
		})
		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(jwtauth.Verifier(h.tokenAuth), jwtauth.Authenticator(h.tokenAuth))
			r.Post("/balance", h.balance)
		})
	})
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var ur dto.UserRequest
	defer r.Body.Close()

	err := json.NewDecoder(r.Body).Decode(&ur)
	if errors.Is(err, io.EOF) {
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, render.M{"message": MsgEmptyBody})
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, render.M{"message": MsgCantParseBody})
		return
	}

	err = h.validate.Struct(ur)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, render.M{"message": MsgInvalidRequest})
		return
	}

	user, err := h.userService.Register(r.Context(), ur.Username, ur.Password)
	if errors.Is(err, entity.ErrUsernameUniqueViolation) {
		w.WriteHeader(http.StatusConflict)
		render.JSON(w, r, render.M{"message": fmt.Sprintf("username %q already exists", ur.Username)})
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, render.M{"message": MsgInternalServerError})
		return
	}

	response := dto.ToUserFromService(user)

	authToken, err := jwt.NewToken(jwt.SigningKey, user.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, render.M{"message": MsgInternalServerError})
		return
	}
	response.AuthToken = authToken

	w.WriteHeader(http.StatusOK)
	render.JSON(w, r, response)
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var ur dto.UserRequest
	defer r.Body.Close()

	err := json.NewDecoder(r.Body).Decode(&ur)
	if errors.Is(err, io.EOF) {
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, render.M{"message": MsgEmptyBody})
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, render.M{"message": MsgCantParseBody})
		return
	}

	err = h.validate.Struct(ur)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, render.M{"message": MsgInvalidRequest})
		return
	}

	user, err := h.userService.Login(r.Context(), ur.Username, ur.Password)
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
		render.JSON(w, r, render.M{"message": MsgInternalServerError})
		return
	}

	response := dto.ToUserFromService(user)

	authToken, err := jwt.NewToken(jwt.SigningKey, user.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, render.M{"message": MsgInternalServerError})
		return
	}
	response.AuthToken = authToken

	w.WriteHeader(http.StatusOK)
	render.JSON(w, r, response)
}

func (h *Handler) balance(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("protected route: %v", claims)))
}
