FROM golang:1.6.1
MAINTAINER Ingensi labs <contact@ingensi.com>

# install pyyaml
RUN cd /tmp && wget http://pyyaml.org/download/pyyaml/PyYAML-3.11.tar.gz && tar -zxvf PyYAML-3.11.tar.gz
RUN cd /tmp/PyYAML-3.11 && python setup.py install
# install glide
RUN go get github.com/Masterminds/glide

COPY . $GOPATH/src/github.com/ingensi/dockerbeat
RUN cd $GOPATH/src/github.com/ingensi/dockerbeat && make && make

RUN mkdir -p /etc/dockerbeat/ \
    && cp $GOPATH/src/github.com/ingensi/dockerbeat/dockerbeat /usr/local/bin/dockerbeat \
    && cp $GOPATH/src/github.com/ingensi/dockerbeat/dockerbeat-docker.yml /etc/dockerbeat/dockerbeat.yml

WORKDIR /etc/dockerbeat
ENTRYPOINT dockerbeat

CMD [ "-c", "dockerbeat.yml", "-e" ]