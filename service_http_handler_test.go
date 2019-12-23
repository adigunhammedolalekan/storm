package storm

import (
	"bytes"
	"context"
	"github.com/adigunhammedolalekan/storm/mocks"
	"github.com/golang/mock/gomock"
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

	if got, want := w.Code, http.StatusBadRequest; want != got {
		t.Fatalf("wanted code badrequest; got %d instead", got)
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