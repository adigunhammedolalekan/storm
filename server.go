package storm

import (
	"github.com/docker/docker/client"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

type Server struct {
	router *mux.Router
}

func NewServer(configPath string) (*Server, error) {
	cfg, err := parseConfig(configPath)
	if err != nil {
		return nil, err
	}
	dockerClient, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}
	router := mux.NewRouter()
	docker := NewDockerService(dockerClient, cfg)

	handler := newServiceHttpHandler(docker)
	router.Use(handler.secureMW)
	router.HandleFunc("/deploy", handler.deploymentHandler)
	return &Server{router:router}, nil
}
func (s *Server) Run(addr string) error {
	log.Println("Serving http on ", addr)
	return http.ListenAndServe(addr, s.router)
}