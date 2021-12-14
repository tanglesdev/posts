package posts

// A Stream is a series of posts. This struct
// holds the metadata about a stream.
type Stream struct {
	// ID is a UUID capable of uniquely identifying the stream.
	ID string

	// Title is a human-friendly description of the stream.
	Title string

	// Slug is a URL-friendly encoding of the title, for use in URLs.
	Slug string

	// Metadata is a collection of parts that can be used to collect
	// arbitrary rendering information for the stream, like header images.
	// These can then be accessed from templates when rendering the stream.
	Metadata []Part

	// Authors is a collection of the IDs of the users that can write to
	// this stream.
	Authors []string
}
