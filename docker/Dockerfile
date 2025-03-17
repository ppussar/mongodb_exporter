FROM alpine
COPY mongodb_exporter /
ENTRYPOINT ["/bin/sh", "-c", "/mongodb_exporter $CONFIG"]
