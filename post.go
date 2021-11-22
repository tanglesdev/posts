package posts // import "tangl.es/code/posts"

import (
	"time"

	"github.com/pkg/errors"
	"github.com/sergi/go-diff/diffmatchpatch"
)

// Post is a single, self-contained entry in a blog.
type Post struct {
	ID       string
	Title    string
	Slug     string
	Authors  []string
	Parts    []Part
	Metadata []Part
	// TODO: instead of storing timestamps here, do we want to have event logs for created, updated, publish, unpublish, etc?
	CreatedAt   time.Time
	UpdatedAt   time.Time
	PublishedAt time.Time
	// worth having draft normalized here instead of reconstructing it from
	// event logs so we can filter on it cheaply when coming up with post
	// listings
	Draft   bool
	Streams []string
}

// Part is a single part of a post, either a paragraph
// or an image, usually. It's a chunk of the post that
// it would make sense to edit atomically from the rest
// of the post.
type Part struct {
	ID       string
	Headers  map[string][]string
	Position int
	Body     []byte
	// If Inline is true, the Part should be stored
	// in the database, for quick retrieval, as it is
	// meant to be rendered as part of the initial
	// response. If Inline is false, the Part can be
	// stored in the blob store, and will be rendered
	// as a subsequent request after the Post is
	// rendered. Text is usually Inline, images are
	// usually not.
	Inline bool
	// SHA256 is the SHA 256 sum of Body. It's mostly
	// used as the filename for non-inline Parts.
	SHA256 string
}

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

// PartDelta tracks the change that occurred between the Part
// of two Posts.
type PartDelta struct {
	PartID       string
	Op           DeltaOp
	FromPosition int
	ToPosition   int
	Headers      map[string][]HeaderDelta
	Body         string
}

// HeaderDelta tracks the change that occurred between a
// specific Header in a Part.
type HeaderDelta struct {
	Op           DeltaOp
	Header       string
	FromPosition int
	ToPosition   int
	Value        string
}

// AuthorsDelta tracks the change of an Authors
// field in a Post.
type AuthorsDelta struct {
	Op           DeltaOp
	FromPosition int
	ToPosition   int
	Value        string
}

// Revision is an atomic update to a Post.
type Revision struct {
	ID string

	// whether the revision should be publicly listed
	// or a silent edit
	Public bool

	// if the revision is public, an optional explanation
	// for why the changes occurred
	Reason string

	// what changes were made in this revision
	TitleDelta     string
	SlugDelta      string
	SummaryDelta   string
	AuthorsDeltas  []AuthorsDelta
	PartsDeltas    []PartDelta
	MetadataDeltas []PartDelta
}

// diffAuthors returns the AuthorsDeltas necessary to describe
// the difference between two lists of authors.
func diffAuthors(a1, a2 []string) []AuthorsDelta {
	var deltas []AuthorsDelta
	a1Pos := make(map[string]int, len(a1))
	a2Pos := make(map[string]int, len(a2))
	for pos, author := range a1 {
		a1Pos[author] = pos
	}
	for pos, author := range a2 {
		a2Pos[author] = pos
	}
	longerAuthors := a1
	if len(a2) > len(a1) {
		longerAuthors = a2
	}
	for _, author := range longerAuthors {
		var delta AuthorsDelta
		pos1, ok := a1Pos[author]
		if !ok {
			// if we can't find the position of the
			// author in the first list, we know the
			// author was added in the second list.
			delta.Op = DeltaAdd
		}
		pos2, ok := a2Pos[author]
		if !ok {
			// if we can't find the position of the
			// author in the second list, we know the
			// author was removed from the second list.
			delta.Op = DeltaRemove
		}
		if pos1 != pos2 && delta.Op == "" {
			// if the positions don't match, and the
			// author is in both lists, we know this was
			// a move, not an addition or deletion.
			delta.Op = DeltaMove
		}
		if delta.Op == "" {
			continue
		}
		delta.FromPosition = pos1
		delta.ToPosition = pos2
		delta.Value = author
		deltas = append(deltas, delta)
	}
	return deltas
}

