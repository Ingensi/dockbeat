FROM golang:1.5.1
MAINTAINER Ingensi labs <contact@ingensi.com>

COPY . /go/src/github.com/ingensi/dockerbeat
RUN cd /go/src/github.com/ingensi/dockerbeat && make

RUN mkdir -p /etc/dockerbeat/ \
    && cp /usr/src/dockerbeat/dockerbeat /etc/dockerbeat/ \
    && cp /usr/src/dockerbeat/etc/dockerbeat-docker.yml /etc/dockerbeat/dockerbeat.yml \
    && rm -rf /usr/src/dockerbeat

WORKDIR /etc/dockerbeat

CMD [ "./dockerbeat", "-c", "dockerbeat.yml" ]
