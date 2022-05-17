package events

import (
	"github.com/aliansys/interview/domain/dtos"
	"github.com/google/uuid"
	"github.com/valyala/fastjson"
	"strings"
	"time"
)

const (
	dateLayout      = "2006-01-02 15:04:05"
	avgEventsNumber = 30
)

func Parse(p *fastjson.Parser, rawEvents string) ([]dtos.Event, error) {
	parsedEvents := make([]dtos.Event, 0, avgEventsNumber)
	start := 0
	for i := range rawEvents {
		metBackSlashN := rawEvents[i] == '\n'
		isPrevCharBracket := i > 0 && rawEvents[i-1] == '}'
		reachedEnd := i == (len(rawEvents) - 1)

		if !(metBackSlashN && isPrevCharBracket) && !reachedEnd {
			continue
		}

		end := i

		if start >= end {
			break
		}

		if reachedEnd {
			end += 1
		}

		ev, err := parseOneEvent(p, rawEvents[start:end])
		if err != nil {
			return nil, err
		}

		parsedEvents = append(parsedEvents, ev)
		start += i + 1
	}

	return parsedEvents, nil
}

func parseOneEvent(p *fastjson.Parser, rawEvent string) (dtos.Event, error) {
	parsed, err := p.Parse(rawEvent)
	if err != nil {
		return dtos.Event{}, err
	}

	rawTime := parsed.GetStringBytes("client_time")
	t, err := ParseCustomDate(string(rawTime))
	if err != nil {
		return dtos.Event{}, err
	}

	uuid, err := uuid.Parse(string(parsed.GetStringBytes("device_id")))
	if err != nil {
		return dtos.Event{}, err
	}
	return dtos.Event{
		DeviceId:   uuid,
		DeviceOs:   string(parsed.GetStringBytes("device_os")),
		Session:    string(parsed.GetStringBytes("session")),
		Event:      string(parsed.GetStringBytes("event")),
		ParamStr:   string(parsed.GetStringBytes("param_str")),
		Sequence:   parsed.GetInt64("sequence"),
		ParamInt:   parsed.GetInt64("param_int"),
		ClientTime: t,
	}, nil
}

func ParseCustomDate(date string) (time.Time, error) {
	s := strings.Trim(date, "\"")

	return time.Parse(dateLayout, s)
}
