## Storm
A tool to deploy Go apps in kubernetes cluster

## How it works
It works by creating/building application binary on the client side, this binary will be uploaded through http `post-form` request. On arrival to
`storm` server, the following events takes place.
1. Uploaded binary is moved to a `tmp` folder along with a predefined `Dockerfile`
2. A docker image will be built for the app using the previously mentioned `Dockerfile`
3. The produced docker image is pushed to the provided `docker registry`
4. A kubernetes service is created at this point. A service of type `LoadBalancer` or `NodePort` is created depending on whether user specified the deployment to be `test/local` or `production`.
A service type `LoadBalancer` is created for `production` and `NodePort` type is creating for `test`
5. A kubernetes deployment is created including all the necessary and user-defined environment variables
6. Repeat

### Requirements
* Go 1.11+
* Kubernetes
* Docker and docker-compose

## Installation
#### Server Installation
...

#### Client Installation
`go install github.com/adigunhammedolalekan/storm/client/cmd`

### Usage
In your Go project directory; execute
1. `storm init` To initialize/setup a project to be deployed by storm, and then
2. `storm deploy` to deploy a project
