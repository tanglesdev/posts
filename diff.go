package posts

import (
	"bytes"
	"errors"

	"github.com/sergi/go-diff/diffmatchpatch"
)

// diffAuthors returns the AuthorsDeltas necessary to describe the difference
// between two lists of authors.
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
	// TODO: still missing any authors that were in the shorter list but
	// not in the longer one
	for _, author := range longerAuthors {
		var delta AuthorsDelta
		pos1, ok := a1Pos[author]
		if !ok {
			// if we can't find the position of the author in the
			// first list, we know the author was added in the
			// second list.
			delta.Op = DeltaAdd

			// position of -1 indicates "not present"
			pos1 = -1
		}
		pos2, ok := a2Pos[author]
		if !ok {
			// if we can't find the position of the author in the
			// second list, we know the author was removed from the
			// second list.
			delta.Op = DeltaRemove

			// position of -1 indicates "not present"
			pos2 = -1
		}
		if pos1 != pos2 && delta.Op == "" {
			// if the positions don't match, and the author is in
			// both lists, we know this was a move, not an addition
			// or deletion.
			delta.Op = DeltaMove
		}
		if delta.Op == "" {
			// if we're not adding, removing, or moving an author
			// around, we're not doing anything to them, skip this.
			continue
		}
		delta.FromPosition = pos1
		delta.ToPosition = pos2
		deltas = append(deltas, delta)
	}
	return deltas
}

// diffParts returns the PartDeltas necessary to describe the difference
// between two lists of parts.
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
	// TODO: still missing any parts that were in the shorter list but not
	// in the longer one
	for _, part := range longer {
		var delta PartDelta
		delta.PartID = part.ID
		pos1, ok := p1Pos[part.ID]
		if !ok {
			// if we can't find the position of the part in the
			// first list, we know the part was added in the second
			// list.
			delta.Op = DeltaAdd
		}
		pos2, ok := p2Pos[part.ID]
		if !ok {
			// if we can't find the position of the part in the
			// second list, we know the part was removed in the
			// second list.
			delta.Op = DeltaRemove
		}
		if pos1 != pos2 && delta.Op == "" {
			// if the positions aren't equal, we obviously moved
			// the part.
			delta.Op = DeltaMove
		}
		part1, part2 := p1[pos1], p2[pos2]
		if delta.Op != DeltaAdd && delta.Op != DeltaRemove {
			// if we're not adding, not deleting, we may still need
			// to modify in place.
			if bytes.Equal(part1.Body, part2.Body) {
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
			// if there's any change at all, we want to record the
			// old position, the new position, and the change the
			// body went through.
			delta.FromPosition = pos1
			delta.ToPosition = pos2

			// if part1 isn't inline, we want to record that in the
			// SHA256From field so we know what SHA256 the
			// non-inline part had at the start. We don't want to
			// record those bytes in the database.
			if !part1.Inline {
				delta.SHA256From = part1.SHA256
			}

			// if part2 isn't inline, we want to record that in the
			// SHA256TO field so we know what SHA256 the non-inline
			// part had at the end. We don't want to record those
			// bytes in the database.
			if !part2.Inline {
				delta.SHA256To = part1.SHA256
			}

			// if part1 is inline and part2 isn't, we're swapping
			// an inline part for a non-inline part. We record this
			// as a patch for deleting the inline body, and rely on
			// the SHA256To (which has already been set) to
			// indicate the new content.
			if part1.Inline && !part2.Inline {
				delta.Body = deltaFromStrings("", string(part2.Body))
			}

			// if part1 isn't inline and part2 is, we're swapping a
			// non-inline part for an inline part. We reord this as
			// a patch for creating the inline body, and rely on
			// the SHA256From (which has already been set) to
			// indicate the old content.
			if !part1.Inline && part2.Inline {
				delta.Body = deltaFromStrings(string(part1.Body), "")
			}

			// if both parts are inline, we're doing a straight
			// text update, and we just want to record the patch of
			// that.
			if part1.Inline && part2.Inline {
				delta.Body = deltaFromStrings(string(part1.Body), string(part2.Body))
			}
			deltas = append(deltas, delta)
		}
	}
	return deltas

}

// diffHeaders returns the HeaderDeltas necessary to describe the difference
// between two header maps.
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
		// TODO: still missing any values that were in the shorter list
		// but not in the longer one
		for _, h := range headerLonger {
			var headerDelta HeaderDelta
			pos, ok := hpos1[h]
			if !ok {
				// we don't have a position in the first part's
				// headers, so it must be newly added
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

// GenerateRevision creates a Revision based on the two Posts. Note that
// GenerateRevision is not commutative, so the order of the two posts matters.
// It is the caller's responsibility to ensure the order of the Posts is
// consistent, in order to obtain meaningful Revisions. As a general rule of
// thumb, the posts should be in ascending chronological order.
func GenerateRevision(p1, p2 Post) (Revision, error) {
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
	rev.MetadataDeltas = diffParts(p1.Metadata, p2.Metadata)
	return rev, nil
}

// get the compact delta format diff between two strings
func deltaFromStrings(str1, str2 string) string {
	dmp := diffmatchpatch.New()

	// find the differences between the strings
	diffs := dmp.DiffMain(str1, str2, true)

	// clean the diffs up to be minimal, semantic diffs
	diffs = dmp.DiffCleanupSemanticLossless(diffs)

	// converts the diffs to compact delta format. E.g.:
	//
	// =3\t-2\t+ing -> Keep 3 chars, delete 2 chars, insert 'ing'.
	return dmp.DiffToDelta(diffs)
}
