package binary

import (
	"github.com/0xThiebaut/PCAPeek/output/files"
)

func New(directory string) files.Factory {
	return &factory{
		Directory: directory,
	}
}

type factory struct {
	Directory string
	id        int
}

func (f *factory) New() files.Stream {
	s := &stream{
		Directory: f.Directory,
		ID:        f.id,
	}
	f.id++
	return s
}
