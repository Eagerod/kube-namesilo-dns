package icanhazip

import (
	"io"
	"net/http"
	"strings"
)

const ICanHazIPUrl string = "https://icanhazip.com"

func GetPublicIP() (string, error) {
	response, err := http.Get(ICanHazIPUrl)
	if err != nil {
		return "", err
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(body)), nil
}
