package controller

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"redditclone/internal/controller/dto"
	"redditclone/internal/domain"
	"redditclone/internal/repository"
)

type PostHandler struct {
	repo repository.PostRepository
}

func NewPostHandler(repo repository.PostRepository) *PostHandler {
	return &PostHandler{repo: repo}
}

func (h *PostHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	user, ok := UserFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, dto.ErrorResponse{Message: "unauthorized"})
		return
	}

	var req dto.CreatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{Message: "bad request"})
		return
	}

	// Field validation
	var fieldErrors []dto.FieldError

	if len(req.Title) < 1 || len(req.Title) > 100 {
		fieldErrors = append(fieldErrors, dto.FieldError{
			Location: "body",
			Param:    "title",
			Msg:      "is required and must be at most 100 characters",
		})
	}
	if req.Category == "" {
		fieldErrors = append(fieldErrors, dto.FieldError{
			Location: "body",
			Param:    "category",
			Msg:      "is required",
		})
	}
	if req.Type != "link" && req.Type != "text" {
		fieldErrors = append(fieldErrors, dto.FieldError{
			Location: "body",
			Param:    "type",
			Msg:      "must be either link or text",
		})
	} else {
		if req.Type == "link" && req.URL == "" {
			fieldErrors = append(fieldErrors, dto.FieldError{
				Location: "body",
				Param:    "url",
				Msg:      "is required for link post",
			})
		}
		if req.Type == "text" && len(req.Text) < 4 {
			fieldErrors = append(fieldErrors, dto.FieldError{
				Location: "body",
				Param:    "text",
				Msg:      "must be at least 4 characters long",
			})
		}
	}

	if len(fieldErrors) > 0 {
		writeJSON(w, http.StatusUnprocessableEntity, dto.ValidationErrorResponse{Errors: fieldErrors})
		return
	}

	now := time.Now().UTC()
	post := &domain.Post{
		Title:    req.Title,
		Category: req.Category,
		Type:     req.Type,
		URL:      req.URL,
		Text:     req.Text,
		Author:   *user,
		Score:    1,
		Votes: []domain.Vote{
			{UserID: user.ID, Value: 1},
		},
		Comments:         make([]domain.Comment, 0),
		CreatedAt:        now,
		Views:            0,
		UpvotePercentage: 100,
	}

	if err := h.repo.Create(post); err != nil {
		writeJSON(w, http.StatusInternalServerError, dto.ErrorResponse{Message: "internal server error"})
		return
	}

	response := dto.NewPostResponse(post)
	writeJSON(w, http.StatusCreated, response)
}

func (h *PostHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	user, ok := UserFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, dto.ErrorResponse{Message: "unauthorized"})
		return
	}

	postID := r.PathValue("postId")

	err := h.repo.Delete(postID, user)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrPostNotFound):
			writeJSON(w, http.StatusNotFound, dto.ErrorResponse{Message: "post not found"})
		case errors.Is(err, domain.ErrUnauthorized):
			writeJSON(w, http.StatusUnauthorized, dto.ErrorResponse{Message: "unauthorized"})
		default:
			writeJSON(w, http.StatusInternalServerError, dto.ErrorResponse{Message: "internal server error"})
		}
		return
	}

	writeJSON(w, http.StatusOK, dto.SuccessResponse{Message: "success"})
}

func (h *PostHandler) ListPosts(w http.ResponseWriter, r *http.Request) {
	posts, err := h.repo.GetAll()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "internal server error"})
		return
	}

	response := dto.NewPostListResponse(posts)
	writeJSON(w, http.StatusOK, response)
}

func (h *PostHandler) GetPost(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("postId")

	p, err := h.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, domain.ErrPostNotFound) {
			writeJSON(w, http.StatusNotFound, dto.ErrorResponse{Message: "post not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, dto.ErrorResponse{Message: "internal server error"})
		return
	}

	response := dto.NewPostResponse(p)
	writeJSON(w, http.StatusOK, response)
}

func (h *PostHandler) ListPostsByCategory(w http.ResponseWriter, r *http.Request) {
	category := r.PathValue("category")

	posts, err := h.repo.GetByCategory(category)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, dto.ErrorResponse{Message: "internal server error"})
		return
	}

	response := dto.NewPostListResponse(posts)
	writeJSON(w, http.StatusOK, response)
}

