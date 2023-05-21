package fork

import (
	"image"
	"time"

	"github.com/0xThiebaut/PCAPeek/output/media"
)

type stream struct {
	Streams []media.Stream
}

func (s *stream) Write(image image.Image, t time.Time) error {
	for _, forked := range s.Streams {
		if err := forked.Write(image, t); err != nil {
			return err
		}
	}
	return nil
}

func (s *stream) Close() (err error) {
	for _, forked := range s.Streams {
		if ferr := forked.Close(); ferr != nil && err == nil {
			err = ferr
		}
	}
	return err
}
