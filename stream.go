package posts

// A Stream is a series of posts. This struct
// holds the metadata about a stream.
type Stream struct {
	ID          string
	Title       string
	Description string
	Authors     []string
}
