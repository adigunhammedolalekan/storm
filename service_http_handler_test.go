package storm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/adigunhammedolalekan/storm/mocks"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
)

type fakeReader struct {
	io.Reader
}

func (f *fakeReader) Read(b []byte) (int, error) {
	return 0, nil
}

func TestDeploymentHandlerBadRequest(t *testing.T) {
	buf := &bytes.Buffer{}
	writer := multipart.NewWriter(buf)
	if err := writer.WriteField("app_name", ""); err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/", buf)

	handler := newServiceHttpHandler(nil, nil, &Config{})
	handler.deploymentHandler(w, r)

	if got, want := w.Code, http.StatusInternalServerError; want != got {
		t.Fatalf("wanted code %d; got %d instead", want, got)
	}
}

func TestDeploymentHandlerMissingFile(t *testing.T) {
	buf := &bytes.Buffer{}
	writer := multipart.NewWriter(buf)
	out, err := writer.CreateFormField("app_name")
	if err != nil {
		t.Fatal(err)
	}
	out.Write([]byte(`testApp`))
	if _, err := writer.CreateFormFile("bin", ""); err != nil {
		t.Error(err)
	}
	writer.Close()
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/", buf)
	r.Header.Add("Content-Type", writer.FormDataContentType())

	handler := newServiceHttpHandler(nil, nil, &Config{})
	handler.deploymentHandler(w, r)

	if got, want := w.Code, http.StatusBadRequest; want != got {
		t.Fatalf("wanted code badrequest; got %d instead", got)
	}
}

func TestDeploymentHandlerSuccess(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	testTag := "localhost:5000/test:0909"

	mockDocker := mocks.NewMockDockerService(controller)
	mockK8s := mocks.NewMockK8sService(controller)

	binData := bytes.NewBufferString("BinaryData")
	mockDocker.EXPECT().BuildImage(context.Background(), "testBuild", "test", binData).Return(testTag, nil)
	mockDocker.EXPECT().PushImage(context.Background(), testTag)

	mockK8s.EXPECT().DeployService(testTag, "test", map[string]string{}, true).Return(nil)

	buf := &bytes.Buffer{}
	writer := multipart.NewWriter(buf)
	out, err := writer.CreateFormField("app_name")
	if err != nil {
		t.Error(err)
	}
	out.Write([]byte(`test`))
	_, err = writer.CreateFormFile("bin", "test.txt")
	if err != nil {
		t.Fatal(err)
	}
	writer.Close()

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/", buf)
	r.Header.Add("Content-Type", writer.FormDataContentType())

	handler := newServiceHttpHandler(mockDocker, mockK8s, &Config{})
	handler.deploymentHandler(w, r)

	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf("wants code %d, got %d", want, got)
	}
}

func TestDefaultK8sService_GetLogsBadRequest(t *testing.T) {
	appName := ""
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", fmt.Sprintf("/logs/%s", appName), nil)

	handler := newServiceHttpHandler(nil, nil, &Config{})
	handler.logsHandler(w, r)

	if got, want := w.Code, http.StatusBadRequest; got != want {
		t.Fatalf("inconsistent code; expected %d; got %d", want, got)
	}
}

func TestDefaultK8sService_GetLogsInternalServerError(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	appName := "fooBar"
	mockK8s := mocks.NewMockK8sService(controller)
	mockK8s.EXPECT().GetLogs(appName).Return("", errors.New("ERROR"))

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", fmt.Sprintf("/logs/%s", appName), nil)
	r = mux.SetURLVars(r, map[string]string{"app": appName})

	handler := newServiceHttpHandler(nil, mockK8s, &Config{})
	handler.logsHandler(w, r)

	var response struct{
		Error bool `json:"error"`
		Message string `json:"message"`
		Data struct{
			Logs string `json:"logs"`
		}
	}
	if want, got := http.StatusInternalServerError, w.Code; got != want {
		t.Fatalf("Error: expected status code %d; got %d", want, got)
	}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}
	assertString(t, "ERROR", response.Message, fmt.Sprintf("expected response.message to be %s; got %s", "ERROR", response.Message))
}

func TestDefaultK8sService_GetLogs(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	appName := "fooBar"
	mockK8s := mocks.NewMockK8sService(controller)
	mockK8s.EXPECT().GetLogs(appName).Return("hey, log", nil)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", fmt.Sprintf("/logs/%s", appName), nil)
	r = mux.SetURLVars(r, map[string]string{"app": appName})

	handler := newServiceHttpHandler(nil, mockK8s, &Config{})
	handler.logsHandler(w, r)

	var response struct{
		Error bool `json:"error"`
		Message string `json:"message"`
		Data struct{
			Logs string `json:"logs"`
		}
	}
	if want, got := http.StatusOK, w.Code; got != want {
		t.Fatalf("Error: expected status code %d; got %d", want, got)
	}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}
	assertString(t, "hey, log", response.Data.Logs, fmt.Sprintf("expected response %s; got %s", "hey, log", response.Data.Logs))
}