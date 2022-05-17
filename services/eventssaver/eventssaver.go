package eventssaver

import (
	"context"
	"fmt"
	"github.com/aliansys/interview/domain/dtos"
	"github.com/aliansys/interview/parsers/dtos/events"
	"github.com/valyala/fastjson"
	"runtime"
)

type (
	Repo interface {
		Store(evnts dtos.EventsWithIP)
		Close()
	}

	saver struct {
		repo Repo

		ctx    context.Context
		cancel context.CancelFunc

		incoming chan dtos.RawEventWithIP
	}
)

// numberOfWorkers - разделен на 2, чтобы половину "воркеров" дать на парсинг и половину на сохранение
var numberOfWorkers = runtime.NumCPU() / 2

func New(r Repo) *saver {
	ctx, cancel := context.WithCancel(context.Background())
	saver := &saver{
		repo:   r,
		ctx:    ctx,
		cancel: cancel,

		incoming: make(chan dtos.RawEventWithIP, numberOfWorkers),
	}

	saver.run()
	return saver
}

func (s *saver) run() {
	output := make(chan dtos.EventsWithIP, numberOfWorkers)

	for i := 0; i < numberOfWorkers; i++ {
		go s.parse(output)
	}

	for i := 0; i < numberOfWorkers; i++ {
		go s.save(output)
	}
}

func (s *saver) Close() {
	if s.cancel != nil {
		s.cancel()
	}
}

func (s *saver) Process(events dtos.RawEventWithIP) {
	s.incoming <- events
}

func (s *saver) parse(output chan dtos.EventsWithIP) {
	p := &fastjson.Parser{}

	for {
		select {
		case rawEvents := <-s.incoming:
			evnts, err := events.Parse(p, rawEvents.Events)
			if err != nil {
				// здесь не до конца понятно насколько нам критичны подобного рода ошибки
				// как минимум такое можно залогировать в систему мониторинга ошибок (условная Sentry)
				fmt.Printf("parsing error: %s\n", err)
				continue
			}

			if len(evnts) == 0 {
				continue
			}

			output <- dtos.EventsWithIP{
				Events: evnts,
				IP:     rawEvents.IP,
			}
		case <-s.ctx.Done():
			return
		}
	}
}

func (s *saver) save(events chan dtos.EventsWithIP) {
	for {
		select {
		case es := <-events:
			s.repo.Store(es)
		case <-s.ctx.Done():
			return
		}
	}
}
