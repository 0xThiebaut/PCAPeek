package jpeg

import (
	"github.com/0xThiebaut/PCAPeek/output/media"
)

func New(directory string, fps int, quality int) media.Factory {
	return &factory{
		Directory: directory,
		FPS:       fps,
		Quality:   quality,
	}
}

type factory struct {
	Directory string
	id        int
	FPS       int
	Quality   int
}

func (f *factory) New() media.Stream {
	s := &stream{
		Directory: f.Directory,
		ID:        f.id,
		FPS:       f.FPS,
		Quality:   f.Quality,
	}
	f.id++
	return s
}
