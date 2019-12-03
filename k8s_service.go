package storm

type K8sService interface {
	DeployService() error
}

