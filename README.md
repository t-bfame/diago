# Diago
[![Test](https://github.com/t-bfame/diago/actions/workflows/test.yml/badge.svg)](https://github.com/t-bfame/diago/actions/workflows/test.yml)
[![Coverage](https://codecov.io/gh/t-bfame/diago/branch/dev/graph/badge.svg)](https://codecov.io/gh/t-bfame/diago)


Diago is a Performance Testing Framework, designed to run on Kubernetes. It integrates with various Kubernetes features to be easy to deploy, scalable and keep cost of operations low. Some of its features are:
- **Runs on Kubernetes**: Diago directly runs on Kubernetes, which make it easy to deploy and manage. Diago interacts with the Kubernetes API to spin up Workers, which are responsible for generating load. It can also use the Kubernetes API to simulate disasters for Pods under load.
- **Distributed**: Diago's architecture is based on a Leader - Worker relationship, where the leader is responsible for spinning up Workers to perform the load test. This lets it be more scalable and allows it to scale to massive amounts of load based on the size of the Kubernetes cluster it runs on.
- **Integrated UI**: Diago comes with a UI which is designed to be minimal and easy to use. Not only does the UI show detailed reports generated after the test results, but it also integrates with Kubernetes based tools like Grafana to display useful metrics that were captured during the load test.
- Integration with Prometheus: Diago can exposes aggregated Prometheus style metrics for monitoring systems to scrape. Users extract these metrics and use them with a monitoring system of their choice.
- **Automation Friendly**: Diago exposes REST APIs for automating creation, running, deletion of tests, disaster simulations etc.
- **Expandable**: Diago's worker use gRPC for communication and are designed to be completely decoupled from the Leader. This allows users to write their own workers for custom load tests, which use the gRPC protobufs in [diago-idl](https://github.com/t-bfame/diago-idl).

## Install 
Refer to the wiki for [Installing Diago](https://github.com/t-bfame/diago/wiki/Installation).

Docker images are available on [Docker Hub](https://hub.docker.com/repository/docker/tbfame/diago) and can be pulled by running:
```
docker pull tbfame/diago:latest
```
## Documentation
Refer to the wiki for [Documentation](https://github.com/t-bfame/diago/wiki).


## Building from source

### Submdoules
Diago contains submodules for the following:
- `ui`: The Diago UI React App
- `idl`: The Diago Protobuf files

For any contributions to the submodules, please refer to the repository's documentation itself.

### Generated files
Diago contains files generated using the following programs. These have to be installed before building diago from source:
- [controller-gen](https://book.kubebuilder.io/reference/controller-gen.html): Used to create diago CRDs in `api/`
- [protoc](https://developers.google.com/protocol-buffers/docs/gotutorial#compiling-your-protocol-buffers): Used to generate the protobuf files in `proto-gen/`
- [packr2](https://github.com/gobuffalo/packr/tree/master/v2): Used to bundle UI assets and files in `static` with the binary

The `Makefile` provides some handy commands to compile diago:
- `crd-gen`: Generates the CRD files
- `proto`: Generates the protobuf files
- `ui`: Generates the ui files
- `packr`: Generates `.go` files from static assets

### Compiling the binary
- To compile the diago binary, run the following command:
```RUN CGO_ENABLED=0 GOOS=linux go build cmd/main.go```
- You can also create a docker image by running the following command. This will use multi staged builds in docker to compile the UI and packr assets.
```
docker build -t diago .
```

## More Information
- Diago uses github workflows for CI, check the actions tab.
- Pushes to docker hub are made by the organization members with new releases.

## License
[Apache License 2.0](https://github.com/t-bfame/diago/blob/main/LICENSE)
