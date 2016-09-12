# Dockbeat

(if you're on the fast lane, check the TL;DR at the bottom of the readme)

Build status : [![Build Status](https://travis-ci.org/Ingensi/dockbeat.svg?branch=develop)](https://travis-ci.org/Ingensi/dockbeat)

Test coverage : [![codecov.io](http://codecov.io/github/Ingensi/dockbeat/coverage.svg?branch=develop)](http://codecov.io/github/Ingensi/dockbeat?branch=develop)

Dockbeat is the new Dockerbeat name. We had to rename the project due to the [Docker trademarking policy](https://www.docker.com/trademark-guidelines). Dockbeat is a [Beat](https://www.elastic.co/products/beats) used for docker daemon monitoring. It is a lightweight agent that installed on your servers, reads periodically docker container statistics and indexes them in Elasticsearch.

## Exported document types

There are five types of documents exported:

- `type: container`: container attributes
- `type: cpu`: container CPU usage statistics. One document per container is generated.
- `type: net`: container network statistics. One document per network container is generated.
- `type: memory`: container memory statistics. One document per container is generated.
- `type: blkio`: container io access statistics. One document per container is generated.
- `type: log`: dockbeat status information. One document per tick is generated if an error occurred.

To get a detailed list of all generated fields, please read the [fields documentation page](docs/fields.asciidoc).

## Elasticsearch template 

To apply Dockbeat template (recommended but not required) :

```bash
curl -XPUT 'http://elastic:9200/_template/dockbeat' -d@etc/dockbeat.template.json
```
    
## Build Dockbeat

Ensure that this folder is at the following location:
`${GOPATH}/github.com/ingensi`


### Requirements

* [Golang](https://golang.org/dl/) 1.7
* [Glide](https://github.com/Masterminds/glide) >= 0.10.0


### Build

To build the binary for Dockbeat run the command below. This will generate a binary
in the same directory with the name dockbeat.

```
make
```
 
## Run dockbeat

Project compilation generate a `dockbeat` executable file in the root directory. To launch dockbeat, run the following command:

```bash
./dockbeat -c etc/dockbeat.yml
```

## Run in a docker container

The easiest way to launch dockbeat is to run it in a container. To achieve this, use the `ingensi/dockbeat` docker image, available on the [docker hub](https://hub.docker.com/r/ingensi/dockbeat/).

Docker run command should:

* mount the target Docker socket to `/var/run/docker.sock`
* link an Elasticsearch node as `elasticsearch`

Example:

```
docker run -d -v /var/run/docker.sock:/var/run/docker.sock \
  --link elastic:elasticsearch ingensi/dockbeat:1.0.0-rc3
```

To override the default configuration, just link yours to `/etc/dockbeat/dockbeat.yml`:

```
docker run -d --link elastic:elasticsearch \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v /volumes/dockbeat-config/:/etc/dockbeat \
  ingensi/dockbeat:1.0.0-rc3
```

By default, when dockbeat is running from this image, it logs into the `/var/log/dockbeat` directory. To access this logs from the host, link a directory to the dockbeat logging directory:
```
docker run -d --link elastic:elasticsearch \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v /volumes/dockbeat-config/:/etc/dockbeat \
  -v /volumes/dockbeat-logs/:/var/logs/dockbeat \
  ingensi/dockbeat:1.0.0-rc3
```

### Configuring Dockbeat

Dockbeat configuration file is located at `etc/dockbeat.yml`. This default template provides the following environment variable mapping:

  - How often to read server statistics 
    - ENV : `PERIOD`
    - Beats variable : `input.period`
    - Default value : `5`
  - Docker socket path
    - ENV : `DOCKER_SOCKET`
    - Beats variable : `input.socket`
    - Default value : `unix:///var/run/docker.sock`
  - Enable TLS encryption
    - ENV : `DOCKER_ENABLE_TLS`
    - Beats variable : `input.tls.enable`
    - Default value : `false`
  - Path to the CA file (when TLS is enabled)
    - ENV : `DOCKER_CA_PATH`
    - Beats variable : `input.tls.ca_path`
    - Default value : no default value
  - Path to the CERT file (when TLS is enabled)
    - ENV : `DOCKER_CERT_PATH`
    - Beats variable : `input.tls.cert_path`
    - Default value : no default value
  - Path to the KEY file (when TLS is enabled)
    - ENV : `DOCKER_KEY_PATH`
    - Beats variable : `input.tls.key_path`
    - Default value : no default value
                                       
When launching it inside a docker container, you can modify the environment variables using the `-e` flag :

```bash
docker run -d \
  -v /var/run/docker.sock:/another/path.sock  \
  --link elastic1:es1 \
  --link elastic2:es2 \
  -e PERIOD=30 \
  -e DOCKER_SOCKET=unix:///another/path.sock \
  ingensi/dockbeat:1.0.0-rc3
```

### Contribute to the project

All contribs are welcome! Read the [CONTRIBUTING](CONTRIBUTING.md) documentation to get more information.

### TL;DR

I want to monitor a host :
(If kibana can't join elastic, check its network configuration.)

```
$ docker network create dockernet

$ docker run -d --net=dockernet --name=elastic \
  -v /mnt/volumes/elastic/config:/usr/share/elasticsearch/config \
  -v /mnt/volumes/elastic/data:/usr/share/elasticsearch/data \
  elasticsearch:2.2.0

$ docker run -d --net=dockernet --name=kibana -p 5601:5601 \
  -e ELASTICSEARCH_URL=http://elastic:9200 \
  kibana:4.4.1

$ docker run -d --net=dockernet --name=dockbeat \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v /mnt/dv/dockbeat:/etc/dockbeat ingensi/dockbeat:latest

```
