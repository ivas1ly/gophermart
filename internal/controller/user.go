package controller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"

	"github.com/ivas1ly/gophermart/internal/controller/dto"
	"github.com/ivas1ly/gophermart/internal/entity"
	"github.com/ivas1ly/gophermart/internal/utils/jwt"
	"github.com/ivas1ly/gophermart/internal/utils/lunh"
)

const (
	AuthorizationSchema = "Bearer"
	AuthorizationHeader = "Authorization"

	MsgEmptyBody           = "empty request body"
	MsgCantParseBody       = "can't parse request body"
	MsgInvalidRequest      = "invalid request format"
	MsgInternalServerError = "internal server error"
)

type UserService interface {
	Register(ctx context.Context, username, password string) (*entity.User, error)
	Login(ctx context.Context, username, password string) (*entity.User, error)
	NewOrder(ctx context.Context, userID, orderNumber string) (*entity.Order, error)
	GetOrders(ctx context.Context, userID string) ([]entity.Order, error)
}

type Handler struct {
	userService UserService
	log         *zap.Logger
	validate    *validator.Validate
	tokenAuth   *jwtauth.JWTAuth
}

func NewHandler(userService UserService, validate *validator.Validate, log *zap.Logger) *Handler {
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

			r.Post("/orders", h.order)
			r.Get("/orders", h.orders)
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

	response := dto.ToUserResponse(user)

	authToken, err := jwt.NewToken(jwt.SigningKey, user.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, render.M{"message": MsgInternalServerError})
		return
	}

	w.Header().Set(AuthorizationHeader, fmt.Sprintf("%s %s", AuthorizationSchema, authToken))
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

	response := dto.ToUserResponse(user)

	authToken, err := jwt.NewToken(jwt.SigningKey, user.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, render.M{"message": MsgInternalServerError})
		return
	}

	w.Header().Set(AuthorizationHeader, fmt.Sprintf("%s %s", AuthorizationSchema, authToken))
	w.WriteHeader(http.StatusOK)
	render.JSON(w, r, response)
}

func (h *Handler) order(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	token, _, _ := jwtauth.FromContext(r.Context())

	userID := token.Subject()

	buf, err := io.ReadAll(r.Body)
	if len(buf) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, render.M{"message": MsgEmptyBody})
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
		render.JSON(w, r, render.M{"message": "incorrect order number format"})
		return
	}

	order, err := h.userService.NewOrder(r.Context(), userID, orderNumber)
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
		render.JSON(w, r, render.M{"message": MsgInternalServerError})
		return
	}

	response := dto.ToOrderResponse(order)

	w.WriteHeader(http.StatusAccepted)
	render.JSON(w, r, response)
}

func (h *Handler) orders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	token, _, _ := jwtauth.FromContext(r.Context())

	userID := token.Subject()

	orders, err := h.userService.GetOrders(r.Context(), userID)
	if errors.Is(err, entity.ErrNoOrdersFound) {
		w.WriteHeader(http.StatusNoContent)
		render.JSON(w, r, render.M{"message": MsgInternalServerError})
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, render.M{"message": MsgInternalServerError})
		return
	}

	response := dto.ToOrdersResponse(orders)

	w.WriteHeader(http.StatusOK)
	render.JSON(w, r, response)
}

func (h *Handler) balance(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(fmt.Sprintf("protected route: %v", claims)))
}
