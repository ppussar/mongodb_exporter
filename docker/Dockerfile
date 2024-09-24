FROM alpine
ADD mongodb_exporter /
ENTRYPOINT ["/bin/sh", "-c", "/mongodb_exporter $CONFIG"]
