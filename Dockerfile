FROM alpine
COPY plenti /
ENTRYPOINT tail -f /dev/null