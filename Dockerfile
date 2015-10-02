FROM golang:1.5.1
MAINTAINER Ingensi labs <contact@ingensi.com>

ENV http_proxy=http://172.31.1.71:8081/
ENV https_proxy=http://172.31.1.71:8081/

COPY . /usr/src/dockerbeat
RUN cd /usr/src/dockerbeat && make

RUN mkdir -p /etc/dockerbeat/ \
    && cp /usr/src/dockerbeat/dockerbeat /etc/dockerbeat/ \
    && cp /usr/src/dockerbeat/etc/dockerbeat-docker.yml /etc/dockerbeat/dockerbeat.yml \
    && rm -rf /usr/src/dockerbeat

WORKDIR /etc/dockerbeat

CMD [ "./dockerbeat", "-c", "dockerbeat.yml" ]
