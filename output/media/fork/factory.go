package fork

import (
	"github.com/0xThiebaut/PCAPeek/output/media"
)

func New(factories ...media.Factory) media.Factory {
	return &factory{
		Factories: factories,
	}
}

type factory struct {
	Factories []media.Factory
}

func (f *factory) New() media.Stream {
	o := &stream{}
	for _, fork := range f.Factories {
		o.Streams = append(o.Streams, fork.New())
	}
	return o
}
