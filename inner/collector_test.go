package inner

import (
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

/*
func TestCollect(t *testing.T) {
	metric := Metric{
		Name: "myMetric",
		Help: "myHelp",
	}
	con := Connection{} // TODO Mock Me
	c := NewCollector(metric, &con)
	ch := make(chan prometheus.Metric, 1)
	c.Collect(ch)

	actual := <-ch

	assert.Equal(t, actual, "todo", "Mismatching metrics vector.")
}
*/
