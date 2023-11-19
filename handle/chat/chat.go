package chat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

type ChatRequest struct {
	History []string `json:"history"`
	Query   string   `json:"query"`
	Prompt  string   `json:"prompt"`
}

func (data ChatRequest) Chat() (string, error) {
	body, err := json.Marshal(data)
	if err != nil {
		fmt.Println("json err:", err)
		return "json err", err
	}
	content, err := os.ReadFile("./chat/chat_url.txt")
	if err != nil {
		fmt.Println("url config file err", err)
		return "url config file err", err
	}
	url := string(content)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		fmt.Println("Request err:", err)
		return "Requeset err", err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{} //Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Request do err:", err)
		return "Requeset do", err
	}
	defer resp.Body.Close()
	respbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Response read err:", err)
		return "Response read", err
	}
	return string(respbody), nil
}
