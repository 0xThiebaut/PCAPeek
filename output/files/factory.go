package files

import (
	"io"
	"time"
)

type Stream interface {
	io.Closer
	Write(reader io.Reader, t time.Time) error
}

type Factory interface {
	New() Stream
}
