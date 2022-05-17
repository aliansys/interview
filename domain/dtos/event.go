package dtos

import (
	"github.com/google/uuid"
	"net"
	"time"
)

type (
	Event struct {
		DeviceId   uuid.UUID
		DeviceOs   string
		Session    string
		Event      string
		ParamStr   string
		Sequence   int64
		ParamInt   int64
		ClientTime time.Time
	}

	EventsWithIP struct {
		Events []Event
		IP     net.IP
	}

	RawEventWithIP struct {
		Events string
		IP     net.IP
	}
)
