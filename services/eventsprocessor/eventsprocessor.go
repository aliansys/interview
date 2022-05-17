package eventsprocessor

import (
	"context"
	"github.com/aliansys/interview/domain/dtos"
	"github.com/aliansys/interview/parsers/dtos/events"
	"github.com/valyala/fastjson"
	"log"
	"runtime"
)

type (
	Storage interface {
		Store(dtos.EnrichmentEvents)
		Close()
	}

	saver struct {
		storage Storage

		ctx    context.Context
		cancel context.CancelFunc

		incoming chan dtos.RawEnrichmentEvents
		logger   *log.Logger
	}
)

// numberOfWorkers - разделен на 2, чтобы половину "воркеров" дать на парсинг и половину на сохранение
var numberOfWorkers = runtime.NumCPU() / 2

func New(s Storage, l *log.Logger) *saver {
	ctx, cancel := context.WithCancel(context.Background())
	saver := &saver{
		storage: s,
		ctx:     ctx,
		cancel:  cancel,

		incoming: make(chan dtos.RawEnrichmentEvents, numberOfWorkers),
		logger:   l,
	}

	saver.run()
	return saver
}

func (s *saver) run() {
	output := make(chan dtos.EnrichmentEvents, numberOfWorkers)

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

func (s *saver) Process(events dtos.RawEnrichmentEvents) {
	s.incoming <- events
}

func (s *saver) parse(output chan dtos.EnrichmentEvents) {
	p := &fastjson.Parser{}

	for {
		select {
		case rawEvents := <-s.incoming:
			evnts, err := events.Parse(p, rawEvents.Events)
			if err != nil {
				// здесь не до конца понятно насколько нам критичны подобного рода ошибки
				// как минимум такое можно залогировать в систему мониторинга ошибок (условная Sentry)
				s.logger.Printf("parsing error: %s\n", err)
				continue
			}

			if len(evnts) == 0 {
				continue
			}

			output <- dtos.EnrichmentEvents{
				Events: evnts,
				IP:     rawEvents.IP,
			}
		case <-s.ctx.Done():
			return
		}
	}
}

func (s *saver) save(events chan dtos.EnrichmentEvents) {
	for {
		select {
		case es := <-events:
			s.storage.Store(es)
		case <-s.ctx.Done():
			return
		}
	}
}
