package = github.com/ppussar/mongodb_exporter

run:
	go run *.go configuration.yaml

build:
	go build -v .

startenv:
	docker-compose -f docker/docker-compose.yaml up -d

stopenv:
	docker-compose -f docker/docker-compose.yaml down --remove-orphans
