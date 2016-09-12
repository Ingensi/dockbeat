# Contributing to Dockbeat

Dockbeat is an open source project and we love to receive contribution from the community!

You can contribute to Dockbeat by many ways:

* write documentation, tutorial, blog posts
* submit bug reports, new github issues for bugs or new features
* implement new feature

Dockbeat is based on the [Libbeat](https://github.com/elastic/beats), an open source
project managed by [Elastic](http://elastic.co).

# Setup a development environment

Dockbeat development environment is similar to other beats. Please read the official [Beats CONTRIBUTING file](https://github.com/elastic/beats/blob/master/CONTRIBUTING.md)


## Getting Started with Dockbeat

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


### Run

To run Dockbeat with debugging output enabled, run:

```
./dockbeat -c dockbeat.yml -e -d "*"
```


### Test

To test Dockbeat, run the following command:

```
make testsuite
```

alternatively:
```
make unit-tests
make system-tests
make integration-tests
make coverage-report
```

The test coverage is reported in the folder `./build/coverage/`

### Update

Each beat has a template for the mapping in elasticsearch and a documentation for the fields
which is automatically generated based on `etc/fields.yml`.
To generate etc/dockbeat.template.json, etc/dockbeat.asciidoc and dockbeat.yml configuration file

```
make update
```

To ensure updating dockbeat-docker.yml

```
make fullupdate
```


### Cleanup

To clean  Dockbeat source code, run the following commands:

```
make fmt
make simplify
```

To clean up the build directory and generated artifacts, run:

```
make clean
```


### Clone

To clone Dockbeat from the git repository, run the following commands:

```
mkdir -p ${GOPATH}/github.com/ingensi
cd ${GOPATH}/github.com/ingensi
git clone https://github.com/ingensi/dockbeat
```


For further development, check out the [beat developer guide](https://www.elastic.co/guide/en/beats/libbeat/current/new-beat.html).
