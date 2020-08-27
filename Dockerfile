FROM alpine
ADD plenti /usr/bin/plenti
ENTRYPOINT tail -f /dev/null