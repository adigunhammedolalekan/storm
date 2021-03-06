package storm

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"strings"
)

const serverAuthKey = "X-Server-Code"
const maxMemory = 32 << 20

type serviceHttpHandler struct {
	docker DockerService
	k8s    K8sService
	cfg    *Config
}

func newServiceHttpHandler(docker DockerService, k8s K8sService, cfg *Config) *serviceHttpHandler {
	return &serviceHttpHandler{docker: docker, k8s: k8s, cfg: cfg}
}

func (handler *serviceHttpHandler) deploymentHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(maxMemory); err != nil {
		handler.respond(w, http.StatusInternalServerError, &Response{Error: true, Message: "failed to parse http form"})
		return
	}
	file, _, err := r.FormFile("bin")
	appName := r.FormValue("app_name")
	if appName == "" {
		handler.badRequest(w, "bad request: app name is missing")
		return
	}
	if err != nil {
		handler.badRequest(w, fmt.Sprintf("file is missing: Error %s", err.Error()))
		return
	}
	appBuildFolder := fmt.Sprintf("%s%s", appName, "Build")
	tag, err := handler.docker.BuildImage(context.Background(), appBuildFolder, appName, file)
	if err != nil {
		handler.respond(w, http.StatusInternalServerError, &Response{Error: true, Message: fmt.Sprintf("failed to build docker image: %s", err.Error())})
		return
	}
	if err := handler.docker.PushImage(context.Background(), tag); err != nil {
		handler.respond(w, http.StatusInternalServerError, &Response{Error: true, Message: "failed to push image to local registry"})
		return
	}
	// copy passed environment variables from form parameters
	envs := make(map[string]string)
	forms := r.PostForm
	for k, v := range forms {
		if k != "app_name" && k != "bin" {
			if len(v) > 0 {
				envs[k] = v[0]
			}
		}
	}

	result, err := handler.k8s.DeployService(tag, strings.ToLower(appName), envs, true)
	if err != nil {
		handler.respond(w, http.StatusInternalServerError, &Response{Error: true, Message: err.Error()})
		return
	}
	handler.ok(w, &Response{Error: false, Message: "success", Data: struct {
		PullUrl   string `json:"pull_url"`
		AccessUrl string `json:"access_url"`
	}{PullUrl: tag, AccessUrl: result.Address}})
}

func (handler *serviceHttpHandler) logsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appName := vars["app"]
	if appName == "" {
		handler.badRequest(w, "error: app_name is missing")
		return
	}
	logs, err := handler.k8s.GetLogs(appName)
	if err != nil {
		handler.respond(w, http.StatusInternalServerError, &Response{Error: true, Message: err.Error()})
		return
	}
	type logResponse struct {
		Logs string `json:"logs"`
	}
	handler.ok(w, &Response{Error: false, Message: "success", Data: logResponse{Logs: logs}})
}

func (handler *serviceHttpHandler) secureMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get(serverAuthKey)
		if authHeader != handler.cfg.ServerAuthToken {
			handler.respond(w, http.StatusForbidden, &Response{Error: true, Message: "forbidden"})
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (handler *serviceHttpHandler) respond(w http.ResponseWriter, code int, message *Response) {
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
	Error   bool        `json:"error"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func (r *Response) bytes() []byte {
	data, err := json.Marshal(r)
	if err != nil {
		return nil
	}
	return data
}
