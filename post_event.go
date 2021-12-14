package posts

import "time"

// PostEventType is an enum of different types of events that can happen to a
// post.
type PostEventType string

const (
	// PostEventTypeCreated is used when an event is recording that a post
	// was created.
	PostEventTypeCreated PostEventType = "created"
	// PostEventTypeUpdated is used when an event is recording that a post
	// was updated.
	PostEventTypeUpdated PostEventType = "updated"
	// PostEventTypeDeleted is used when an event is recording that a post
	// was deleted.
	PostEventTypeDeleted PostEventType = "deleted"
	// PostEventTypePublished is used when an event is recording that a
	// draft post was published.
	PostEventTypePublished PostEventType = "published"
	// PostEventTypeUnpublished is used when an event is recording that a
	// published post was reverted to a draft.
	PostEventTypeUnpublished PostEventType = "unpublished"
)

// PostEventActorType is an enum of different types of actors that can take
// actions on a post.
type PostEventActorType string

const (
	// PostEventActorTypeUser is used when a human logs in and takes a
	// manual action on a post.
	PostEventActorTypeUser PostEventActorType = "user"
	// PostEventActorTypeSystem is used when Tangles takes an automated
	// action on a post.
	PostEventActorTypeSystem PostEventActorType = "system"
)

// PostEvent records an action that was taken on a post.
type PostEvent struct {
	// A UUID for this event.
	ID string
	// The type of the event, describing what happened.
	Type PostEventType
	// The IP the action was taken from, for audit purposes.
	IP string
	// The ID of the actor that took the action.
	Actor string
	// The type of actor that took the action.
	ActorType PostEventActorType
	// some session management strategies, like Lockbox, will let us work
	// backwards from a session ID to how the user started the session.
	// This is good audit information to have in case of a breach.
	SessionID string
	// The date and time the action was taken.
	Timestamp time.Time
}
