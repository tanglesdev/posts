package posts

import (
	"time"
)

// Post is a single, self-contained entry in a stream.
type Post struct {
	// ID is a UUID uniquely identifying the post.
	ID string

	// Title is a human-friendly name title for the post, suitable for
	// display.
	Title string

	// Slug is a human-friendly URL component identifying the post, usually
	// an encoding of the Title.
	Slug string

	// Authors is a list of IDs for the authors that wrote or contributed
	// to the post.
	Authors []string

	// Parts is a collection of pieces that make up the post, each having
	// their own content type and headers. All parts are expected to be
	// rendered as part of the post body in at least one view.
	Parts []Part

	// Metadata is a collection of information about the post, like its
	// summary, that should not be rendered as part of the post body but
	// may be surfaced elsewhere.
	Metadata []Part

	// Streams contains the IDs of the streams that the post is in.
	Streams []string

	// Draft indicates whether the post is currently unpublished or not.
	// It's worth having draft status normalized here instead of
	// reconstructing it from event logs so we can filter on it cheaply
	// when coming up with post listings.
	Draft bool

	// Deleted indicates whether the post is currently soft-deleted or not.
	// It's worth having deletion status normalized here instead of
	// reconstructing it from event logs so we can filter on it cheaply
	// when coming up with post listings.
	Deleted bool

	// PublishedAt indicates the last time the post was marked as pubished.
	// It's worth having the latest publication timestamp normalized here
	// instead of reconstructing it from event logs so we can filter and
	// sort on it cheaply when coming up with post listings.
	PublishedAt time.Time
}

// Part is a single part of a post, either a paragraph
// or an image, usually. It's a chunk of the post that
// it would make sense to edit atomically from the rest
// of the post. It can have its own metadata like content type and rendering
// options.
type Part struct {
	// ID is a UUID capable of uniquely identifying the part.
	ID string

	// Headers contain metadata about the part, including its content type
	// and any rendering parameters.
	Headers map[string][]string

	// Position indicates the order of the part in the post.
	Position int

	// Body is the content of the part.
	Body []byte

	// Inline indicates if the Part should be stored in the database, for
	// quick retrieval, as it is meant to be rendered as part of the
	// initial response. If Inline is false, the Part can be stored in a
	// blob store, and will be rendered as a subsequent request after the
	// Post is rendered. Text is usually Inline, images and other media are
	// usually not.
	Inline bool

	// SHA256 is the SHA 256 sum of Body. It's mostly
	// used as the filename for non-inline Parts.
	SHA256 string
}