func diffParts(p1, p2 []Part) []PartDelta {
	var deltas []PartDelta
	p1Pos := make(map[string]int, len(p1))
	p2Pos := make(map[string]int, len(p2))
	for pos, part := range p1 {
		p1Pos[part.ID] = pos
	}
	for pos, part := range p2 {
		p2Pos[part.ID] = pos
	}
	longer := p1
	if len(p2) > len(p1) {
		longer = p2
	}
	for _, part := range longer {
		var delta PartDelta
		delta.PartID = part.ID
		pos1, ok := p1Pos[part.ID]
		if !ok {
			// if we can't find the position of the part
			// in the first post, we know the part was added
			// in the second post.
			delta.Op = DeltaAdd
		}
		pos2, ok := p2Pos[part.ID]
		if !ok {
			// if we can't find the position of the part
			// in the second post, we know the part was removed
			// in the second post.
			delta.Op = DeltaRemove
		}
		if pos1 != pos2 && delta.Op == "" {
			// if the positions aren't equal, we obviously moved
			// the part.
			delta.Op = DeltaMove
		}
		part1, part2 := p1[pos1], p2[pos2]
		if delta.Op != DeltaAdd && delta.Op != DeltaRemove {
			// if we're not adding, not deleting, we may still
			// need to modify in place.
			if string(part1.Body) != string(part2.Body) {
				// need to check if we're already moving, in
				// which case this is a move and update, not
				// just a move.
				if delta.Op == DeltaMove {
					delta.Op = DeltaMoveUpdate
				} else {
					delta.Op = DeltaUpdate
				}
			}
		}
		delta.Headers = diffHeaders(part1.Headers, part2.Headers)
		if len(delta.Headers) != 0 && delta.Op == "" {
			delta.Op = DeltaUpdate
		} else if len(delta.Headers) != 0 && delta.Op == DeltaMove {
			delta.Op = DeltaMoveUpdate
		}
		if delta.Op != "" {
			// if there's any change at all, we want to record
			// the old position, the new position, and the change
			// the body went through.
			delta.FromPosition = pos1
			delta.ToPosition = pos2
			delta.Body = deltaFromStrings(string(part1.Body), string(part2.Body))
			deltas = append(deltas, delta)
		}
	}
	return deltas

}

func diffHeaders(h1, h2 map[string][]string) map[string][]HeaderDelta {
	deltas := map[string][]HeaderDelta{}
	headers := map[string]struct{}{}
	for header := range h1 {
		headers[header] = struct{}{}
	}
	for header := range h2 {
		headers[header] = struct{}{}
	}
	for header := range headers {
		headerLonger := h1[header]
		if len(h2[header]) > len(headerLonger) {
			headerLonger = h2[header]
		}
		hpos1 := make(map[string]int, len(h1[header]))
		for pos, val := range h1[header] {
			hpos1[val] = pos
		}
		hpos2 := make(map[string]int, len(h2[header]))
		for pos, val := range h2[header] {
			hpos2[val] = pos
		}
		for _, h := range headerLonger {
			var headerDelta HeaderDelta
			pos, ok := hpos1[h]
			if !ok {
				// we don't have a position in the first
				// part's headers, so it must be newly
				// added
				headerDelta.Op = DeltaAdd
			}
			headerDelta.FromPosition = pos
			pos, ok = hpos2[h]
			if !ok {
				// we don't have a position in the second
				// part's headers, so it must be removed
				headerDelta.Op = DeltaRemove
			}
			headerDelta.ToPosition = pos
			if headerDelta.FromPosition != headerDelta.ToPosition && headerDelta.Op == "" {
				headerDelta.Op = DeltaMove
			}
			headerDelta.Value = h
			if headerDelta.Op != "" {
				deltas[header] = append(deltas[header], headerDelta)
			}
		}
	}
	return deltas

}

// GenRevision creates a Revision based on the two Posts.
// Note that GenRevision is not commutative, so the order
// of the two posts matters. It is the caller's responsibility
// to ensure the order of the Posts is consistent, in order
// to obtain meaningful Revisions.
func GenRevision(p1, p2 Post) (Revision, error) {
	var rev Revision
	if p1.ID != p2.ID {
		return rev, errors.New("post IDs must match")
	}
	if p1.Title != p2.Title {
		rev.TitleDelta = deltaFromStrings(p1.Title, p2.Title)
	}
	if p1.Slug != p2.Slug {
		rev.SlugDelta = deltaFromStrings(p1.Slug, p2.Slug)
	}
	rev.AuthorsDeltas = diffAuthors(p1.Authors, p2.Authors)
	rev.PartsDeltas = diffParts(p1.Parts, p2.Parts)
	// TODO: diff metadata
	return rev, nil
}

// get the compact delta format diff between two strings
func deltaFromStrings(str1, str2 string) string {
	dmp := diffmatchpatch.New()
	// find the differences between the strings
	diffs := dmp.DiffMain(str1, str2, true)
	// clean the diffs up to be minimal, semantic diffs
	diffs = dmp.DiffCleanupSemanticLossless(diffs)
	// converts the diffs to compact delta format. E.g.
	// =3\t-2\t+ing -> Keep 3 chars, delete 2 chars,
	// insert 'ing'.
	return dmp.DiffToDelta(diffs)
}
