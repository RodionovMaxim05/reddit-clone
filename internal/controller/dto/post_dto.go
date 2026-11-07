package dto

import (
	"time"

	"redditclone/internal/domain"
)

type CreatePostRequest struct {
	Title    string `json:"title"`
	Category string `json:"category"`
	Type     string `json:"type"`
	URL      string `json:"url,omitempty"`
	Text     string `json:"text,omitempty"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

type UserResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Admin    bool   `json:"admin,omitempty"`
}

type VoteResponse struct {
	User string `json:"user"`
	Vote int    `json:"vote"`
}

type CommentResponse struct {
	ID      string       `json:"id"`
	Author  UserResponse `json:"author"`
	Body    string       `json:"body"`
	Created time.Time    `json:"created"`
}

type PostResponse struct {
	ID               string            `json:"id"`
	Title            string            `json:"title"`
	Category         string            `json:"category"`
	Type             string            `json:"type"`
	URL              string            `json:"url,omitempty"`
	Text             string            `json:"text,omitempty"`
	Author           UserResponse      `json:"author"`
	Score            int               `json:"score"`
	Votes            []VoteResponse    `json:"votes"`
	Comments         []CommentResponse `json:"comments"`
	Created          time.Time         `json:"created"`
	Views            int               `json:"views"`
	UpvotePercentage int               `json:"upvotePercentage"`
}

func NewPostResponse(p *domain.Post) PostResponse {
	votes := make([]VoteResponse, 0, len(p.Votes))
	for _, v := range p.Votes {
		votes = append(votes, VoteResponse{
			User: v.UserID,
			Vote: v.Value,
		})
	}

	comments := make([]CommentResponse, 0, len(p.Comments))
	for _, c := range p.Comments {
		comments = append(comments, CommentResponse{
			ID: c.ID,
			Author: UserResponse{
				ID:       c.Author.ID,
				Username: c.Author.Username,
				Admin:    c.Author.Admin,
			},
			Body:    c.Body,
			Created: c.CreatedAt,
		})
	}

	return PostResponse{
		ID:       p.ID,
		Title:    p.Title,
		Category: p.Category,
		Type:     p.Type,
		URL:      p.URL,
		Text:     p.Text,
		Author: UserResponse{
			ID:       p.Author.ID,
			Username: p.Author.Username,
			Admin:    p.Author.Admin,
		},
		Score:            p.Score,
		Votes:            votes,
		Comments:         comments,
		Created:          p.CreatedAt,
		Views:            p.Views,
		UpvotePercentage: p.UpvotePercentage,
	}
}

func NewPostListResponse(posts []*domain.Post) []PostResponse {
	res := make([]PostResponse, 0, len(posts))
	for _, p := range posts {
		res = append(res, NewPostResponse(p))
	}
	return res
}
