package storm

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/google/uuid"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const baseBuildDir = "/tmp/mnt/build"
const rawDockerfile = `
FROM alpine:3.2
RUN apk update && apk add --no-cache ca-certificates
ADD . /app
WORKDIR /app
RUN chmod +x /app/%s
ENTRYPOINT [ "/app/%s" ]`

//go:generate mockgen -destination=mocks/docker_service_mock.go -package=mocks github.com/adigunhammedolalekan/storm DockerService
type DockerService interface {
	BuildImage(ctx context.Context, buildDir, name string, r io.Reader) (string, error)
	PushImage(ctx context.Context, tag string) error
}

type defaultDockerService struct {
	client *client.Client
	config *Config
}

func NewDockerService(cli *client.Client, cfg *Config) DockerService {
	return &defaultDockerService{client: cli, config: cfg}
}

func (d *defaultDockerService) BuildImage(ctx context.Context, buildDir, name string, r io.Reader) (string, error) {
	buildCtx, err := d.writeBuild(buildDir, name, r)
	if err != nil {
		return "", err
	}
	tag := d.md5()[:6]
	pushUrl := fmt.Sprintf("%s/%s:%s", d.config.Registry.Url, strings.ToLower(name), tag)
	_, err = d.client.ImageBuild(ctx, buildCtx, types.ImageBuildOptions{
		NoCache:    false,
		Remove:     false,
		Dockerfile: "Dockerfile",
		Tags:       []string{pushUrl},
	})
	if err != nil {
		return "", err
	}
	return pushUrl, nil
}

func (d *defaultDockerService) PushImage(ctx context.Context, tag string) error {
	_, err := d.client.ImagePush(ctx, tag, types.ImagePushOptions{RegistryAuth: d.registryAuthAsBase64()})
	if err != nil {
		return err
	}
	return nil
}

func (d *defaultDockerService) writeBuild(buildDir, name string, r io.Reader) (io.Reader, error) {
	dir := filepath.Join(baseBuildDir, buildDir)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return nil, err
	}
	filename := filepath.Join(dir, name)
	out, err := os.Create(filename)
	if err != nil {
		return nil, err
	}
	// write binary
	_, err = io.Copy(out, r)
	if err != nil {
		return nil, err
	}
	// write Dockerfile
	dockerfilePath := filepath.Join(dir, "Dockerfile")
	dockerfileContent := fmt.Sprintf(rawDockerfile, name, name)
	if err := ioutil.WriteFile(dockerfilePath, []byte(dockerfileContent), os.ModePerm); err != nil {
		return nil, err
	}
	log.Println("creating build ctx from ", dir)
	return d.createBuildContext(dir)
}

func (d *defaultDockerService) createBuildContext(filename string) (io.Reader, error) {
	return archive.Tar(filename, archive.Uncompressed)
}

func (d *defaultDockerService) md5() string {
	m5 := md5.New()
	m5.Write([]byte(uuid.New().String()))
	return fmt.Sprintf("%+x", string(m5.Sum(nil)))
}

func (d *defaultDockerService) registryAuthAsBase64() string {
	authConfig := types.AuthConfig{
		Username: d.config.Registry.Username,
		Password: d.config.Registry.Password,
	}
	encoded, err := json.Marshal(authConfig)
	if err != nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(encoded)
}
