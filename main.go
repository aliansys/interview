package main

import (
	"flag"
	"fmt"
	"github.com/aliansys/interview/helpers/signals"
	configparser "github.com/aliansys/interview/parsers/config/config"
	"github.com/aliansys/interview/services/eventsprocessor"
	"github.com/aliansys/interview/storage/clickhouse"
	"github.com/aliansys/interview/storage/fake"
	"github.com/aliansys/interview/web/controllers/events"
	"log"
	"net/http"
	"syscall"
)

func main() {
	cfg, err := configparser.Parse("./config/config.yml")
	if err != nil {
		panic(err)
	}

	l := log.Default()

	noCHPtr := flag.Bool("no-ch", false, "run with no clickhouse server")
	flag.Parse()

	var storage eventsprocessor.Storage

	if !*noCHPtr {
		storage, err = clickhouse.New(clickhouse.Config(cfg.ClickHouse), l)
		if err != nil {
			panic(err)
		}
		defer storage.Close()
	} else {
		storage, _ = fake.New(l)
	}

	mux := http.NewServeMux()

	processor := eventsprocessor.New(storage, l)
	defer processor.Close()

	controller := events.New(processor)
	controller.Register(mux)

	fmt.Printf("Server is listening on %v\n", cfg.Api.Address)
	go func() {
		err = http.ListenAndServe(cfg.Api.Address, mux)
		if err != nil {
			panic(err)
		}
	}()

	<-signals.Signal(syscall.SIGINT, syscall.SIGTERM)
}
