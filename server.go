package storm

import (
	"github.com/docker/docker/client"
	"github.com/gorilla/mux"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

type Server struct {
	router *mux.Router
}

func NewServer(configPath string) (*Server, error) {
	f, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	cfg, err := parseConfig(f)
	if err != nil {
		return nil, err
	}
	dockerClient, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}
	k8sConfigPath := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", k8sConfigPath)
	if err != nil {
		return nil, err
	}
	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	router := mux.NewRouter()
	docker := NewDockerService(dockerClient, cfg)
	k8sService := newDefaultK8sService(k8sClient, cfg)

	handler := newServiceHttpHandler(docker, k8sService, cfg)
	router.Use(handler.secureMW)
	router.HandleFunc("/deploy", handler.deploymentHandler)
	router.HandleFunc("/logs/{app}", handler.logsHandler)
	return &Server{router: router}, nil
}

func (s *Server) Run(addr string) error {
	log.Println("Serving http on ", addr)
	return http.ListenAndServe(addr, s.router)
}
