package storm

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type serviceHttpHandler struct {
	docker DockerService
}

func newServiceHttpHandler(docker DockerService) *serviceHttpHandler {
	return &serviceHttpHandler{docker:docker}
}

func (handler *serviceHttpHandler) deploymentHandler(w http.ResponseWriter, r *http.Request) {
	file, _, err := r.FormFile("bin")
	appName := r.FormValue("app_name")
	if appName == "" {
		handler.badRequest(w, "bad request: app name is missing")
		return
	}
	appBuildFolder := fmt.Sprintf("%s%s", appName, "Build")
	if err != nil {
		log.Println(err)
		handler.badRequest(w, "bad request: bin is missing")
		return
	}
	tag, err := handler.docker.BuildImage(context.Background(), appBuildFolder, appName, file)
	if err != nil {
		log.Println(err)
		handler.respond(w, http.StatusInternalServerError, &Response{Error: true, Message: "failed to build docker image"})
		return
	}

	if err := handler.docker.PushImage(context.Background(), tag); err != nil {
		log.Println(err)
		handler.respond(w, http.StatusInternalServerError, &Response{Error: true, Message: "failed to push image to local registry"})
		return
	}
	handler.ok(w, &Response{Error: false, Message: "success", Data: struct {
		PullUrl string `json:"pull_url"`
	}{PullUrl: tag}})
}

func (handler *serviceHttpHandler) secureMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

func (handler *serviceHttpHandler) respond(w http.ResponseWriter, code int, message *Response)  {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(message.bytes())
}

func (handler *serviceHttpHandler) badRequest(w http.ResponseWriter, message string) {
	handler.respond(w, http.StatusBadRequest, &Response{Error: true, Message: message})
}

func (handler *serviceHttpHandler) ok(w http.ResponseWriter, r *Response) {
	handler.respond(w, http.StatusOK, r)
}

type Response struct {
	Error bool `json:"error"`
	Message string `json:"message"`
	Data interface{} `json:"data"`
}

func (r *Response) bytes() []byte {
	data, err := json.Marshal(r)
	if err != nil {
		return nil
	}
	return data
}