func (h *PostHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
	user, ok := UserFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, dto.ErrorResponse{Message: "unauthorized"})
		return
	}

	postID := r.PathValue("postId")

	var req dto.CreateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{Message: "bad request"})
		return
	}

	if len(req.Comment) < 1 || len(req.Comment) > 2000 {
		writeJSON(w, http.StatusUnprocessableEntity, dto.ValidationErrorResponse{
			Errors: []dto.FieldError{
				{
					Location: "body",
					Param:    "comment",
					Msg:      "is required and must be at most 2000 characters",
				},
			},
		})
		return
	}

	comment := &domain.Comment{
		Author: *user,
		Body:   req.Comment,
	}

	updatedPost, err := h.repo.AddComment(postID, comment)
	if err != nil {
		if errors.Is(err, domain.ErrPostNotFound) {
			writeJSON(w, http.StatusNotFound, dto.ErrorResponse{Message: "post not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, dto.ErrorResponse{Message: "internal server error"})
		return
	}

	response := dto.NewPostResponse(updatedPost)
	writeJSON(w, http.StatusCreated, response)
}

func (h *PostHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	user, ok := UserFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, dto.ErrorResponse{Message: "unauthorized"})
		return
	}

	postID := r.PathValue("postId")
	commentID := r.PathValue("commentId")

	updatedPost, err := h.repo.DeleteComment(postID, commentID, user)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrPostNotFound):
			writeJSON(w, http.StatusNotFound, dto.ErrorResponse{Message: "post not found"})
		case errors.Is(err, domain.ErrCommentNotFound):
			writeJSON(w, http.StatusNotFound, dto.ErrorResponse{Message: "comment not found"})
		case errors.Is(err, domain.ErrUnauthorized):
			writeJSON(w, http.StatusUnauthorized, dto.ErrorResponse{Message: "unauthorized"})
		default:
			writeJSON(w, http.StatusInternalServerError, dto.ErrorResponse{Message: "internal server error"})
		}
		return
	}

	response := dto.NewPostResponse(updatedPost)
	writeJSON(w, http.StatusOK, response)
}

func (h *PostHandler) UpvotePost(w http.ResponseWriter, r *http.Request) {
	user, ok := UserFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, dto.ErrorResponse{Message: "unauthorized"})
		return
	}

	postID := r.PathValue("postId")

	updatedPost, err := h.repo.Vote(postID, user.ID, 1)
	if err != nil {
		if errors.Is(err, domain.ErrPostNotFound) {
			writeJSON(w, http.StatusNotFound, dto.ErrorResponse{Message: "post not found"})

			return
		}
		writeJSON(w, http.StatusInternalServerError, dto.ErrorResponse{Message: "internal server error"})
		return
	}

	response := dto.NewPostResponse(updatedPost)
	writeJSON(w, http.StatusOK, response)
}

func (h *PostHandler) DownvotePost(w http.ResponseWriter, r *http.Request) {
	user, ok := UserFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, dto.ErrorResponse{Message: "unauthorized"})
		return
	}

	postID := r.PathValue("postId")

	updatedPost, err := h.repo.Vote(postID, user.ID, -1)
	if err != nil {
		if errors.Is(err, domain.ErrPostNotFound) {
			writeJSON(w, http.StatusNotFound, dto.ErrorResponse{Message: "post not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, dto.ErrorResponse{Message: "internal server error"})
		return
	}

	response := dto.NewPostResponse(updatedPost)
	writeJSON(w, http.StatusOK, response)
}

func (h *PostHandler) UnvotePost(w http.ResponseWriter, r *http.Request) {
	user, ok := UserFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, dto.ErrorResponse{Message: "unauthorized"})
		return
	}

	postID := r.PathValue("postId")

	updatedPost, err := h.repo.Unvote(postID, user.ID)
	if err != nil {
		if errors.Is(err, domain.ErrPostNotFound) {
			writeJSON(w, http.StatusNotFound, dto.ErrorResponse{Message: "post not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, dto.ErrorResponse{Message: "internal server error"})
		return
	}

	response := dto.NewPostResponse(updatedPost)
	writeJSON(w, http.StatusOK, response)
}
