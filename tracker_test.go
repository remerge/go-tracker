package tracker

import (
	"testing"
	"time"

	timestr "github.com/remerge/go-timestr"
	"github.com/stretchr/testify/assert"
)

type GE struct {
	given    interface{}
	expected string
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

	for _, e := range []GE{
		{[]byte{1, 2, 3, 4}, string([]byte{1, 2, 3, 4})},
		{"1234", "1234"},
		{map[string]interface{}{"hallo": "test"}, `{"hallo":"test","ts":"1982-04-03T12:00:00Z"}`},
		{&testEvent{}, `{"Service":"Service","Environment":"Environment","Cluster":"Cluster","Host":"Host","Release":"Release"}`},
		{&testEventSimple{}, `{"Service":"Service","Environment":"Environment","Cluster":"Cluster","Host":"Host","Release":"Release"}`},
		{&customJSONEvent{}, "test"},
		{&timestampableEvent{}, `{"Now":"1982-04-03T12:00:00Z"}`},
		{nil, ""},
	} {
		actual, err := tracker.Encode(e.given)
		assert.NoError(t, err)
		assert.Equal(t, []byte(e.expected), actual)
	}

}
