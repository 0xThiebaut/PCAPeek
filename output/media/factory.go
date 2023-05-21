package media

import (
	"image"
	"io"
	"time"
)

type Stream interface {
	io.Closer
	Write(image image.Image, t time.Time) error
}

type Factory interface {
	New() Stream
}
