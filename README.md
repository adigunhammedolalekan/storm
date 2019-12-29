## Storm
A tool to deploy Go apps in kubernetes cluster

## How it works

### Requirements
* Go 1.11+
* Kubernetes
* Docker and docker-compose

## Installation
#### Server Installation
...

#### Client Installation
`go install github.com/adigunhammedolalekan/storm/client/cmd storm`

### Usage
In your Go project directory; execute
1. `storm init` To initialize/setup a project to be deployed by storm, and then
2. `storm deploy` to deploy a project
