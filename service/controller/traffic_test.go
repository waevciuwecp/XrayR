package controller

import (
	"errors"
	"github.com/XrayR-project/XrayR/api"
	appstats "github.com/xtls/xray-core/app/stats"
	"testing"
)

type trafficAPI struct {
	api.API
	report func(*[]api.UserTraffic) error
}

func (a trafficAPI) ReportUserTraffic(traffic *[]api.UserTraffic) error { return a.report(traffic) }

func TestTrafficAcknowledgement(t *testing.T) {
	for _, fail := range []bool{false, true} {
		t.Run(map[bool]string{false: "success", true: "failure"}[fail], func(t *testing.T) {
			counter := new(appstats.Counter)
			counter.Add(100)
			c := &Controller{config: &Config{}}
			c.apiClient = trafficAPI{report: func(traffic *[]api.UserTraffic) error {
				if (*traffic)[0].Upload != 100 {
					t.Fatal("wrong snapshot")
				}
				done := make(chan struct{})
				go func() { counter.Add(37); close(done) }()
				<-done
				if fail {
					return errors.New("panel unavailable")
				}
				return nil
			}}
			err := c.reportTraffic([]api.UserTraffic{{Upload: 100}}, []trafficSnapshot{{counter: counter, value: 100}}, nil)
			want := int64(37)
			if fail {
				want = 137
			}
			if (err != nil) != fail || counter.Value() != want {
				t.Fatalf("remaining=%d, want=%d, err=%v", counter.Value(), want, err)
			}
		})
	}
}
