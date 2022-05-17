package eventsprocessor

import (
	"fmt"
	"github.com/aliansys/interview/domain/dtos"
	"log"
	"net"
	"sync"
	"testing"
	"time"
)

var benchmarkSizes = []int{1, 30, 100, 300}

func BenchmarkProcess(b *testing.B) {
	events := make(map[int]dtos.RawEnrichmentEvents)

	sample := `{"client_time":"2020-12-01 23:59:00","device_id":"0287D9AA-4ADF-4B37-A60F-3E9E645C821E","device_os":"iOS 13.5.1","session":"ybuRi8mAUypxjbxQ","sequence":1,"event":"app_start","param_int":0,"param_str":"some text"}`

	for _, size := range benchmarkSizes {
		str := ""
		for i := 0; i < size; i++ {
			str += sample + "\n"
		}
		events[size] = dtos.RawEnrichmentEvents{
			Events: str,
			IP:     net.IPv4('8', '8', '8', '8'),
		}
	}

	b.ResetTimer()
	for _, size := range benchmarkSizes {
		b.Run(fmt.Sprintf("%d", size), func(b *testing.B) {
			t0 := time.Now()
			storage := &mockStorage{
				wg: sync.WaitGroup{},
			}
			s := New(storage, log.Default())
			s.run()

			for i := 0; i < b.N; i++ {
				storage.wg.Add(1)
				s.Process(events[size])
			}
			storage.wg.Wait()
			s.Close()
			b.ReportMetric(float64(time.Since(t0))/float64(b.N), "ns/Process()")
		})
	}
}

type (
	mockStorage struct {
		wg sync.WaitGroup
	}
)

func (m *mockStorage) Store(_ dtos.EnrichmentEvents) {
	m.wg.Done()
}

func (m *mockStorage) Close() {}
