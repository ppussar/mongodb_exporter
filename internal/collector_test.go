package internal

import (
	"testing"
	"time"

	"github.com/ppussar/mongodb_exporter/internal/mocks"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestCollector(t *testing.T) {
	t.Run("NewCollector", func(t *testing.T) {

		metric := Metric{
			Name: "name",
			Help: "help",
			Tags: map[string]string{
				"tagKey": "tagValue",
			},
			TagAttributes: map[string]string{
				"tagAttrKey": "tagAttrValue",
			},
		}
		con := Connection{}
		c := NewCollector(metric, con, make(chan error, 1))

		assert.Equal(t, "Desc{fqName: \"name\", help: \"help\", unit: \"\", constLabels: {tagKey=\"tagValue\"}, variableLabels: {tagAttrKey}}", c.String())
	})
}

func TestDescribe(t *testing.T) {
	t.Run("Describe returns all MetricDescriptions of the collector", func(t *testing.T) {
		metric := Metric{
			Name: "myMetric",
			Help: "myHelp",
		}
		con := Connection{}
		c := NewCollector(metric, con, make(chan error, 1))
		ch := make(chan *prometheus.Desc, 1)
		c.Describe(ch)

		actual := <-ch

		assert.Equal(t, actual.String(), "Desc{fqName: \"myMetric\", help: \"myHelp\", unit: \"\", constLabels: {}, variableLabels: {}}", "Mismatching metrics description.")
	})

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
		mongoMock.On("Aggregate", mock.Anything, metric.Db, metric.Collection, metric.Aggregate).Return(&mongoCursor, nil).Once()

		c := NewCollector(metric, &mongoMock, make(chan error, 1))
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
		mongoMock.On("Aggregate", mock.Anything, metric.Db, metric.Collection, metric.Aggregate).Return(&mongoCursor, nil).Once()

		c := NewCollector(metric, &mongoMock, make(chan error, 1))
		ch := make(chan prometheus.Metric, 1)
		c.Collect(ch)

		assert.Equal(t, 1, len(ch))
		actualMetric := <-ch
		assert.Equal(t, `Desc{fqName: "myMetric", help: "myHelp", unit: "", constLabels: {constTag="value"}, variableLabels: {dynTag}}`, actualMetric.Desc().String())
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
		mongoMock.On("Find", mock.Anything, metric.Db, metric.Collection, metric.Find).Return(&mongoCursor, nil).Once()

		c := NewCollector(metric, &mongoMock, make(chan error, 1))
		ch := make(chan prometheus.Metric, 1)
		c.Collect(ch)

		assert.Equal(t, 0, len(ch))

		mongoMock.AssertExpectations(t)
	})

}

// QueryDuration is global, so every subtest uses its own metric name.
func TestCollectQueryDuration(t *testing.T) {
	const queryLatency = 100 * time.Millisecond

	t.Run("Collect with aggregate: duration includes the query", func(t *testing.T) {
		metric, _ := testMetric()
		metric.Name = "slowAggregateMetric"
		metric.Aggregate = `[{"$group": { "_id": "$deliverer", "pieces": {"$sum": "$qty"}}}]`

		mongoCursor := emptyCursor()
		mongoMock := mocks.IConnection{}
		mongoMock.On("Aggregate", mock.Anything, metric.Db, metric.Collection, metric.Aggregate).
			Run(func(mock.Arguments) { time.Sleep(queryLatency) }).
			Return(mongoCursor, nil).Once()

		c := NewCollector(metric, &mongoMock, make(chan error, 1))
		c.Collect(make(chan prometheus.Metric, 1))

		assert.GreaterOrEqual(t, queryDurationSum(t, metric.Name, "aggregate"), queryLatency.Seconds())
		mongoMock.AssertExpectations(t)
	})

	t.Run("Collect with find: duration includes the query", func(t *testing.T) {
		metric, _ := testMetric()
		metric.Name = "slowFindMetric"
		metric.Find = "{}"

		mongoCursor := emptyCursor()
		mongoMock := mocks.IConnection{}
		mongoMock.On("Find", mock.Anything, metric.Db, metric.Collection, metric.Find).
			Run(func(mock.Arguments) { time.Sleep(queryLatency) }).
			Return(mongoCursor, nil).Once()

		c := NewCollector(metric, &mongoMock, make(chan error, 1))
		c.Collect(make(chan prometheus.Metric, 1))

		assert.GreaterOrEqual(t, queryDurationSum(t, metric.Name, "find"), queryLatency.Seconds())
		mongoMock.AssertExpectations(t)
	})
}

func emptyCursor() *mocks.ICursor {
	cursor := &mocks.ICursor{}
	cursor.On("Next", mock.Anything).Return(false).Once()
	cursor.On("Err").Return(nil).Once()
	cursor.On("Close", mock.Anything).Return(nil).Once()
	return cursor
}

// queryDurationSum returns the total query duration observed for the given metric and query type.
func queryDurationSum(t *testing.T, metricName string, queryType string) float64 {
	t.Helper()
	families, err := prometheus.DefaultGatherer.Gather()
	assert.NoError(t, err)
	for _, family := range families {
		if family.GetName() != "mongodb_exporter_query_duration_seconds" {
			continue
		}
		for _, m := range family.GetMetric() {
			labels := map[string]string{}
			for _, label := range m.GetLabel() {
				labels[label.GetName()] = label.GetValue()
			}
			if labels["metric_name"] == metricName && labels["query_type"] == queryType {
				return m.GetHistogram().GetSampleSum()
			}
		}
	}
	t.Fatalf("no %s query duration observed for %s", queryType, metricName)
	return 0
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
		"_id":   "112233",
		"value": 42.0,
	}

	return metric, value
}
