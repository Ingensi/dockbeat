FROM debian:jessie 
MAINTAINER Ingensi labs <contact@ingensi.com>

RUN mkdir -p /opt/dockerbeat
ADD dockerbeat /opt/dockerbeat/dockerbeat
ADD etc/dockerbeat.yml /opt/dockerbeat/dockerbeat.yml


CMD [ "/opt/dockerbeat/dockerbeat", "-c", "/opt/dockerbeat/dockerbeat.yml" ]
