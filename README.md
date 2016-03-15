[![Build Status](https://travis-ci.org/Ingensi/dockerbeat.svg?branch=develop)](https://travis-ci.org/Ingensi/dockerbeat)

[![codecov.io](http://codecov.io/github/Ingensi/dockerbeat/coverage.svg?branch=develop)](http://codecov.io/github/Ingensi/dockerbeat?branch=develop)

### Dockerbeat

Dockerbeat is the [Beat](https://www.elastic.co/products/beats) used for docker daemon monitoring. It is a lightweight agent that installed on your servers, reads periodically docker container statistics and indexes them in Elasticsearch.

We reached the Release Candidate 1 : it's almost stable today, some minor issues can still appear.

#### Exported document types

There are five types of documents exported:

- `type: container`: container attributes
- `type: cpu`: container CPU usage statistics. One document per container is generated.
- `type: net`: container network statistics. One document per network container is generated.
- `type: memory`: container memory statistics. One document per container is generated.
- `type: blkio`: container io access statistics. One document per container is generated.
- `type: daemon`: daemon status information. One document per tick is generated if an error occurred.

To get a detailed list of all generated fields, please read the [fields documentation page](doc/fields.asciidoc).

#### Elasticsearch template 

To apply dockerbeat template (recommended but not required) :

```bash
curl -XPUT 'http://elastic:9200/_template/dockerbeat' -d@etc/dockerbeat.template.json
```
    
#### Build dockerbeat

To launch a dockerbeat, build and run the executable. Executable can be compiled either with make command (this requires a fully functional golang environment) or in a docker container.

#### Build with make

Just Simply run the `make` command at the root project directory. Your golang development environment should be fully functional).

#### Build in a container

If you don't have (and don't want) to setup a golang environment in your host, you can run a `make dockermake` to launch compilation into a golang doker container (you just need a fully functionnal docker environment).
 
#### Run dockerbeat

Project compilation generate a `dockerbeat` executable file in the root directory. To launch dockerbeat, run the following command:

```bash
./dockerbeat -c etc/dockerbeat.yml
```

#### Run as docker container

The easiest way to launch dockerbeat is to run it in a container. To achieve this, use the `ingensi/dockerbeat` docker image, available on the [docker hub](https://hub.docker.com/r/ingensi/dockerbeat/).
 
 Docker run command should:
 
 * mount the target Docker socket to `/var/run/docker.sock`
 * link an Elasticsearch node as `elasticsearch`
 
 Example:

 ```
 docker run -d -v /var/run/docker.sock:/var/run/docker.sock --link elastic:elasticsearch ingensi/dockerbeat:1.0.0-beta2
 ```
 
 To override the default configuration, just link yours to `/etc/dockerbeat/dockerbeat.yml`:
 
 ```
  docker run -d -v /var/run/docker.sock:/var/run/docker.sock -v /your/custom/conf.yml:/etc/dockerbeat/dockerbeat.yml --link elastic:elasticsearch ingensi/dockerbeat:1.0.0-beta2
 ```

# Configuring Dockerbeat

Dockerbeat configuration file is located at `etc/dockerbeat.yml`. This default template provides the following environment variable mapping:

| Environment variable   | Beats variable                | Default value                 | Example                                       | Description                                                 |
| ---------------------- | ----------------------------- | ----------------------------- | --------------------------------------------- | ----------------------------------------------------------- |
| `PERIOD`               | `input.period`                | `5`                           | `export PERIOD=10`                            | How often to read server statistics                         |
| `DOCKER_SOCKET`        | `input.socket`                | `unix:///var/run/docker.sock` | `export DOCKER_SOCKET=tcp://127.0.0.1:2376`   | Docker socket path                                          |
| `ES_HOSTS`             | `output.elasticsearch.hosts`  | `localhost:9200`              | `export ES_HOSTS=[\"es1:9200\",\"es2:9200\"]` | Array list of elasticsearch nodes (where data will be send) |
| `SHIPPER_NAME`         | `shipper.name`                |  Hostname of the machine      | `export SHIPPER_NAME=dockerbeat`              | Name of the Beat                                            |
| `SHIPPER_TAGS`         | `shipper.tags`                |                               | `export SHIPPER_TAGS=[tag1,tag2]`             | Array of tags                                               |
| `DOCKERBEAT_LOG_LEVEL` | `logging.level`               | `error`                       | `export DOCKERBEAT_LOG_LEVEL=debug`           | Dockerbeat log level                                        |
                                                        
This environment can be used when dockerbeat is launched with docker:

```
docker run -d \
  -v /var/run/docker.sock:/var/run/docker.sock \
  --link elastic1:es1 \
  --link elastic2:es2 \
  -e SHIPPER_NAME=$(hostname) \
  -e ES_HOSTS=es1:9200,es2,9200 \
  ingensi/dockerbeat:1.0.0-beta2
```

# Contribute to the project

All contribs are welcome! Read the [CONTRIBUTING](CONTRIBUTING.md) documentation to get more information.
