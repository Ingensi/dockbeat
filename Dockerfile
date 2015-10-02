FROM golang:1.5.1
MAINTAINER Ingensi labs <contact@ingensi.com>

COPY . /usr/src/dockerbeat
RUN cd /usr/src/dockerbeat && make

RUN mkdir -p /etc/dockerbeat/ \
    && cp /usr/src/dockerbeat/dockerbeat /etc/dockerbeat/ \
    && cp /usr/src/dockerbeat/etc/dockerbeat-docker.yml /etc/dockerbeat/dockerbeat.yml \
    && rm -rf /usr/src/dockerbeat

WORKDIR /etc/dockerbeat

CMD [ "./dockerbeat", "-c", "dockerbeat.yml" ]
