package storm

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	appsV1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"log"
	"math/rand"
	"net"
	"strings"
	"time"
)

const stormNs = "namespace-storm"
const registrySecretName = "storm-secret"

//go:generate mockgen -destination=mocks/k8s_service_mock.go -package=mocks github.com/adigunhammedolalekan/storm K8sService
type K8sService interface {
	DeployService(tag, name string, envs map[string]string, isLocal bool) (*DeploymentResult, error)
	GetLogs(appName string) (string, error)
}

type DeploymentResult struct {
	Address string
}

type defaultK8sService struct {
	client *kubernetes.Clientset
	config *Config
}

func newDefaultK8sService(client *kubernetes.Clientset, config *Config) K8sService {
	d := &defaultK8sService{client: client, config: config}
	if err := d.createRegistrySecret(); err != nil {
		log.Println(err)
	}
	if err := d.createNameSpace(); err != nil {
		log.Println("Error occurred while creating namespace: ", err)
	}
	return d
}

func (d *defaultK8sService) createNameSpace() error {
	c := d.client.CoreV1().Namespaces()
	if _, err := c.Get(stormNs, metav1.GetOptions{}); err != nil {
		log.Println("Error: ", err, ": Creating namespace...")
		ns := &v1.Namespace{}
		ns.Name = stormNs
		if _, err := c.Create(ns); err != nil {
			return err
		}
	}
	return nil
}

func (d *defaultK8sService) DeployService(tag, name string, envs map[string]string, isLocal bool) (*DeploymentResult, error) {
	var serviceType = v1.ServiceTypeNodePort
	if !isLocal {
		serviceType = v1.ServiceTypeLoadBalancer
	}
	svc, err := d.createService(name, serviceType)
	if err != nil {
		return nil, err
	}
	ports := svc.Spec.Ports
	deployment, err := d.client.AppsV1().Deployments(stormNs).Get(name, metav1.GetOptions{})
	if err == nil && deployment.Name != "" {
		if err := d.client.AppsV1().Deployments(stormNs).Delete(name, &metav1.DeleteOptions{}); err != nil {
			return nil, err
		}
	}
	if err := d.createDeployment(tag, name, envs, svc.Labels, ports); err != nil {
		return nil, err
	}

	addr := ""
	for _, p := range ports {
		if nodePort := p.NodePort; nodePort != 0 {
			addr = fmt.Sprintf("http://localhost:%d", nodePort)
		}
	}
	return &DeploymentResult{Address: addr}, nil
}

func (d *defaultK8sService) createNodePortService(name string) (*v1.Service, error) {
	return d.createService(name, v1.ServiceTypeNodePort)
}

func (d *defaultK8sService) createLoadBalancerService(name string) (*v1.Service, error) {
	return d.createService(name, v1.ServiceTypeLoadBalancer)
}

func (d *defaultK8sService) createService(serviceName string, serviceType v1.ServiceType) (*v1.Service, error) {
	name := serviceName
	client := d.client.CoreV1()
	svc, err := client.Services(stormNs).Get(name, metav1.GetOptions{})
	if err == nil && svc.Name != "" {
		// service already exists, delete it because we'll need to recreate it
		if err := client.Services(stormNs).Delete(name, &metav1.DeleteOptions{}); err != nil {
			return nil, err
		}
	}

	labels := d.createLabel(serviceName)
	svc = &v1.Service{}
	svc.Name = name
	svc.Labels = labels
	svc.Namespace = stormNs
	servicePort := findAvailablePort()
	port := v1.ServicePort{
		Name:       truncString(fmt.Sprintf("%15s", name+"-service-port")),
		Protocol:   "TCP",
		Port:       int32(servicePort),
		TargetPort: intstr.FromInt(servicePort),
	}
	svc.Spec = v1.ServiceSpec{
		Ports:    []v1.ServicePort{port},
		Selector: labels,
		Type:     serviceType,
	}
	return client.Services(stormNs).Create(svc)
}

