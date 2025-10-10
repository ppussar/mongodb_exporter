package internal

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ppussar/mongodb_exporter/internal/logger"
	"github.com/ppussar/mongodb_exporter/internal/wrapper"
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/mgo.v2/bson"
)

// Collector queries one prometheus metric from mongoDB
type Collector struct {
	desc             *prometheus.Desc
	config           Metric
	mongo            wrapper.IConnection
	varTagValueNames []string
	errorC           chan error
	mu               sync.RWMutex
}

var log = logger.GetInstance()
var collectErrorMsg = "Error during collect: %v"

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
		mongo:            con,
		varTagValueNames: varTagValues,
		errorC:           errorC,
	}
}

// UpdateConnection safely updates the MongoDB connection
func (col *Collector) UpdateConnection(con wrapper.IConnection) {
	col.mu.Lock()
	defer col.mu.Unlock()
	col.mongo = con
}

// Describe must be implemented by a prometheus collector
// It essentially writes all descriptors to the prometheus desc channel.
func (col *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- col.desc
}

// Collect implements required collect function for all prometheus collectors
func (col *Collector) Collect(ch chan<- prometheus.Metric) {
	col.mu.RLock()
	mongo := col.mongo
	col.mu.RUnlock()
	
	if mongo == nil {
		col.sendError(fmt.Errorf("no MongoDB connection available"))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var cur wrapper.ICursor
	var err error
	
	if len(col.config.Aggregate) != 0 {
		cur, err = mongo.Aggregate(ctx, col.config.Db, col.config.Collection, col.config.Aggregate)
	} else if len(col.config.Find) != 0 {
		cur, err = mongo.Find(ctx, col.config.Db, col.config.Collection, col.config.Find)
	} else {
		col.sendError(fmt.Errorf("no query configured for metric: %s", col.config.Name))
		return
	}
	
	if err != nil {
		col.sendError(fmt.Errorf("query failed: %w", err))
		return
	}
	
	defer func() {
		if cur != nil {
			if err := cur.Close(ctx); err != nil {
				col.sendError(fmt.Errorf("cursor close failed: %w", err))
			}
		}
	}()

	for cur.Next(ctx) {
		var result bson.M
		if err := cur.Decode(&result); err != nil {
			col.sendError(fmt.Errorf("decode failed: %w", err))
			return
		}

		floatVal, err := col.extractMetricValue(result)
		if err != nil {
			col.sendError(err)
			return
		}

		tagValues, err := col.extractVarTagsValues(result)
		if err != nil {
			col.sendError(err)
			return
		}

		ch <- prometheus.MustNewConstMetric(col.desc, prometheus.GaugeValue, floatVal, tagValues...)
	}
	
	if err := cur.Err(); err != nil {
		col.sendError(fmt.Errorf("cursor iteration failed: %w", err))
	}
}

func (col *Collector) extractMetricValue(result bson.M) (float64, error) {
	val, exists := result[col.config.MetricsAttribute]
	if !exists {
		return 0, fmt.Errorf("metric attribute '%s' not found in result", col.config.MetricsAttribute)
	}

	switch v := val.(type) {
	case float64:
		return v, nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case int:
		return float64(v), nil
	default:
		return 0, fmt.Errorf("unsupported metric value type %T for %s", val, col.config.MetricsAttribute)
	}
}

func (col *Collector) extractVarTagsValues(result bson.M) ([]string, error) {
	tagValues := make([]string, len(col.varTagValueNames))
	for i, tagName := range col.varTagValueNames {
		tagValue, exists := result[tagName]
		if !exists {
			return nil, fmt.Errorf("tag attribute '%s' not found in result", tagName)
		}
		
		switch v := tagValue.(type) {
		case string:
			tagValues[i] = v
		case int, int32, int64, float32, float64:
			tagValues[i] = fmt.Sprintf("%v", v)
		default:
			return nil, fmt.Errorf("unsupported tag value type %T for %s", tagValue, tagName)
		}
	}
	return tagValues, nil
}

func (col *Collector) sendError(err error) {
	log.Error(fmt.Sprintf(collectErrorMsg, err))
	select {
	case col.errorC <- err:
	default:
		// Channel full, drop error to prevent blocking
	}
}

func (col *Collector) String() string {
	return col.desc.String()
}
