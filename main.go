package main

import (
	"flag"
	"fmt"
	"github.com/aliansys/interview/helpers/signals"
	configparser "github.com/aliansys/interview/parsers/config/config"
	"github.com/aliansys/interview/services/eventssaver"
	"github.com/aliansys/interview/storage/clickhouse"
	"github.com/aliansys/interview/storage/fake"
	"github.com/aliansys/interview/web/controllers/events"
	"net/http"
	"syscall"
)

func main() {
	cfg, err := configparser.Parse("./config/config.yml")
	if err != nil {
		panic(err)
	}

	noCHPtr := flag.Bool("no-ch", false, "run with no clickhouse server")
	flag.Parse()

	var storage eventssaver.Repo

	if !*noCHPtr {
		storage, err = clickhouse.New(clickhouse.Config(cfg.ClickHouse))
		if err != nil {
			panic(err)
		}
		defer storage.Close()
	} else {
		storage, _ = fake.New()
	}

	mux := http.NewServeMux()

	saver := eventssaver.New(storage)
	defer saver.Close()

	eventController := events.New(saver)
	eventController.Register(mux)

	fmt.Printf("Server is listening on %v\n", cfg.Api.Address)
	go func() {
		err = http.ListenAndServe(cfg.Api.Address, mux)
		if err != nil {
			panic(err)
		}
	}()

	<-signals.Signal(syscall.SIGINT, syscall.SIGTERM)
}
