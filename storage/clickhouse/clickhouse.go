package clickhouse

import (
	"context"
	"fmt"
	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/aliansys/interview/domain/dtos"
	"github.com/google/uuid"
	"log"
	"net"
	"sync"
	"time"
)

type (
	Config struct {
		DSN       string
		BatchSize int
	}

	repo struct {
		config Config

		conn  driver.Conn
		batch []event

		m *sync.RWMutex

		ctx    context.Context
		cancel context.CancelFunc

		logger *log.Logger
	}

	event struct {
		DeviceId   uuid.UUID `ch:"device_id"`
		DeviceOs   string    `ch:"device_os"`
		Session    string    `ch:"session"`
		Event      string    `ch:"event"`
		ParamStr   string    `ch:"param_str"`
		Sequence   int64     `ch:"sequence"`
		ParamInt   int64     `ch:"param_int"`
		ClientTime time.Time `ch:"client_time"`
		ServerTime time.Time `ch:"server_time"`
		IP         net.IP    `ch:"ip"`
	}
)

const (
	tableName = "events"
)

var prepareInsertStatement = fmt.Sprintf("INSERT INTO %s", tableName)

func New(cfg Config, l *log.Logger) (*repo, error) {
	opts, err := clickhouse.ParseDSN(cfg.DSN)
	if err != nil {
		return nil, err
	}

	conn, err := clickhouse.Open(opts)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	// проверим, что соединение установлено
	err = conn.Ping(ctx)
	if err != nil {
		cancel()
		return nil, err
	}

	repo := &repo{
		config: cfg,
		conn:   conn,
		batch:  make([]event, 0, cfg.BatchSize),
		m:      new(sync.RWMutex),
		ctx:    ctx,
		cancel: cancel,
		logger: l,
	}

	go repo.run()

	return repo, nil
}

func (r *repo) run() {
	for {
		r.m.RLock()
		l := len(r.batch)
		r.m.RUnlock()
		if l >= r.config.BatchSize {
			r.m.Lock()

			err := r.save(r.ctx)
			if err != nil {
				r.m.Unlock()
				// здесь не до конца понятно насколько нам критичны подобного рода ошибки
				// как минимум такое можно залогировать в систему мониторинга ошибок (условная Sentry)
				r.logger.Printf("save went wrong: %s\n", err)
				return
			}

			r.batch = make([]event, 0, r.config.BatchSize)
			r.m.Unlock()
		}

		select {
		case <-r.ctx.Done():
			return
		default:
		}

		time.Sleep(time.Second)
	}
}

func (r *repo) Close() {
	if r.cancel != nil {
		r.cancel()
	}

	// сохраним оставшиеся события перед остановкой
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r.m.Lock()
	defer r.m.Unlock()

	err := r.save(ctx)
	if err != nil {
		r.logger.Printf("save went wrong: %s\n", err)
	}

	<-ctx.Done()

	r.conn.Close()
}

func (r *repo) Store(events dtos.EnrichmentEvents) {
	serverTime := time.Now()

	r.m.Lock()
	for i := range events.Events {
		e := events.Events[i]
		r.batch = append(r.batch, event{
			DeviceId:   e.DeviceId,
			DeviceOs:   e.DeviceOs,
			Session:    e.Session,
			Event:      e.Event,
			ParamStr:   e.ParamStr,
			Sequence:   e.Sequence,
			ParamInt:   e.ParamInt,
			ClientTime: e.ClientTime,
			ServerTime: serverTime,
			IP:         events.IP,
		})
	}
	r.m.Unlock()
}

func (r *repo) save(ctx context.Context) error {
	if len(r.batch) == 0 {
		return nil
	}
	batch, err := r.conn.PrepareBatch(ctx, prepareInsertStatement)
	if err != nil {
		return err
	}

	for i := range r.batch {
		err := batch.AppendStruct(&r.batch[i])

		if err != nil {
			return err
		}
	}

	return batch.Send()
}
