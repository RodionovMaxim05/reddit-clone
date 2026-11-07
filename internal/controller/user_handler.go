package controller

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"

	"redditclone/internal/controller/dto"
	"redditclone/internal/domain"
	"redditclone/internal/pkg/auth"
	"redditclone/internal/repository"
)

type UserHandler struct {
	repo         repository.UserRepository
	tokenManager *auth.TokenManager
}

func NewUserHandler(repo repository.UserRepository, tokenManager *auth.TokenManager) *UserHandler {
	return &UserHandler{
		repo:         repo,
		tokenManager: tokenManager,
	}
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.CredentialsRequest

	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))

	if err := decoder.Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{Message: "bad request"})
		return
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{Message: "bad request"})
		return
	}

	// Field validation
	validUsername, _ := regexp.MatchString(`^[a-zA-Z0-9_-]{1,32}$`, req.Username)
	if !validUsername {
		writeJSON(w, http.StatusUnprocessableEntity, dto.NewValidationError("username", "must be 1-32 letters, digits, underscores or hyphens"))
		return
	}
	if utf8.RuneCountInString(req.Password) < 8 || utf8.RuneCountInString(req.Password) > 72 {
		writeJSON(w, http.StatusUnprocessableEntity, dto.NewValidationError("password", "must be between 8 and 72 characters"))
		return
	}

	// Password hashing
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, dto.ErrorResponse{Message: "internal server error"})
		return
	}

	user := &domain.User{
		Username: req.Username,
		Password: string(hashedPassword),
	}

	if err := h.repo.Create(user); err != nil {
		if errors.Is(err, domain.ErrUserAlreadyExists) {
			writeJSON(w, http.StatusUnprocessableEntity, dto.NewValidationError("username", "already exists"))
			return
		}
		writeJSON(w, http.StatusInternalServerError, dto.ErrorResponse{Message: "internal server error"})
		return
	}

	token, err := h.tokenManager.GenerateToken(user)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, dto.ErrorResponse{Message: "internal server error"})
		return
	}

	writeJSON(w, http.StatusCreated, dto.AuthResponse{Token: token})
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.CredentialsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{Message: "bad request"})
		return
	}

	user, err := h.repo.GetByUsername(req.Username)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			writeJSON(w, http.StatusUnauthorized, dto.ErrorResponse{Message: "user not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, dto.ErrorResponse{Message: "internal server error"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		writeJSON(w, http.StatusUnauthorized, dto.ErrorResponse{Message: "invalid password"})
		return
	}

	token, err := h.tokenManager.GenerateToken(user)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, dto.ErrorResponse{Message: "internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, dto.AuthResponse{Token: token})
}
