package posts

import (
	"context"
	"time"
)

type Storer interface {
	Create(ctx context.Context, post Post) error
	Update(ctx context.Context, post Post, rev Revision) error
	Delete(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (Post, error)
	List(ctx context.Context, filter PostFilter) ([]Post, error)
}

type StringListFilterMode uint

const (
	StringListFilterModeInvalid = iota
	StringListFilterModeExact
	StringListFilterModeContains
	StringListFilterModeExcludes
)

// TODO: can we do better than this? Can we offer something similar in power to GMail's search syntax? Is that even a good idea, operationally speaking?
type PostFilter struct {
	Title           *string
	Slug            *string
	Authors         []string
	AuthorsMode     StringListFilterMode
	PublishedBefore *time.Time
	PublishedAfter  *time.Time
	PublishedAt     *time.Time
	Draft           *bool
	Streams         []string
	StreamsMode     StringListFilterMode
}

func (p PostFilter) IsEmpty() bool {
	if p.Title != nil {
		return false
	}
	if p.Slug != nil {
		return false
	}
	if p.Summary != nil {
		return false
	}
	if p.Author != nil {
		return false
	}
	if p.PublishedBefore != nil {
		return false
	}
	if p.PublishedAfter != nil {
		return false
	}
	if p.PublishedAt != nil {
		return false
	}
	if p.Draft != nil {
		return false
	}
	if p.Stream != nil {
		return false
	}
	return true
}
