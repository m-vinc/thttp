FROM debian:11-slim

RUN apt-get update && apt-get install -y tor

RUN mkdir /etc/tor/run && \
    chmod -R 600 /etc/tor && \
    chown -R root:root /var/lib/tor

COPY torrc /etc/tor/torrc

ENTRYPOINT [ "tor" ]