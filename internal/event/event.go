package event

import "oss/pkg/event_bus"

const (
	EventTypeObjectCreated  = "s3:ObjectCreated:*"
	EventTypeObjectRemoved  = "s3:ObjectRemoved:*"
	EventTypeObjectAccessed = "s3:ObjectAccessed:*"
	EventTypeObjectDownload = "s3:ObjectDownload:*"

	EventTypeBucketCreated  = "s3:BucketCreated:*"
	EventTypeBucketRemoved  = "s3:BucketRemoved:*"
	EventTypeBucketAccessed = "s3:BucketAccessed:*"
)

var Bus *event_bus.EventBus

func init() {
	Bus = event_bus.NewEventBus(1024)
}
