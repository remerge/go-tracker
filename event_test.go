package tracker

import (
	"fmt"
	"testing"

	"github.com/remerge/go-timestr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type EventTestSuite struct {
	suite.Suite
}

func TestEvent(t *testing.T) {
	suite.Run(t, new(EventTestSuite))
}

func (suite *EventTestSuite) SetupTest() {
	timestr.UpdateTimeStr()
}

func (suite *EventTestSuite) TestEventMetadata() {
	m := &EventMetadata{
		Service:     "test",
		Environment: "testing",
		Cluster:     "t1",
		Host:        "testhost",
		Release:     "123abc",
	}

	var e Event = &EventBase{}
	e.SetMetadata(m)

	eb := e.(*EventBase)
	assert.Equal(suite.T(), eb.Service, "test")
	assert.Equal(suite.T(), eb.Environment, "testing")
	assert.Equal(suite.T(), eb.Cluster, "t1")
	assert.Equal(suite.T(), eb.Host, "testhost")
	assert.Equal(suite.T(), eb.Release, "123abc")
}

type testEvent struct {
	EventBase
	MyField string
}

func (suite *EventTestSuite) TestEventSerialize() {
	m := &EventMetadata{
		Service:     "test",
		Environment: "testing",
		Cluster:     "t1",
		Host:        "testhost",
		Release:     "123abc",
	}

	event := &testEvent{
		MyField: "foo",
	}
	event.SetMetadata(m)

	bytes, err := event.MarshalJSON()
	assert.Nil(suite.T(), err)
	fmt.Println(string(bytes))
}
