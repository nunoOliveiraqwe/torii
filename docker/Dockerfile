FROM debian:stable-slim

COPY torii /torii

WORKDIR /data

ENTRYPOINT ["/torii"]
CMD ["-config", "/etc/torii/config.yaml"]

