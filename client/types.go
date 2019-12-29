package client

type DeploymentResult struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
	Data    struct {
		PullUrl   string `json:"pull_url"`
		AccessUrl string `json:"access_url"`
	}
}
