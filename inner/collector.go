package inner

import (
	"log"

	"github.com/prometheus/client_golang/prometheus"
	"go.mongodb.org/mongo-driver/mongo"
	"gopkg.in/mgo.v2/bson"
)

// Collector queries one prometheus metric from mongoDB
type Collector struct {
	desc             *prometheus.Desc
	config           Metric
	mongo            Connection
	varTagValueNames []string
}

//NewCollector constructor
//initializes every descriptor and returns a pointer to the collector
func NewCollector(c Metric, con Connection) *Collector {
	varTagNames := make([]string, 0, len(c.TagAttributes))
	varTagValues := make([]string, 0, len(c.TagAttributes))
	for key, value := range c.TagAttributes {
		varTagNames = append(varTagNames, key)
		varTagValues = append(varTagValues, value)
	}
	return &Collector{
		desc: prometheus.NewDesc(
			c.Name,
			c.Help,
			varTagNames,
			c.Tags,
		),
		config:           c,
		mongo:            con,
		varTagValueNames: varTagValues,
	}
}

//Describe must be implemented by a prometheus collector
//It essentially writes all descriptors to the prometheus desc channel.
func (col *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- col.desc
}

//Collect implements required collect function for all prometheus collectors
func (col *Collector) Collect(ch chan<- prometheus.Metric) {

	var err error
	var cur *mongo.Cursor
	if len(col.config.Aggregate) != 0 {
		cur, err = col.mongo.aggregate(col.config.Db, col.config.Collection, col.config.Aggregate)
	} else {
		cur, err = col.mongo.find(col.config.Db, col.config.Collection, col.config.Find)
	}
	if err != nil {
		log.Fatal(err)
	}
	defer cur.Close(col.mongo.Context)
	for cur.Next(col.mongo.Context) {
		var result bson.M
		err := cur.Decode(&result)
		if err != nil {
			log.Fatal(err)
		}

		val := result[col.config.MetricsAttribute]
		tagValues := col.extractVarTagsValues(result)
		ch <- prometheus.MustNewConstMetric(col.desc, prometheus.GaugeValue, val.(float64), tagValues...)
	}
	if err := cur.Err(); err != nil {
		log.Fatal(err)
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
