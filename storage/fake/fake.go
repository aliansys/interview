package fake

import (
	"fmt"
	"github.com/aliansys/interview/domain/dtos"
)

type fake struct {
	saved int
}

func New() (*fake, error) {
	return new(fake), nil
}

func (f *fake) Close() {}

func (f *fake) Store(_ dtos.EventsWithIP) {
	f.saved += 1
	fmt.Printf("total events saved %d\n", f.saved)
}
