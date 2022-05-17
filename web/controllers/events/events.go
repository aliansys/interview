package events

import (
	"fmt"
	"github.com/aliansys/interview/domain/dtos"
	net2 "github.com/aliansys/interview/helpers/net"
	"net/http"
)

type (
	Processor interface {
		Process(events dtos.RawEventWithIP)
	}

	events struct {
		processor Processor
	}
)

func New(p Processor) *events {
	return &events{
		processor: p,
	}
}

func (e *events) Register(mux *http.ServeMux) {
	mux.HandleFunc("/v1/events", e.post)
}

func (e *events) post(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		return
	}

	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "form err: %v", err)
		return
	}

	events := r.FormValue("events")
	userIP, err := net2.IpFromAddressString(r.RemoteAddr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "something went wrong. err %s", err)
		return
	}

	e.processor.Process(dtos.RawEventWithIP{
		Events: events,
		IP:     userIP,
	})
}
