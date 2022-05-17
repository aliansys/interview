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

	processor struct {
		storage Storage

		ctx    context.Context
		cancel context.CancelFunc

		incoming chan dtos.RawEnrichmentEvents
		logger   *log.Logger
	}
)

// numberOfWorkers - разделен на 2, чтобы половину "воркеров" дать на парсинг и половину на сохранение
var numberOfWorkers = runtime.NumCPU() / 2

func New(s Storage, l *log.Logger) *processor {
	ctx, cancel := context.WithCancel(context.Background())
	saver := &processor{
		storage: s,
		ctx:     ctx,
		cancel:  cancel,

		incoming: make(chan dtos.RawEnrichmentEvents, numberOfWorkers),
		logger:   l,
	}

	saver.run()
	return saver
}

func (p *processor) run() {
	output := make(chan dtos.EnrichmentEvents, numberOfWorkers)

	for i := 0; i < numberOfWorkers; i++ {
		go p.parse(output)
	}

	for i := 0; i < numberOfWorkers; i++ {
		go p.save(output)
	}
}

func (p *processor) Close() {
	if p.cancel != nil {
		p.cancel()
	}
}

func (p *processor) Process(events dtos.RawEnrichmentEvents) {
	p.incoming <- events
}

func (p *processor) parse(output chan dtos.EnrichmentEvents) {
	parser := &fastjson.Parser{}

	for {
		select {
		case rawEvents := <-p.incoming:
			evnts, err := events.Parse(parser, rawEvents.Events)
			if err != nil {
				// здесь не до конца понятно насколько нам критичны подобного рода ошибки
				// как минимум такое можно залогировать в систему мониторинга ошибок (условная Sentry)
				p.logger.Printf("parsing error: %s\n", err)
				continue
			}

			if len(evnts) == 0 {
				continue
			}

			output <- dtos.EnrichmentEvents{
				Events: evnts,
				IP:     rawEvents.IP,
			}
		case <-p.ctx.Done():
			return
		}
	}
}

func (p *processor) save(events chan dtos.EnrichmentEvents) {
	for {
		select {
		case es := <-events:
			p.storage.Store(es)
		case <-p.ctx.Done():
			return
		}
	}
}
