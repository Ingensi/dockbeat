# dockerbeat

Dockerbeat is the [Beat](https://www.elastic.co/products/beats) used for
docker daemon monitoring. It is a lightweight agent that installed on your servers,
reads periodically docker container statistics and indexes them in
Elasticsearch.

This is quite early stage and not yet released.

## Exported fields

There are four types of documents exported:
- `type: container` for container attributes
- `type: cpu` for per process container statistics. One per container is generated.
- `type: net` for container network statistics. One per container is generated.
- `type: mem` for container memory statistics. One per container is generated.

<pre>

{
   "_index":"dockerbeat-2015.09.30",
   "_type":"container",
   "_id":"AVAdnKkuup-HrVvNggvv",
   "_score":null,
   "_source":{
      "container":{
         "command":"/docker-entrypoint.sh elasticsearch",
         "created":"2015-07-15T10:35:12+02:00",
         "id":"e2ccef3bcc5aba5afd353285332cc89a1c88475c078e75ac58c7794c3d18114f",
         "image":"elasticsearch",
         "labels":{

         },
         "names":[
            "/elastic",
            "/kibana/elasticsearch"
         ],
         "ports":[
            {
               "ip":"0.0.0.0",
               "privatePort":9200,
               "publicPort":9200,
               "type":"tcp"
            },
            {
               "ip":"0.0.0.0",
               "privatePort":9300,
               "publicPort":9300,
               "type":"tcp"
            }
         ],
         "sizeRootFs":0,
         "sizeRw":0,
         "status":"Up About an hour"
      },
      "containerID":"e2ccef3bcc5aba5afd353285332cc89a1c88475c078e75ac58c7794c3d18114f",
      "containerNames":[
         "/elastic",
         "/kibana/elasticsearch"
      ],
      "count":1,
      "shipper":"localhost",
      "timestamp":"2015-09-30T09:36:57.204Z",
      "type":"container"
   }
}


{
   "_index":"dockerbeat-2015.09.30",
   "_type":"cpu",
   "_id":"AVAdnKkuup-HrVvNggvw",
   "_score":null,
   "_source":{
      "containerID":"e2ccef3bcc5aba5afd353285332cc89a1c88475c078e75ac58c7794c3d18114f",
      "containerNames":[
         "/elastic",
         "/kibana/elasticsearch"
      ],
      "count":1,
      "cpu":{
         "percpuUsage":[
            21170204515,
            21759529205,
            17644558507,
            17160956989
         ],
         "totalUsage":77735249216,
         "usageInKernelmode":12490000000,
         "usageInUsermode":57690000000
      },
      "shipper":"localhost",
      "timestamp":"2015-09-30T09:36:57.204Z",
      "type":"cpu"
   }
}

</pre>

## Elasticsearch template

To apply dockerbeat template:

    curl -XPUT 'http://localhost:9200/_template/dockerbeat' -d@etc/dockerbeat.template.json
