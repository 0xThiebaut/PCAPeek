package fork

import (
	"bytes"
	"io"
	"time"

	"github.com/0xThiebaut/PCAPeek/output/files"
)

type stream struct {
	Streams []files.Stream
}

func (s *stream) Write(reader io.Reader, t time.Time) error {
	d, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	for _, forked := range s.Streams {
		if err = forked.Write(bytes.NewReader(d), t); err != nil {
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
