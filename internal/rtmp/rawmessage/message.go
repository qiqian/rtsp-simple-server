// Package rawmessage contains a RTMP raw message reader/writer.
package rawmessage

import (
	"time"

	"github.com/aler9/rtsp-simple-server/internal/rtmp/chunk"
)

// Message is a raw message.
type Message struct {
	Typ             byte
	ChunkStreamID   int
	Timestamp       time.Duration
	Type            chunk.MessageType
	MessageStreamID uint32
	Body            []byte
}
