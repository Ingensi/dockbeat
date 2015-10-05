[![Build Status](https://travis-ci.org/Ingensi/dockerbeat.svg?branch=develop)](https://travis-ci.org/Ingensi/dockerbeat)
[![codecov.io](http://codecov.io/github/Ingensi/dockerbeat/coverage.svg?branch=develop)](http://codecov.io/github/Ingensi/dockerbeat?branch=develop)

# dockerbeat

Dockerbeat is the [Beat](https://www.elastic.co/products/beats) used for
docker daemon monitoring. It is a lightweight agent that installed on your servers,
reads periodically docker container statistics and indexes them in
Elasticsearch.

This is quite early stage and not yet released.

## Exported document types

There are four types of documents exported:

- `type: container` for container attributes
- `type: cpu` for per process container statistics. One per container is generated.
- `type: net` for container network statistics. One per container is generated.
- `type: memory` for container memory statistics. One per container is generated.

### Container type

<pre>
{
  "_index": "dockerbeat-2015.10.02",
  "_type": "container",
  "_id": "AVAow1NYKDyuAT4RG9KO",
  "_score": null,
  "_source": {
    "container": {
      "command": "/docker-entrypoint.sh kibana",
      "created": "2015-08-10T13:33:10Z",
      "id": "7e91fbb0c7885f55ef8bf9402bbe4b366f88224c8baf31d36265061aa5ba2735",
      "image": "kibana",
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
      "status": "Up 2 minutes"
    },
    "containerID": "7e91fbb0c7885f55ef8bf9402bbe4b366f88224c8baf31d36265061aa5ba2735",
    "containerNames": [
      "/kibana"
    ],
    "count": 1,
    "shipper": "0b42b9dded44",
    "timestamp": "2015-10-02T13:35:00.338Z",
    "type": "container"
  },
  "fields": {
    "timestamp": [
      1443792900338
    ]
  },
  "sort": [
    1443792900338
  ]
}
</pre>

### cpu type

<pre>
{
  "_index": "dockerbeat-2015.10.02",
  "_type": "cpu",
  "_id": "AVAoxIvYKDyuAT4RG9NL",
  "_score": null,
  "_source": {
    "containerID": "0b42b9dded44697f4f7e40a3090b3cdb20f679b8718818b44cd20f215c0f14d2",
    "containerNames": [
      "/furious_mayer"
    ],
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
    "shipper": "0b42b9dded44",
    "timestamp": "2015-10-02T13:36:20.338Z",
    "type": "cpu"
  },
  "fields": {
    "timestamp": [
      1443792980338
    ]
  },
  "sort": [
    1443792980338
  ]
}
</pre>

### net type

<pre>
{
  "_index": "dockerbeat-2015.10.02",
  "_type": "net",
  "_id": "AVAo0iRPKDyuAT4RG9ul",
  "_score": null,
  "_source": {
    "containerID": "0b42b9dded44697f4f7e40a3090b3cdb20f679b8718818b44cd20f215c0f14d2",
    "containerNames": [
      "/furious_mayer"
    ],
    "count": 1,
    "net": {
      "rxBytes_ps": 332.3976262820712,
      "rxDropped_ps": 0,
      "rxErrors_ps": 0,
      "rxPackets_ps": 0.3999971435403985,
      "txBytes_ps": 1231.3912063891169,
      "txDropped_ps": 0,
      "txErrors_ps": 0,
      "txPackets_ps": 0.5999957153105978
    },
    "shipper": "0b42b9dded44",
    "timestamp": "2015-10-02T13:51:11.338Z",
    "type": "net"
  },
  "fields": {
    "timestamp": [
      1443793871338
    ]
  },
  "sort": [
    1443793871338
  ]
}
</pre>

### memory type

<pre>
{
  "_index": "dockerbeat-2015.10.02",
  "_type": "memory",
  "_id": "AVAo0vsnKDyuAT4RG9ww",
  "_score": null,
  "_source": {
    "containerID": "7e91fbb0c7885f55ef8bf9402bbe4b366f88224c8baf31d36265061aa5ba2735",
    "containerNames": [
      "/kibana"
    ],
    "count": 1,
    "memory": {
      "failcnt": 0,
      "limit": 7950745600,
      "maxUsage": 77565952,
      "usage": 77475840,
      "usage_p": 0.9744474782339911
    },
    "shipper": "0b42b9dded44",
    "timestamp": "2015-10-02T13:52:06.338Z",
    "type": "memory"
  },
  "fields": {
    "timestamp": [
      1443793926338
    ]
  },
  "sort": [
    1443793926338
  ]
}
</pre>

## Elasticsearch template

To apply dockerbeat template:

    curl -XPUT 'http://localhost:9200/_template/dockerbeat' -d@etc/dockerbeat.template.json

## Run in a docker container

To launch dockerbeat in a container, use the `ingensi/dockerbeat` docker image, available on the [docker hub](https://hub.docker.com/r/ingensi/dockerbeat/).
 
 Docker run command should:
 
 * mount the target Docker socket to `/var/run/docker.sock`
 * link an Elasticsearch node as `elasticsearch`
 
 Example:

 ```
 docker run -d -v /var/run/docker.sock:/var/run/docker.sock --link elastic:elasticsearch ingensi/dockerbeat:1.0.0-beta1
 ```
 
 To override the default configuration, just link yours to `/etc/dockerbeat/dockerbeat.yml`:
 
 ```
  docker run -d -v /var/run/docker.sock:/var/run/docker.sock -v /your/custom/conf.yml:/etc/dockerbeat/dockerbeat.yml --link elastic:elasticsearch ingensi/dockerbeat:1.0.0-beta1
  ```