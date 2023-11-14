package internal

import (
	"context"
	"fmt"
	"github.com/ppussar/mongodb_exporter/internal/logger"
	"github.com/ppussar/mongodb_exporter/internal/wrapper"
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/mgo.v2/bson"
	"time"
)

// Collector queries one prometheus metric from mongoDB
type Collector struct {
	desc             *prometheus.Desc
	config           Metric
	Mongo            wrapper.IConnection
	varTagValueNames []string
	ErrorC           chan error
}

var log = logger.GetInstance()

// NewCollector constructor
// initializes every descriptor and returns a pointer to the collector
func NewCollector(m Metric, con wrapper.IConnection, errorC chan error) *Collector {
	varTagNames := make([]string, 0, len(m.TagAttributes))
	varTagValues := make([]string, 0, len(m.TagAttributes))
	for key, value := range m.TagAttributes {
		varTagNames = append(varTagNames, key)
		varTagValues = append(varTagValues, value)
	}
	return &Collector{
		desc: prometheus.NewDesc(
			m.Name,
			m.Help,
			varTagNames,
			m.Tags,
		),
		config:           m,
		Mongo:            con,
		varTagValueNames: varTagValues,
		ErrorC:           errorC,
	}
}

// Describe must be implemented by a prometheus collector
// It essentially writes all descriptors to the prometheus desc channel.
func (col *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- col.desc
}

// Collect implements required collect function for all prometheus collectors
func (col *Collector) Collect(ch chan<- prometheus.Metric) {

	var err error
	var cur wrapper.ICursor
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if len(col.config.Aggregate) != 0 {
		cur, err = col.Mongo.Aggregate(ctx, col.config.Db, col.config.Collection, col.config.Aggregate)
	} else if len(col.config.Find) != 0 {
		cur, err = col.Mongo.Find(ctx, col.config.Db, col.config.Collection, col.config.Find)
	} else {
		log.Error(fmt.Sprintf("Nothing to do, check config of metric: %v", col))
	}
	if err != nil {
		log.Error(fmt.Sprintf("Error during collect: %v", err))
		col.ErrorC <- err
		return
	}
	defer cur.Close(context.Background())
	for cur.Next(ctx) {
		var result bson.M
		err := cur.Decode(&result)
		if err != nil {
			log.Error(fmt.Sprintf("Error during collect: %v", err))
			col.ErrorC <- err
			return
		}

		val := result[col.config.MetricsAttribute]
		tagValues := col.extractVarTagsValues(result)
		ch <- prometheus.MustNewConstMetric(col.desc, prometheus.GaugeValue, val.(float64), tagValues...)
	}
	if err := cur.Err(); err != nil {
		log.Error(fmt.Sprintf("Error during collect: %v", err))
		col.ErrorC <- err
	}
}

func (col *Collector) extractVarTagsValues(result bson.M) []string {
	tagValues := make([]string, len(col.varTagValueNames))
	for i, tagName := range col.varTagValueNames {
		tagValue := result[tagName]
		tagValues[i] = tagValue.(string)
	}
	return tagValues
}

func (col *Collector) String() string {
	return col.desc.String()
}
