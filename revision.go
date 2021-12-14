package posts

// DeltaOp is the type of change that is happenging to a
// Part. It can be added, removed, updated, moved, or
// moved and updated.
type DeltaOp string

const (
	// DeltaAdd is a signifier that a Part is being
	// added to a Post.
	DeltaAdd DeltaOp = "add"

	// DeltaRemove is a signifier that a Part is being
	// removed from a Post.
	DeltaRemove DeltaOp = "rm"

	// DeltaUpdate is a signifier that the contents of
	// a Part are being updated.
	DeltaUpdate DeltaOp = "up"

	// DeltaMove is a signifier that the position of a
	// Part within the Post is being updated.
	DeltaMove DeltaOp = "mv"

	// DeltaMoveUpdate is a signifier that the position
	// of a Part within the Post is being updated, and
	// also that the contents of that Part are being
	// updated.
	DeltaMoveUpdate DeltaOp = "mvup"
)

// Revision is an atomic update to a Post.
type Revision struct {
	// ID is a UUID suitable for uniquely identifying a revision.
	ID string

	// Public tracks whether the revision should be publicly visible or is
	// a silent edit.
	Public bool

	// Reason indicates why the revision was made, when a revision is
	// public.
	Reason string

	// TitleDelta contains a diff of the post's title before the revision
	// and after the revision, such that patching the post's title before
	// the revision with TitleDelta will result in the post's title after
	// the revision.
	TitleDelta string

	// SlugDelta contains a diff of the post's slug before the revision and
	// after the revision, such that patching the post's slug before the
	// revision with SlugDelta will result in the post's slug after the
	// revision.
	SlugDelta string

	// AuthorsDeltas describes a set of changes to the collection of
	// authors for the post.
	AuthorsDeltas []AuthorsDelta

	// PartsDeltas describes a set of changes to the parts of the post
	// body.
	PartsDeltas []PartDelta

	// MetadataDeltas describes a set of changes to the metadata of a post.
	MetadataDeltas []PartDelta
}

// PartDelta tracks the change that occurred between two versions of a Part.
type PartDelta struct {
	// PartID records the ID of the part that the change being described
	// applies to.
	PartID string

	// Op indicates the type of change being described.
	Op DeltaOp

	// FromPosition indicates the position the part started in. It must
	// always be set, even when Op is not DeltaMove or DelteMoveUpdate.
	FromPosition int

	// ToPosition indicates the position the part ended up in. It must
	// always be set, even when Op is not DeltaMove or DeltaMoveUpdate. In
	// these situations, it should match FromPosition.
	ToPosition int

	// Headers tracks the change to the headers of the part.
	Headers map[string][]HeaderDelta

	// Body is a textual diff of the change between the two parts, suitable
	// for patching the first part to match the second part.
	//
	// When an inline part becomes a non-inline part, this will be a diff
	// for deleting the entire body of the part.
	//
	// When a non-inline part becomes an inline part, this will be a diff
	// for creating the entire body of the part.
	//
	// This will be empty for non-inline parts that remain non-inline
	// parts; instead, SHA256From and SHA256To will record those changes.
	Body string

	// SHA256From describes the SHA256 hash the part started with. This is
	// used in lieu of Body for non-inline parts that are stored in blob
	// storage. If this is set, it means the first
	SHA256From string

	// SHA256To describes the SHA256 hash the part ended with. This is used
	// in lieu of Body for non-inline parts that are stored in blob
	// storage.
	SHA256To string
}

// HeaderDelta tracks the change that occurred between a
// specific Header in a Part.
type HeaderDelta struct {
	// Op indicates the type of change being described.
	Op DeltaOp

	// Header indicates the key of the headers map being changed.
	Header string

	// FromPosition indicates the original position of the value in the
	// header's list of values. It must always be set, even when Op is not
	// DeltaMove or DeltaMoveUpdate.
	FromPosition int

	// ToPosition indicates the final position of the value in the header's
	// list of values. It must always be set, even when Op is not DeltaMove
	// or DeltaMoveUpdate. In these situations, it should match
	// FromPosition.
	ToPosition int

	// Value is a textual diff of the two header values, suitable for
	// patching the original value to match the final value.
	Value string
}

// AuthorsDelta tracks the change of an Authors
// field in a Post.
//
// No diffing is done on values, as values are opaque IDs. So the Op should
// never be set to DeltaUpdate or DeltaMoveUpdate. Instead, it should always be
// DeltaAdd, DeltaRemove, or DeltaMove.
type AuthorsDelta struct {
	// Op indicates the type of change being described.
	Op DeltaOp

	// FromPosition indicates the original position of the author in the
	// list of authors. It must always be set, even when Op is not
	// DeltaMove or DeltaMoveUpdate.
	FromPosition int

	// ToPosition indicates the final position of the author in the list of
	// authors. It must always be set, even when Op is not DeltaMove. In
	// that situation, it should match FromPosition.
	ToPosition int
}