func (d *defaultK8sService) createDeployment(tag, name string, envs, labels map[string]string, ports []v1.ServicePort) error {
	deployment := &appsV1.Deployment{}
	deployment.Name = name
	deployment.Labels = labels

	container := v1.Container{}
	envVars := make([]v1.EnvVar, 0, len(envs))

	for k, v := range envs {
		envVars = append(envVars, v1.EnvVar{
			Name:  k,
			Value: v,
		})
	}
	var port int32 = 0
	for _, p := range ports {
		if targetPort := p.TargetPort.IntVal; targetPort != 0 {
			port = targetPort
		}
	}
	envVars = append(envVars, v1.EnvVar{Name: "PORT", Value: fmt.Sprintf("%d", port)})
	container.Name = name
	container.Env = envVars
	container.Image = tag
	container.Ports = []v1.ContainerPort{{
		Name:          truncString(fmt.Sprintf("%15s", name+"-port")),
		ContainerPort: port,
		Protocol:      "TCP",
	}}
	container.ImagePullPolicy = v1.PullAlways
	podTemplate := v1.PodTemplateSpec{}
	podTemplate.Labels = labels
	podTemplate.Name = name
	podTemplate.Spec = v1.PodSpec{
		Containers: []v1.Container{
			container,
		},
		ImagePullSecrets: []v1.LocalObjectReference{{Name: registrySecretName}},
	}
	deployment.Spec = appsV1.DeploymentSpec{
		Replicas: Int32(1),
		Selector: &metav1.LabelSelector{MatchLabels: labels},
		Template: podTemplate,
	}
	if _, err := d.client.AppsV1().Deployments(stormNs).Create(deployment); err != nil {
		return err
	}
	return nil
}

func (d *defaultK8sService) createLabel(name string) map[string]string {
	return map[string]string{"web": fmt.Sprintf("%s-service", strings.ToLower(name))}
}

func (d *defaultK8sService) createRegistrySecret() error {
	secret := &v1.Secret{}
	secret.Name = registrySecretName
	secret.Type = v1.SecretTypeDockerConfigJson
	data, err := d.dockerConfigJson()
	if err != nil {
		return err
	}
	secret.Data = map[string][]byte{
		v1.DockerConfigJsonKey: data,
	}
	if _, err := d.client.CoreV1().Secrets(stormNs).Create(secret); err != nil {
		return err
	}
	return nil
}

func (d *defaultK8sService) GetLogs(appName string) (string, error) {
	selector := fmt.Sprintf("web=%s-service", strings.ToLower(appName))
	pods, err := d.getPodsBySelector(selector)
	if err != nil {
		return "", err
	}
	logs := ""
	for _, p := range pods {
		l := d.client.CoreV1().Pods(stormNs).GetLogs(p.Name, &v1.PodLogOptions{})
		r, err := l.Stream()
		if err != nil {
			log.Println("failed to get stream handle for ", p.Name)
			continue
		}
		data, err := ioutil.ReadAll(r)
		if err != nil {
			log.Println("failed to read stream data for ", p.Name)
		}
		logs += "\n"
		logs += string(data)
	}
	return logs, nil
}

func (d *defaultK8sService) getPodsBySelector(selector string) ([]v1.Pod, error) {
	pods, err := d.client.CoreV1().Pods(stormNs).List(metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return nil, err
	}
	return pods.Items, nil
}

// dockerConfigJson returns a json rep of user's
// docker registry auth credentials.
func (d *defaultK8sService) dockerConfigJson() ([]byte, error) {
	type authData struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Email    string `json:"email"`
		Auth     string `json:"auth"`
	}
	username, password := d.config.Registry.Username, d.config.Registry.Password
	ad := authData{
		Username: username,
		Password: password,
	}
	type auths struct {
		Auths map[string]authData `json:"auths"`
	}
	usernamePassword := fmt.Sprintf("%s:%s", username, password)
	encoded := base64.StdEncoding.EncodeToString([]byte(usernamePassword))
	ad.Auth = encoded
	a := &auths{Auths: map[string]authData{
		d.config.Registry.Url: ad,
	}}
	return json.Marshal(a)
}

func findAvailablePort() int {
	port := rand.Intn(59999)
	addr := fmt.Sprintf("localhost:%d", port)
	conn, err := net.DialTimeout("tcp", addr, 5 * time.Second)
	if err != nil {
		return port
	}
	if err := conn.Close(); err != nil { /*no-op*/ }
	return findAvailablePort()
}

func Int32(i int32) *int32 {
	return &i
}

func truncString(s string) string {
	if len(s) < 15 {
		return s
	}
	return s[0:14]
}
