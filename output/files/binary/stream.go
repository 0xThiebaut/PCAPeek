package binary

import (
	"fmt"
	"io"
	"os"
	"path"
	"time"
)

type stream struct {
	Directory string
	ID        int
	sequence  int
}

func (s *stream) Write(reader io.Reader, t time.Time) error {
	name := fmt.Sprintf("%s.%02d.bin", t.UTC().Format("2006-01-02T15-04-05,000"), s.ID)
	f, err := os.OpenFile(path.Join(s.Directory, name), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o666)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, reader)
	return err
}

func (s *stream) Close() error {
	return nil
}
