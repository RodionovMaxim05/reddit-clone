package domain

import "time"

type Vote struct {
	UserID string
	Value  int
}

type Comment struct {
	ID        string
	Author    User
	Body      string
	CreatedAt time.Time
}

type Post struct {
	ID               string
	Title            string
	Category         string
	Type             string
	URL              string
	Text             string
	Author           User
	Score            int
	Votes            []Vote
	Comments         []Comment
	CreatedAt        time.Time
	Views            int
	UpvotePercentage int
}
