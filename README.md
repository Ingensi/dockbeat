[![Build Status](https://travis-ci.org/Ingensi/dockerbeat.svg?branch=develop)](https://travis-ci.org/Ingensi/dockerbeat)
[![codecov.io](http://codecov.io/github/Ingensi/dockerbeat/coverage.svg?branch=develop)](http://codecov.io/github/Ingensi/dockerbeat?branch=develop)

# dockerbeat

Dockerbeat is the [Beat](https://www.elastic.co/products/beats) used for
docker daemon monitoring. It is a lightweight agent that installed on your servers,
reads periodically docker container statistics and indexes them in
Elasticsearch.

This is quite early stage and not yet released.

## Exported document types

There are five types of documents exported:

- `type: container`: container attributes
- `type: cpu`: container CPU usage statistics. One document per container is generated.
- `type: net`: container network statistics. One document per network container is generated.
- `type: memory`: container memory statistics. One document per container is generated.
- `type: blkio`: container io access statistics. One document per container is generated.

To get a detailled list of all generated fields, please read the [fields documentation page](doc/fields.asciidoc).

### Container type

<pre>
{
  "_index": "dockerbeat-2016.01.12",
  "_type": "container",
  "_id": "AVI1H82SG7YyM5rPIFuI",
  "_score": null,
  "_source": {
    "@timestamp": "2016-01-12T09:17:02.527Z",
    "beat": {
      "hostname": "machine",
      "name": "machine"
    },
    "container": {
      "command": "/docker-entrypoint.sh kibana",
      "created": "2015-08-10T15:33:10+02:00",
      "id": "7e91fbb0c7885f55ef8bf9402bbe4b366f88224c8baf31d36265061aa5ba2735",
      "image": "5f5f2d8e229dcd39efaca74ae49ee15c8344dd94dc4f0c3333f37a56942d55a5",
      "labels": {},
      "names": [
        "/kibana"
      ],
      "ports": [
        {
          "ip": "0.0.0.0",
          "privatePort": 5601,
          "publicPort": 5601,
          "type": "tcp"
        }
      ],
      "sizeRootFs": 0,
      "sizeRw": 0,
      "status": "Up 15 seconds"
    },
    "containerID": "7e91fbb0c7885f55ef8bf9402bbe4b366f88224c8baf31d36265061aa5ba2735",
    "containerName": "kibana",
    "dockerSocket": "unix:///var/run/docker.sock",
    "count": 1,
    "type": "container"
  },
  "fields": {
    "@timestamp": [
      1452590222527
    ],
    "container.created": [
      1439213590000
    ]
  },
  "sort": [
    1452590222527
  ]
}
</pre>

### cpu type

<pre>
{
  "_index": "dockerbeat-2016.01.12",
  "_type": "cpu",
  "_id": "AVI1H82SG7YyM5rPIFuJ",
  "_score": null,
  "_source": {
    "@timestamp": "2016-01-12T09:17:02.527Z",
    "beat": {
      "hostname": "machine",
      "name": "machine"
    },
    "containerID": "7e91fbb0c7885f55ef8bf9402bbe4b366f88224c8baf31d36265061aa5ba2735",
    "containerName": "kibana",
    "dockerSocket": "unix:///var/run/docker.sock",
    "count": 1,
    "cpu": {
      "percpuUsage": {
        "cpu0": 0,
        "cpu1": 0,
        "cpu2": 0,
        "cpu3": 0
      },
      "totalUsage": 0,
      "usageInKernelmode": 0,
      "usageInUsermode": 0
    },
    "type": "cpu"
  },
  "fields": {
    "@timestamp": [
      1452590222527
    ]
  },
  "sort": [
    1452590222527
  ]
}
</pre>

### net type

<pre>
{
  "_index": "dockerbeat-2016.01.12",
  "_type": "net",
  "_id": "AVI1H82SG7YyM5rPIFuM",
  "_score": null,
  "_source": {
    "@timestamp": "2016-01-12T09:17:02.527Z",
    "beat": {
      "hostname": "machine",
      "name": "machine"
    },
    "containerID": "7e91fbb0c7885f55ef8bf9402bbe4b366f88224c8baf31d36265061aa5ba2735",
    "containerName": "kibana",
    "dockerSocket": "unix:///var/run/docker.sock",
    "count": 1,
    "net": {
      "name": "eth0",
      "rxBytes_ps": 5218.326579188697,
      "rxDropped_ps": 0,
      "rxErrors_ps": 0,
      "rxPackets_ps": 19.199729863640766,
      "txBytes_ps": 5097.328281610544,
      "txDropped_ps": 0,
      "txErrors_ps": 0,
      "txPackets_ps": 19.199729863640766
    },
    "type": "net"
  },
  "fields": {
    "@timestamp": [
      1452590222527
    ]
  },
  "sort": [
    1452590222527
  ]
}
</pre>

### memory type

<pre>
{
  "_index": "dockerbeat-2016.01.12",
  "_type": "memory",
  "_id": "AVI1H82SG7YyM5rPIFuK",
  "_score": null,
  "_source": {
    "@timestamp": "2016-01-12T09:17:02.527Z",
    "beat": {
      "hostname": "machine",
      "name": "machine"
    },
    "containerID": "7e91fbb0c7885f55ef8bf9402bbe4b366f88224c8baf31d36265061aa5ba2735",
    "containerName": "kibana",
    "dockerSocket": "unix:///var/run/docker.sock",
    "count": 1,
    "memory": {
      "failcnt": 0,
      "limit": 7950876672,
      "maxUsage": 74997760,
      "usage": 74817536,
      "usage_p": 0.009409973149687913
    },
    "type": "memory"
  },
  "fields": {
    "@timestamp": [
      1452590222527
    ]
  },
  "sort": [
    1452590222527
  ]
}
</pre>

### blkio type

<pre>
{
  "_index": "dockerbeat-2016.01.12",
  "_type": "blkio",
  "_id": "AVI1H82SG7YyM5rPIFuL",
  "_score": null,
  "_source": {
    "@timestamp": "2016-01-12T09:17:02.527Z",
    "beat": {
      "hostname": "machine",
      "name": "machine"
    },
    "blkio": {
      "read": 0.5999915582387739,
      "total": 0.5999915582387739,
      "write": 0
    },
    "containerID": "7e91fbb0c7885f55ef8bf9402bbe4b366f88224c8baf31d36265061aa5ba2735",
    "containerName": "kibana",
    "dockerSocket": "unix:///var/run/docker.sock",
    "count": 1,
    "type": "blkio"
  },
  "fields": {
    "@timestamp": [
      1452590222527
    ]
  },
  "sort": [
    1452590222527
  ]
}
</pre>

## Elasticsearch template

To apply dockerbeat template:

```bash
curl -XPUT 'http://localhost:9200/_template/dockerbeat' -d@etc/dockerbeat.template.json
```
    
## Build dockerbeat

To launch a dockerbeat, build and run the executable. Executable can be compiled either with make command (this requires a fully functional golang environment) or in a docker container.

### Build with make

Just Simply run the `make` command at the root project directory. Your golang development environment should be fully functional).

### Build in a container

If you don't have (and don't want) to setup a golang environment in your host, you can run a `make dockermake` to launch compilation into a golang doker container (you just need a fully functionnal docker environment).
 
## Run dockerbeat

Project compilation generate a `dockerbeat` executable file in the root directory. To launch dockerbeat, run the following command:

```bash
./dockerbeat -c etc/dockerbeat.yml
```

## Run in a docker container

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
