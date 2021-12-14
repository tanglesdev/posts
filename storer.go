package posts

import (
	"context"
	"time"
)

// Storer captures the interface for storing and retrieving post contents in a
// database of some kind.
type Storer interface {
	// Create persists the Post as it is, returning an error if any
	// necessary fields are missing or if the Post can't be written.
	Create(ctx context.Context, post Post) error

	// Update applies the specified Revision to the Post indicated by the
	// passed postID.
	Update(ctx context.Context, postID string, rev Revision) error

	// Delete marks the Post indicated by the passed ID as deleted,
	// returning the Post that was deleted.
	Delete(ctx context.Context, id string) error

	// Get retrieves a Post by its ID, returning an error if it can't be
	// found.
	Get(ctx context.Context, id string) (Post, error)

	// List retrieves an list of Posts sorted by their PublishedAt property
	// descending, filtered according to the passed filter.
	List(ctx context.Context, filter PostFilter) ([]Post, error)
	// TODO: query, for full-text search?
}

// StringListFilterMode is an enum for indicating how a list of strings should
// be interpreted when filtering.
type StringListFilterMode string

const (
	// StringListFilterModeInvalid is an invalid value placeholder that
	// should never be intentionally used.
	StringListFilterModeInvalid StringListFilterMode = ""

	// StringListFilterModeExact filters for Posts that match those strings
	// exactly, with no deviation on order or contents.
	StringListFilterModeExact StringListFilterMode = "exact"

	// StringListFilterModeExactUnordered filters for Posts that match
	// those strings exactly, with no deviation in contents, but without
	// needing to match the order exactly.
	StringListFilterModeExactUnordered StringListFilterMode = "exact_unordered"

	// StringListFilterModeContainsAll filters for Posts that contain all
	// those strings, but maybe not in that order, and other values may
	// also be present in the list.
	StringListFilterModeContainsAll StringListFilterMode = "contains_all"

	// StringListFilterModeContainsAny filters for Posts that contain any
	// of those strings, although maybe not in that order, and other values
	// may be present in the list.
	StringListFilterModeContainsAny StringListFilterMode = "contains_any"

	// StringListFilterModeExcludes filters for Posts that contain none of
	// the strings in the list, regardless of order.
	StringListFilterModeExcludes StringListFilterMode = "excludes"
)

// PostFilter represents a filter that can be applied to Posts to return only
// the Posts the caller is interested in.
type PostFilter struct {
	// Slug, when non-nil, filters for Posts with a slug that matches its
	// value.
	Slug *string

	// Authors, when non-nil and non-empty, filters for Posts with authors
	// that match its value, where AuthorsMode controls how "match" is
	// defined.
	Authors []string

	// AuthorsMode specifies the type of values that will be considered a
	// match for the Authors property.
	AuthorsMode StringListFilterMode

	// PublishedBefore specifies the maximum timestamp, exclusive, that
	// Posts should have in their PublishedAt property.
	PublishedBefore *time.Time

	// PublishedAfter specifies the minimum timestamp, exclusive, that
	// Posts should have in their PublishedAt property.
	PublishedAfter *time.Time

	// Draft, when non-nil, filters out Posts with a Draft property
	// different than its value.
	Draft *bool

	// Streams, when non-nil and non-empty, filters for Posts with streams
	// that match its value, where StreamsMode controls how "match" is
	// defined.
	Streams []string

	// StreamsMode specifies the type of values that will be considered a
	// match for the Streams property.
	StreamsMode StringListFilterMode
}

// IsEmpty returns true if the PostFilter is semantically an empty value, i.e.,
// not set.
func (p PostFilter) IsEmpty() bool {
	if p.Slug != nil {
		return false
	}
	if len(p.Authors) != 0 {
		return false
	}
	if p.AuthorsMode == StringListFilterModeInvalid {
		return false
	}
	if p.PublishedBefore != nil {
		return false
	}
	if p.PublishedAfter != nil {
		return false
	}
	if p.Draft != nil {
		return false
	}
	if len(p.Streams) != 0 {
		return false
	}
	if p.StreamsMode == StringListFilterModeInvalid {
		return false
	}
	return true
}
