package storm

import (
	"bytes"
	"fmt"
	"testing"
)

func TestParseConfig(t *testing.T) {
	data := `{"registry": {"username": "username", "password": "password", "url": "url"}, "server_auth_token": "token"}`
	buf := bytes.NewBufferString(data)
	cfg, err := parseConfig(buf)
	if err != nil {
		t.Fatal(err)
	}
	reg := cfg.Registry
	assertString(t, reg.Username, "username", fmt.Sprintf("expected reg.Username to be username, %s gotten", reg.Username))
	assertString(t, reg.Password, "password", fmt.Sprintf("expected reg.Password to be password, %s gotten", reg.Password))
	assertString(t, cfg.ServerAuthToken, "token", fmt.Sprintf("expected cfg.ServerToken to be token, %s gotten", cfg.ServerAuthToken))
}

func assertString(t *testing.T, value, expected, message string) {
	if value != expected {
		t.Fatalf(message)
	}
}
