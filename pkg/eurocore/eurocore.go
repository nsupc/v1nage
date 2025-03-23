package eurocore

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	client       http.Client
	base_url     string
	username     string
	password     string
	token        string
	last_refresh time.Time
}

func (c *Client) refreshToken() error {
	type LoginData struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	type ResponseData struct {
		Token string `json:"token"`
	}

	loginData := LoginData{
		Username: c.username,
		Password: c.password,
	}

	url := fmt.Sprintf("%s/login", c.base_url)
	data, err := json.Marshal(loginData)
	if err != nil {
		return err
	}

	resp, err := c.client.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}

	responseData := ResponseData{}

	err = json.Unmarshal(body, &responseData)
	if err != nil {
		return err
	}

	c.token = responseData.Token
	c.last_refresh = time.Now()

	return nil
}

func (c *Client) SendTelegram(t Telegram) {
	if time.Since(c.last_refresh) > time.Hour {
		err := c.refreshToken()
		if err != nil {
			slog.Error("refresh token error", slog.Any("error", err))
			return
		}
	}

	url := fmt.Sprintf("%s/telegrams", c.base_url)

	telegramList := []Telegram{t}

	data, err := json.Marshal(telegramList)
	if err != nil {
		slog.Error("data marshaling error", slog.Any("error", err))
		return
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		slog.Error("request build error", slog.Any("error", err))
		return
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.token))

	resp, err := c.client.Do(req)
	if err != nil {
		slog.Error("request error", slog.Any("error", err))
		return
	}

	slog.Info("telegram request sent", slog.String("nation", t.Recipient), slog.Int("statusCode", resp.StatusCode))
}

func New(base_url string, username string, password string) *Client {
	base_url = strings.Trim(base_url, "/")

	client := Client{
		client:   http.Client{},
		base_url: base_url,
		username: username,
		password: password,
	}

	return &client
}

type Telegram struct {
	Sender    string `json:"sender"`
	Recipient string `json:"recipient"`
	Id        string `json:"id"`
	Secret    string `json:"secret_key"`
	Type      string `json:"tg_type"`
}
