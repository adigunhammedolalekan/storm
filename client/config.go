package client

type Config struct {
	ServerUrl      string `json:"server_url"`
	ServerAuthCode string `json:"server_auth_code"`
	AppName        string `json:"app_name"`
}
