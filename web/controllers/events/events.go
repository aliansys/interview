package events

import (
	"fmt"
	"github.com/aliansys/interview/domain/dtos"
	nethelpers "github.com/aliansys/interview/helpers/net"
	"net/http"
)

type (
	Processor interface {
		Process(events dtos.RawEnrichmentEvents)
	}

	controller struct {
		processor Processor
	}
)

func New(p Processor) *controller {
	return &controller{
		processor: p,
	}
}

func (c *controller) Register(mux *http.ServeMux) {
	mux.HandleFunc("/v1/events", c.post)
}

func (c *controller) post(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		return
	}

	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "form err: %v", err)
		return
	}

	events := r.FormValue("events")
	userIP, err := nethelpers.IpFromAddressString(r.RemoteAddr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "something went wrong. err %s", err)
		return
	}

	c.processor.Process(dtos.RawEnrichmentEvents{
		Events: events,
		IP:     userIP,
	})
}
