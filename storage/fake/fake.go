package fake

import (
	"github.com/aliansys/interview/domain/dtos"
	"log"
)

type fake struct {
	saved  int
	logger *log.Logger
}

func New(l *log.Logger) (*fake, error) {
	return &fake{
		logger: l,
	}, nil
}

func (f *fake) Close() {}

func (f *fake) Store(_ dtos.EventsWithIP) {
	f.saved += 1
	f.logger.Printf("total events saved %d\n", f.saved)
}
