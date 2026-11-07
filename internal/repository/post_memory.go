package repository

import (
	"cmp"
	"fmt"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"redditclone/internal/domain"
)

type PostRepository interface {
	GetAll() ([]*domain.Post, error)
	GetByID(id string) (*domain.Post, error)
	GetByCategory(category string) ([]*domain.Post, error)
	Create(post *domain.Post) error
	AddComment(postID string, comment *domain.Comment) (*domain.Post, error)
	DeleteComment(postID, commentID string, user *domain.User) (*domain.Post, error)
}

type MemoryPostRepository struct {
	mu         sync.RWMutex
	posts      []*domain.Post
	seq        uint64
	commentSeq uint64
}

func NewMemoryPostRepository() *MemoryPostRepository {
	repo := &MemoryPostRepository{
		posts: make([]*domain.Post, 0),
	}
	repo.seedData()
	return repo
}

func (r *MemoryPostRepository) GetAll() ([]*domain.Post, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := slices.Clone(r.posts)

	slices.SortFunc(result, func(a, b *domain.Post) int {
		return cmp.Compare(b.Score, a.Score)
	})

	return result, nil
}

func (r *MemoryPostRepository) GetByID(id string) (*domain.Post, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, p := range r.posts {
		if p.ID == id {
			p.Views++
			return p, nil
		}
	}

	return nil, domain.ErrPostNotFound
}

func (r *MemoryPostRepository) GetByCategory(category string) ([]*domain.Post, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*domain.Post, 0)
	for _, p := range r.posts {
		if p.Category == category {
			result = append(result, p)
		}
	}

	slices.SortFunc(result, func(a, b *domain.Post) int {
		return cmp.Compare(b.Score, a.Score)
	})

	return result, nil
}

func (r *MemoryPostRepository) Create(p *domain.Post) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	idNum := atomic.AddUint64(&r.seq, 1) + 1000
	p.ID = fmt.Sprintf("post_%d", idNum)

	r.posts = append(r.posts, p)
	return nil
}

func (r *MemoryPostRepository) AddComment(postID string, comment *domain.Comment) (*domain.Post, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var targetPost *domain.Post
	for _, p := range r.posts {
		if p.ID == postID {
			targetPost = p
			break
		}
	}

	if targetPost == nil {
		return nil, domain.ErrPostNotFound
	}

	idNum := atomic.AddUint64(&r.commentSeq, 1) + 500
	comment.ID = fmt.Sprintf("comment_%d", idNum)
	comment.CreatedAt = time.Now().UTC()

	targetPost.Comments = append(targetPost.Comments, *comment)
	return targetPost, nil
}

func (r *MemoryPostRepository) DeleteComment(postID, commentID string, user *domain.User) (*domain.Post, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var targetPost *domain.Post
	for _, p := range r.posts {
		if p.ID == postID {
			targetPost = p
			break
		}
	}
	if targetPost == nil {
		return nil, domain.ErrPostNotFound
	}

	commentIdx := -1
	for i, c := range targetPost.Comments {
		if c.ID == commentID {
			commentIdx = i
			break
		}
	}
	if commentIdx == -1 {
		return nil, domain.ErrCommentNotFound
	}

	targetComment := targetPost.Comments[commentIdx]

	// A comment can only be deleted by its author, the author of the post, or an administrators
	if user.ID != targetComment.Author.ID && user.ID != targetPost.Author.ID && !user.Admin {
		return nil, domain.ErrUnauthorized
	}

	// Removing an element from a slice
	targetPost.Comments = append(targetPost.Comments[:commentIdx], targetPost.Comments[commentIdx+1:]...)
	return targetPost, nil
}

func (r *MemoryPostRepository) seedData() {
	user1 := domain.User{ID: "507f1f77bcf86cd799439011", Username: "alice"}
	user2 := domain.User{ID: "507f1f77bcf86cd799439012", Username: "bob"}

	now := time.Now().UTC()

	post1 := &domain.Post{
		ID:       "507f1f77bcf86cd799439013",
		Title:    "The Go programming language",
		Category: "programming",
		Type:     "link",
		URL:      "https://go.dev/",
		Author:   user1,
		Score:    2,
		Votes: []domain.Vote{
			{UserID: user1.ID, Value: 1},
			{UserID: user2.ID, Value: 1},
		},
		Comments:         make([]domain.Comment, 0),
		CreatedAt:        now.Add(-1 * time.Hour),
		Views:            5,
		UpvotePercentage: 100,
	}

	post2 := &domain.Post{
		ID:       "507f1f77bcf86cd799439014",
		Title:    "Great music tracks",
		Category: "music",
		Type:     "text",
		Text:     "What are your favorite ambient tracks for coding?",
		Author:   user2,
		Score:    1,
		Votes: []domain.Vote{
			{UserID: user2.ID, Value: 1},
		},
		Comments:         make([]domain.Comment, 0),
		CreatedAt:        now,
		Views:            2,
		UpvotePercentage: 100,
	}

	r.posts = append(r.posts, post1, post2)
}
