package kilo_api

import (
	"boodlebox2api/common"
	"boodlebox2api/common/config"
	logger "boodlebox2api/common/loggger"
	"boodlebox2api/cycletls"
	"fmt"
	"github.com/gin-gonic/gin"
	"strings"
)

const (
	baseURL            = "https://kilocode.ai"
	chatEndpoint       = baseURL + "/api/claude/v1/messages"
	openRouterEndpoint = baseURL + "/api/openrouter/chat/completions"
)

func MakeStreamChatRequest(c *gin.Context, client cycletls.CycleTLS, jsonData []byte, cookie string, modelInfo common.ModelInfo) (<-chan cycletls.SSEResponse, error) {
	split := strings.Split(cookie, "=")
	if len(split) >= 2 {
		cookie = split[0]
	}

	headers := make(map[string]string)
	endpoint := ""
	if modelInfo.Source == "claude" {
		endpoint = chatEndpoint
		headers = map[string]string{
			"User-Agent":                  "Rs/JS 0.37.0",
			"Connection":                  "close",
			"Accept":                      "application/json",
			"Accept-Encoding":             "gzip,deflate",
			"Content-Type":                "application/json",
			"x-stainless-lang":            "js",
			"x-stainless-package-version": "0.37.0",
			"x-stainless-os":              "MacOS",
			"x-stainless-arch":            "arm64",
			"x-stainless-runtime":         "node",
			"x-stainless-runtime-version": "v20.18.3",
			"authorization":               fmt.Sprintf("Bearer %s", cookie),
			"anthropic-version":           "2023-06-01",
			"anthropic-beta":              "prompt-caching-2024-07-31,output-128k-2025-02-19",
			"x-stainless-retry-count":     "0",
			"x-stainless-timeout":         "600000",
		}
	} else if modelInfo.Source == "openrouter" {
		endpoint = openRouterEndpoint
		headers = map[string]string{
			"User-Agent":                  "Sa/JS 4.78.1",
			"Connection":                  "close",
			"Accept":                      "application/json",
			"Accept-Encoding":             "gzip,deflate",
			"Content-Type":                "application/json",
			"x-stainless-lang":            "js",
			"x-stainless-package-version": "4.78.1",
			"x-stainless-os":              "MacOS",
			"x-stainless-arch":            "arm64",
			"x-stainless-runtime":         "node",
			"x-stainless-runtime-version": "v20.18.3",
			"authorization":               fmt.Sprintf("Bearer %s", cookie),
			"http-referer":                "https://github.com/Kilo-Org/kilocode",
			"x-title":                     "Kilo Code",
			"x-stainless-retry-count":     "0",
		}
	}

	options := cycletls.Options{
		Timeout: 10 * 60 * 60,
		Proxy:   config.ProxyUrl, // 在每个请求中设置代理
		Body:    string(jsonData),
		Method:  "POST",
		Headers: headers,
	}

	logger.Debug(c.Request.Context(), fmt.Sprintf("cookie: %v", cookie))

	sseChan, err := client.DoSSE(endpoint, options, "POST")
	if err != nil {
		logger.Errorf(c, "Failed to make stream request: %v", err)
		return nil, fmt.Errorf("Failed to make stream request: %v", err)
	}
	return sseChan, nil
}
