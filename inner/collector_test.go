package inner

import (
	"context"
	"github.com/ppussar/mongodb_exporter/mocks"
	"github.com/stretchr/testify/mock"
	"gopkg.in/mgo.v2/bson"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestDescribeReturnsAllMetricDescriptionsOfTheCollector(t *testing.T) {

	metric := Metric{
		Name: "myMetric",
		Help: "myHelp",
	}
	con := Connection{}
	c := NewCollector(metric, con)
	ch := make(chan *prometheus.Desc, 1)
	c.Describe(ch)

	actual := <-ch

	assert.Equal(t, actual.String(), "Desc{fqName: \"myMetric\", help: \"myHelp\", constLabels: {}, variableLabels: []}", "Mismatching metrics description.")
}

func TestCollect(t *testing.T) {
	t.Run("Collect with aggregate: handle empty result", func(t *testing.T) {
		metric, _ := testMetric()
		metric.Aggregate = `[{"$group": { "_id": "$deliverer", "pieces": {"$sum": "$qty"}}}]`

		mongoCursor := mocks.ICursor{}
		mongoCursor.On("Next", mock.Anything).Return(false).Once()
		mongoCursor.On("Err").Return(nil).Once()
		mongoCursor.On("Close", mock.Anything).Return(nil).Once()
		mongoMock := mocks.IConnection{}
		mongoMock.On("Aggregate", metric.Db, metric.Collection, metric.Aggregate, context.Background()).Return(&mongoCursor, nil).Once()

		c := NewCollector(metric, &mongoMock)
		ch := make(chan prometheus.Metric, 1)
		c.Collect(ch)

		assert.Equal(t, 0, len(ch))

		mongoMock.AssertExpectations(t)
	})

	t.Run("Collect with aggregate: handle result", func(t *testing.T) {
		metric, doc := testMetric()
		metric.Aggregate = `[{"$group": { "_id": "$deliverer", "pieces": {"$sum": "$qty"}}}]`

		mongoCursor := mocks.ICursor{}
		mongoCursor.On("Next", mock.Anything).Return(true).Once()
		mongoCursor.On("Next", mock.Anything).Return(false).Once()
		mongoCursor.On("Err").Return(nil).Once()
		mongoCursor.On("Decode", mock.Anything).Return(nil).Run(func(args mock.Arguments) {
			arg := args.Get(0).(*bson.M)
			*arg = doc
		})

		mongoCursor.On("Close", mock.Anything).Return(nil).Once()
		mongoMock := mocks.IConnection{}
		mongoMock.On("Aggregate", metric.Db, metric.Collection, metric.Aggregate, context.Background()).Return(&mongoCursor, nil).Once()

		c := NewCollector(metric, &mongoMock)
		ch := make(chan prometheus.Metric, 1)
		c.Collect(ch)

		assert.Equal(t, 1, len(ch))
		actualMetric := <-ch
		assert.Equal(t, `Desc{fqName: "myMetric", help: "myHelp", constLabels: {constTag="value"}, variableLabels: [dynTag]}`, actualMetric.Desc().String())
		mongoMock.AssertExpectations(t)
	})

	t.Run("Collect with find: handle empty result", func(t *testing.T) {
		metric, _ := testMetric()
		metric.Find = "{}"

		mongoCursor := mocks.ICursor{}
		mongoCursor.On("Next", mock.Anything).Return(false).Once()
		mongoCursor.On("Err").Return(nil).Once()
		mongoCursor.On("Close", mock.Anything).Return(nil).Once()
		mongoMock := mocks.IConnection{}
		mongoMock.On("Find", metric.Db, metric.Collection, metric.Find, context.Background()).Return(&mongoCursor, nil).Once()

		c := NewCollector(metric, &mongoMock)
		ch := make(chan prometheus.Metric, 1)
		c.Collect(ch)

		assert.Equal(t, 0, len(ch))

		mongoMock.AssertExpectations(t)
	})

}

func testMetric() (Metric, bson.M) {
	metric := Metric{
		Name:       "myMetric",
		Help:       "myHelp",
		Db:         "myDB",
		Collection: "myCollection",
		Tags: map[string]string{
			"constTag": "value",
		},
		MetricsAttribute: "value",
		TagAttributes: map[string]string{
			"dynTag": "_id",
		},
	}

	value := bson.M{
		"_id": "112233",
		"value": 42.0,
	}

	return metric, value



}
