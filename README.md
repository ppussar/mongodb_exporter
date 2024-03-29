# MongoDB Query Exporter

[![GoReport](https://goreportcard.com/badge/github.com/ppussar/mongodb_exporter)](https://goreportcard.com/report/github.com/ppussar/mongodb_exporter)

Prometheus exporter for MongoDB queries. Extract metrics from mongoDB queries results.

## Build and Run

### Bash

```bash
make build
./bin/mongodb_exporter configuration.yaml
```

### Docker

```bash
docker run -v /local/path/to/configuration.yaml:/configuration.yaml -e CONFIG=/configuration.yaml ppussar/mongodb_exporter
```

### Run Demo Application

(Requires docker)

Start mongodb and application container.
```bash
make start-demo
curl localhost:9090/prometheus
```

Stop mongodb and application container.
```bash
make stop-demo
```

## Configuration

The exporter is configured via yaml.

### HTTP Endpoints

#### Prometheus

Prometheus scrapes its monitored applications regularly via a provided http endpoint. 
The configuration below opens a http endpoint on [http://localhost:9090/prometheus](http://localhost:9090/prometheus)

#### Liveliness

Liveliness-Endpoint returns http status code 204 - [No Content] as soon as the exporter is started. 
An HTTP error status code indicates that the application is in bad condition and should be restarted.  
The configuration below opens a http endpoint on [http://localhost:9090/live](http://localhost:9090/live)

#### Health

Health-Endpoint returns http status code 200 as soon as the exporter is ready to serve user request.
An HTTP error status code indicates that the application is currently not able to collect db metrics.
The configuration below opens a http endpoint on [http://localhost:9090/health](http://localhost:9090/health)

```yaml class:"lineNo"
http:
  port: 9090
  path: /prometheus
  health: '/health'
  liveliness: '/live'
```

`https is currently not supported.

### Mongo-DB connection

MongoDB [connection-string](https://docs.mongodb.com/manual/reference/connection-string/)

```yaml class:"lineNo"
mongodb:
  uri: mongodb://localhost:27017
```

### Metric Queries

```yaml class:"lineNo"
metrics:
  - name: my.metric.name
    db: mydb
    collection: mycollection
    tags:
      myTag: "abc"
      bla: "blub"
    aggregation: "{$group: { _id: '$version', count: { $sum: 1 }}}"
    metricsAttribute: count
    tagAttributes:
      version: "_id"
  - name: another.metric.name
    ...
```

| key              | description                                                                    | example                                          | reference                                                                 |
|------------------|--------------------------------------------------------------------------------|--------------------------------------------------|---------------------------------------------------------------------------|
| name             | Metric name                                                                    | my.metric.name                                   | <https://micrometer.io/docs/concepts#_naming_meters>       |
| help             | Metric help value on prometheus scrape page                                    | Yet another metric                               |                                                                           |
| db               | MongoDB DB instance, which should be used.                                     | myDB                                             |                                                                           |
| collection       | MongoDB collection, which becomes queried.                                     | myCollection                                     |                                                                           |
| tags             | Map of static tags. Will be added to all resulting metrics.                    | tagKey: tagValue                                 |                                                                           |
| aggregation      | MongoDB aggregation query.                                                     | {$group: { _id: '$version', count: { $sum: 1 }}} | <https://docs.mongodb.com/manual/reference/method/db.collection.aggregate/> |
| find             | MongoDB find query.                                                            | {}                                               | <https://docs.mongodb.com/manual/reference/method/db.collection.find/>      |
| metricsAttribute | Attribute of the query result, which will be taken as gauge value.             | metricsAttribute: resultFieldName                |                                                                           |
| tagAttributes    | Map of attributes of the query result, which will be taken as additional tags. | tagKey: resultFieldName                          |                                                                           |

## Example Configuration

### Given Collection 'fruits'

```json
[
  { "_id": "apples", "qty": 5, "deliverer" : "Fruit Express"},
  { "_id": "bananas", "qty": 7, "deliverer" : "Bananas Daily" },
  { "_id": "oranges", "qty": 12, "deliverer" : "Fruit Express"},
  { "_id": "avocados", "qty": 14, "deliverer" : "Fruit Marked" }
]
```

### Configuration with adapted prometheus endpoint port

```yaml
version: 1.0
http:
  port: 9090
  path: /prometheus
mongodb:
  uri: mongodb://localhost:27017
metrics:
  - name: fruitstore_stock
    help: "Shows the current stock"
    db: fruitstore
    collection: fruits
    tags:
      provider: mongodb_exporter
    find: "{}"
    metricsAttribute: qty
    tagAttributes:
      type: _id
  - name: fruitstore_total
    db: fruitstore
    collection: fruits
    tags:
      provider: mongodb_exporter
    aggregate: '[{"$group": { "_id": "$deliverer", "pieces": {"$sum": "$qty"}}}]'
    metricsAttribute: pieces
    tagAttributes:
      type: _id
```

### scrape

```bash
curl http://localhost:9090/prometheus
```

### Output

```bash
# HELP Shows the current stock
# TYPE fruitstore_stock gauge
fruitstore_stock{provider="mongodb_exporter",type="avocados",} 14.0
fruitstore_stock{provider="mongodb_exporter",type="oranges",} 12.0
fruitstore_stock{provider="mongodb_exporter",type="apples",} 5.0
fruitstore_stock{provider="mongodb_exporter",type="bananas",} 7.0
# HELP fruitstore_total
# TYPE fruitstore_total gauge
fruitstore_total{provider="mongodb_exporter",type="Fruit Express",} 17.0
fruitstore_total{provider="mongodb_exporter",type="Bananas Daily",} 7.0
fruitstore_total{provider="mongodb_exporter",type="Fruit Marked",} 14.0
```
