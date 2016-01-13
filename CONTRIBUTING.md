# Contributing to Dockerbeat

Dockerbeat is an open source project and we love to receive contribution from the community!

You can contribute to Dockerbeat by many ways:

* write documentation, tutorial, blog posts
* submit bug reports, new github issues for bugs or new features
* implement new feature

Docker is based on the [Libbeat](https://github.com/elastic/libbeathttps://github.com/elastic/libbeat), an open source
project managed by [Elastic](http://elastic.co).

# Setup a development environment

Dockerbeat development environment is similar to other beats. Please read the official [Beats CONTRIBUTING file](https://github.com/elastic/beats/blob/master/CONTRIBUTING.md)

## Make targets

This repository provide a `makefile` with the following targets:

* **dockerbeat**: make the project (default target)
* **getdeps**: get project dependencies
* **test**: run unit tests
* **updatedeps**: update `Godeps` dependencies
* **dockermake**: build the project into a docker container (this allow to build project on a system without golang dev environment)
* **gofmt**: format go source code
* **cover**: run test and generate cover reports
* **clean**: clean repository