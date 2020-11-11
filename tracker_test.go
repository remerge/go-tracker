package tracker

import (
	"testing"
	"time"

	timestr "github.com/remerge/go-timestr"
	"github.com/stretchr/testify/assert"
)

type GE struct {
	given    interface{}
	expected []byte
}

type testEvent struct {
	*EventMetadata
}
type testEventSimple struct {
	*EventMetadata
}

func (e *testEvent) SetMetadata(md *EventMetadata) {
	e.EventMetadata = md
}

func (e *testEventSimple) SetMetadata(service, environment, cluster, host, release string) {
	e.EventMetadata = &EventMetadata{service, environment, cluster, host, release}
}

type customJSONEvent struct{}

func (*customJSONEvent) MarshalJSON() ([]byte, error) {
	return []byte("test"), nil
}

type timestampableEvent struct {
	Now time.Time
}

func (t *timestampableEvent) SetTimestampFrom(now time.Time, _ string) {
	t.Now = now
}

func TestBaseTrackerEncoding(t *testing.T) {
	timestr.Stop()
	ts := time.Date(1982, 4, 3, 12, 0, 0, 0, time.UTC)
	timestr.SetTimeStr(ts)

	tracker := &BaseTracker{Metadata: &EventMetadata{
		Service:     "Service",
		Environment: "Environment",
		Cluster:     "Cluster",
		Host:        "Host",
		Release:     "Release",
	}}

	tests := []GE{
		{
			[]byte{1, 2, 3, 4},
			[]byte{1, 2, 3, 4},
		},
		{
			"1234",
			[]byte("1234"),
		},
		{
			map[string]interface{}{"hallo": "test"},
			[]byte(`{"hallo":"test","ts":"1982-04-03T12:00:00Z"}`),
		},
		{
			&testEvent{},
			[]byte(`{"Service":"Service","Environment":"Environment","Cluster":"Cluster","Host":"Host","Release":"Release"}`),
		},
		{
			&testEventSimple{},
			[]byte(`{"Service":"Service","Environment":"Environment","Cluster":"Cluster","Host":"Host","Release":"Release"}`),
		},
		{
			&customJSONEvent{},
			[]byte("test"),
		},
		{
			&timestampableEvent{},
			[]byte(`{"Now":"1982-04-03T12:00:00Z"}`),
		},
		{
			nil,
			nil,
		},
	}

	for _, e := range tests {
		actual, err := tracker.Encode(e.given)
		assert.NoError(t, err)
		assert.Equal(t, e.expected, actual)
	}

}
