package fork

import (
	"github.com/0xThiebaut/PCAPeek/output/files"
)

func New(factories ...files.Factory) files.Factory {
	return &factory{
		Factories: factories,
	}
}

type factory struct {
	Factories []files.Factory
}

func (f *factory) New() files.Stream {
	o := &stream{}
	for _, fork := range f.Factories {
		o.Streams = append(o.Streams, fork.New())
	}
	return o
}
