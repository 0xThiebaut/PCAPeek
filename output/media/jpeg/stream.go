package jpeg

import (
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"path"
	"time"
)

type stream struct {
	Directory string
	ID        int
	Quality   int
	FPS       int
	sequence  int
	next      time.Time
	frame     image.Image
}

func (s *stream) Write(image image.Image, t time.Time) error {
	if s.next.IsZero() {
		name := fmt.Sprintf("%s.%02d", t.UTC().Format("2006-01-02T15-04-05,000"), s.ID)
		s.Directory = path.Join(s.Directory, name)
		if err := os.MkdirAll(s.Directory, 0o666); err != nil {
			return err
		}
	}
	if s.FPS != 0 {
		if s.next.IsZero() {
			s.next = t.Add(time.Second / time.Duration(s.FPS))
		}
		for s.next.Before(t) {
			if err := s.write(); err != nil {
				return err
			}
			s.next = s.next.Add(time.Second / time.Duration(s.FPS))
		}
	}
	s.frame = image
	if s.FPS == 0 {
		s.next = t
		return s.write()
	}
	return nil
}

func (s *stream) write() error {
	name := fmt.Sprintf("%06d.%s.jpeg", s.sequence, s.next.UTC().Format("2006-01-02T15-04-05,000"))
	f, err := os.OpenFile(path.Join(s.Directory, name), os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0o666)
	if err != nil {
		return err
	}
	defer f.Close()
	s.sequence++
	return jpeg.Encode(f, s.frame, &jpeg.Options{Quality: s.Quality})
}

func (s *stream) Close() error {
	if s.FPS != 0 {
		return s.write()
	}
	return nil
}
