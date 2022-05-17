package events

import (
	"fmt"
	"github.com/aliansys/interview/domain/dtos"
	"github.com/google/uuid"
	"github.com/valyala/fastjson"
	"reflect"
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	type args struct {
		events string
	}

	clientTime := "2020-12-01 23:59:00"
	deviceId, _ := uuid.Parse("0287D9AA-4ADF-4B37-A60F-3E9E645C821E")
	evnt := `{"param_str":"%s", "client_time":"%s", "device_id":"0287D9AA-4ADF-4B37-A60F-3E9E645C821E"}`
	tm, _ := ParseCustomDate(clientTime)
	tests := []struct {
		name    string
		args    args
		want    []dtos.Event
		wantErr bool
	}{
		{
			name: "1 event",
			args: args{
				events: fmt.Sprintf(evnt, "hey", clientTime),
			},
			want: []dtos.Event{
				{
					ParamStr:   "hey",
					DeviceId:   deviceId,
					ClientTime: tm,
				},
			},
			wantErr: false,
		},
		{
			name: "1 event with \\n in paramStr",
			args: args{
				events: fmt.Sprintf(evnt, "hey\n", clientTime),
			},
			want: []dtos.Event{
				{
					ParamStr:   "hey\n",
					DeviceId:   deviceId,
					ClientTime: tm,
				},
			},
			wantErr: false,
		},
		{
			name: "1 event with \\n in the beginning of an input",
			args: args{
				events: "\n" + fmt.Sprintf(evnt, "hey", clientTime),
			},
			want: []dtos.Event{
				{
					ParamStr:   "hey",
					DeviceId:   deviceId,
					ClientTime: tm,
				},
			},
			wantErr: false,
		},
		{
			name: "1 event with \\n in the end of an input",
			args: args{
				events: fmt.Sprintf(evnt, "hey", clientTime) + "\n",
			},
			want: []dtos.Event{
				{
					ParamStr:   "hey",
					DeviceId:   deviceId,
					ClientTime: tm,
				},
			},
			wantErr: false,
		},
		{
			name: "empty string",
			args: args{
				events: ``,
			},
			want:    make([]dtos.Event, 0),
			wantErr: false,
		},
		{
			name: "broken json",
			args: args{
				events: `"param_int":1}`,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "broken date",
			args: args{
				events: fmt.Sprintf(evnt, "hey", ""),
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &fastjson.Parser{}
			got, err := Parse(p, tt.args.events)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parse() got = %+v, want %+v", got, tt.want)
			}
		})
	}
}

var benchmarkSizes = []int{1, 30, 100, 300}

func BenchmarkParse(b *testing.B) {
	events := make(map[int]string)

	sample := `{"client_time":"2020-12-01 23:59:00","device_id":"0287D9AA-4ADF-4B37-A60F-3E9E645C821E","device_os":"iOS 13.5.1","session":"ybuRi8mAUypxjbxQ","sequence":1,"event":"app_start","param_int":0,"param_str":"some text"}`

	for _, size := range benchmarkSizes {
		str := ""
		for i := 0; i < size; i++ {
			str += sample + "\n"
		}
		events[size] = str
	}

	b.ResetTimer()
	for _, size := range benchmarkSizes {
		b.Run(fmt.Sprintf("%d", size), func(b *testing.B) {
			t0 := time.Now()
			p := &fastjson.Parser{}
			for i := 0; i < b.N; i++ {
				_, err := Parse(p, events[size])
				if err != nil {
					b.Fatalf("err %s\n", err)
				}
			}
			b.ReportMetric(float64(time.Since(t0))/float64(b.N), "ns/Parse()")
		})
	}
}
