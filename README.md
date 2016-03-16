### Dockerbeat

(if you're on the fast lane, check the TL;DR at the bottom of the readme)

Build status : [![Build Status](https://travis-ci.org/Ingensi/dockerbeat.svg?branch=develop)](https://travis-ci.org/Ingensi/dockerbeat)

Test coverage : [![codecov.io](http://codecov.io/github/Ingensi/dockerbeat/coverage.svg?branch=develop)](http://codecov.io/github/Ingensi/dockerbeat?branch=develop)

Dockerbeat is the [Beat](https://www.elastic.co/products/beats) used for docker daemon monitoring. It is a lightweight agent that installed on your servers, reads periodically docker container statistics and indexes them in Elasticsearch.

We've reached the Release Candidate 1 : it's almost stable today, some minor issues can still appear.

#### Exported document types

There are five types of documents exported:

- `type: container` for container attributes
- `type: cpu` for per process container statistics. One document per container is generated.
- `type: net` for container network statistics. One document per network container is generated.
- `type: memory` for container memory statistics. One document per container is generated.
- `type: blkio` for container io access statistics. One document per container is generated.

To get a detailled list of all generated fields, please read the [fields documentation page](doc/fields.asciidoc).

#### Elasticsearch template 

To apply Dockerbeat template (recommended but not required) :

```bash
curl -XPUT 'http://elastic:9200/_template/dockerbeat' -d@etc/dockerbeat.template.json
```
    
#### Build Dockerbeat

To launch Dockerbeat, build and run the executable. Executable can be compiled either with make command (this requires a fully functional golang environment) or in a docker container.

#### Build with make

Simply run the `make` command at the root project directory. Your golang development environment should be fully functional).

#### Build in a container

If you don't have (and don't want) to setup a golang environment in your host, you can run a `make dockermake` to launch compilation into a golang docker container (you just need a fully functionnal docker environment).
 
#### Run dockerbeat

Project compilation generate a `dockerbeat` executable file in the root directory. To launch dockerbeat, run the following command:

```bash
./dockerbeat -c etc/dockerbeat.yml
```

#### Run in a docker container

The easiest way to launch dockerbeat is to run it in a container. To achieve this, use the `ingensi/dockerbeat` docker image, available on the [docker hub](https://hub.docker.com/r/ingensi/dockerbeat/).

Docker run command should:

* mount the target Docker socket to `/var/run/docker.sock`
* link an Elasticsearch node as `elasticsearch`

Example:

```
docker run -d -v /var/run/docker.sock:/var/run/docker.sock \
  --link elastic:elasticsearch ingensi/dockerbeat:1.0.0-rc1
```

To override the default configuration, just link yours to `/etc/dockerbeat/dockerbeat.yml`:

```
docker run -d --link elastic:elasticsearch \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v /volumes/dockerbeat-config/:/etc/dockerbeat \
  ingensi/dockerbeat:1.0.0-rc1
```

### Configuring Dockerbeat

Dockerbeat configuration file is located at `etc/dockerbeat.yml`. This default template provides the following environment variable mapping:

  - How often to read server statistics 
    - ENV : `PERIOD`
    - Beats variable : `input.period`
    - Default value : `5`
  - Where data will be send (array list of elasticsearch nodes)
    - ENV : `ES_HOSTS`
    - Beats variable : `output.elasticsearch.hosts`
    - Default value : `localhost:9200`
  - Docker socket path
    - ENV : `DOCKER_SOCKET`
    - Beats variable : `input.socket`
    - Default value : `unix:///var/run/docker.sock`
  - Name of the Beat 
    - ENV : `SHIPPER_NAME`
    - Beats variable : `shipper.name`
    - Default value : Hostname of the machine
  - Array of tags
    - ENV : `SHIPPER_TAGS`
    - Beats variable : `shipper.tags`
    - Default value : none
  - Dockerbeat log level
    - ENV : `DOCKERBEAT_LOG_LEVEL`
    - Beats variable : `logging.level`
    - Default value : `error`
                                       
When launching it inside a docker container, you can modify the environement variables using the `-e` flag :

```bash
docker run -d \
  -v /var/run/docker.sock:/var/run/docker.sock \
  --link elastic1:es1 \
  --link elastic2:es2 \
  -e SHIPPER_NAME=$(hostname) \
  -e ES_HOSTS=es1:9200,es2,9200 \
  ingensi/dockerbeat:1.0.0-rc1
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

$ docker run -d --net=dockernet --name=dockerbeat \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v /mnt/dv/dockerbeat:/etc/dockerbeat ingensi/dockerbeat:latest

```