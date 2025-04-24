package controller

import (
	"boodlebox2api/common"
	logger "boodlebox2api/common/loggger"
	"boodlebox2api/cycletls"
	"boodlebox2api/model"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/samber/lo"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	errServerErrMsg  = "Service Unavailable"
	responseIDFormat = "chatcmpl-%s"
)

type BoodleClient struct {
	Cookie      string
	UserID      string
	AssistantID string
	UserAgent   string
	HttpClient  *http.Client
	ChatID      string // 固定使用的聊天ID
}

// ImageGenerationRequest 表示图像生成请求
type ImageGenerationRequest struct {
	Model          string `json:"model"`
	Prompt         string `json:"prompt"`
	N              int    `json:"n,omitempty"`
	Size           string `json:"size,omitempty"`
	Quality        string `json:"quality,omitempty"`
	Style          string `json:"style,omitempty"`
	ResponseFormat string `json:"response_format,omitempty"`
}

// ImageGenerationResponse 表示图像生成响应
type ImageGenerationResponse struct {
	Created int64 `json:"created"`
	Data    []struct {
		URL           string `json:"url"`
		RevisedPrompt string `json:"revised_prompt,omitempty"`
	} `json:"data"`
}

// ErrorResponse 表示错误响应
type ErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}

// 模型到assistantId的映射
var modelToAssistantID = map[string]string{
	"dall-e-3":             "ec252a5c-cd59-4ca5-b92b-6ee6e6864ebc",
	"flux-pro":             "fabc04cf-662f-4af0-9b55-2fece45a51e7",
	"ideogram-v2":          "1e678939-395d-4921-b6ce-d4be3d2e72c4",
	"stable-diffusion-3.5": "9f382632-43b1-41a4-b85f-9a599ea3caf5",
	"stable-diffusion-xl":  "9fa7e69d-ee00-471c-bb9b-2f553588325a",
}

// NewBoodleClient 创建一个新的BoodleClient
func NewBoodleClient(cookie, userID, assistantID, chatID string) *BoodleClient {
	return &BoodleClient{
		Cookie:      cookie,
		UserID:      userID,
		AssistantID: assistantID,
		ChatID:      chatID,
		UserAgent:   "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36",
		HttpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// HandleImageGenerationRequest 处理图像生成请求
func (c *BoodleClient) ImagesForOpenAI(ctx *gin.Context) {
	var req ImageGenerationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Error: struct {
				Message string `json:"message"`
				Type    string `json:"type"`
			}{
				Message: "无效的请求参数",
				Type:    "invalid_request_error",
			},
		})
		return
	}

	logger.Debug(ctx.Request.Context(), fmt.Sprintf("收到图像生成请求: %+v", req))

	// 从模型映射获取assistantId
	assistantID, ok := modelToAssistantID[req.Model]
	if !ok {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Error: struct {
				Message string `json:"message"`
				Type    string `json:"type"`
			}{
				Message: "不支持的模型",
				Type:    "invalid_request_error",
			},
		})
		return
	}

	// 创建带超时的上下文
	//reqCtx, cancel := context.WithTimeout(ctx.Request.Context(), 120*time.Second)
	//defer cancel()

	// 生成图像
	result, err := c.GenerateImage(ctx, req.Prompt, assistantID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: struct {
				Message string `json:"message"`
				Type    string `json:"type"`
			}{
				Message: err.Error(),
				Type:    "api_error",
			},
		})
		return
	}

	// 返回成功响应
	ctx.JSON(http.StatusOK, ImageGenerationResponse{
		Created: time.Now().Unix(),
		Data: []struct {
			URL           string `json:"url"`
			RevisedPrompt string `json:"revised_prompt,omitempty"`
		}{
			{
				URL:           result.URL,
				RevisedPrompt: result.RevisedPrompt,
			},
		},
	})
}

// OpenaiModels @Summary OpenAI模型列表接口
// @Description OpenAI模型列表接口
// @Tags OpenAI
// @Accept json
// @Produce json
// @Param Authorization header string true "Authorization API-KEY"
// @Success 200 {object} common.ResponseResult{data=model.OpenaiModelListResponse} "成功"
// @Router /v1/models [get]
func OpenaiModels(c *gin.Context) {
	var modelsResp []string

	modelsResp = lo.Union(common.GetModelList())

	var openaiModelListResponse model.OpenaiModelListResponse
	var openaiModelResponse []model.OpenaiModelResponse
	openaiModelListResponse.Object = "list"

	for _, modelResp := range modelsResp {
		openaiModelResponse = append(openaiModelResponse, model.OpenaiModelResponse{
			ID:     modelResp,
			Object: "model",
		})
	}
	openaiModelListResponse.Data = openaiModelResponse
	c.JSON(http.StatusOK, openaiModelListResponse)
	return
}

func safeClose(client cycletls.CycleTLS) {
	if client.ReqChan != nil {
		close(client.ReqChan)
	}
	if client.RespChan != nil {
		close(client.RespChan)
	}
}

// GenerateImage 生成图像
func (c *BoodleClient) GenerateImage(reqCtx *gin.Context, prompt string, assistantID string) (*ImageResult, error) {
	// 获取WebSocket票据
	ticket, err := c.GetWSTicket(reqCtx)
	if err != nil {
		return nil, fmt.Errorf("获取WebSocket票据失败: %v", err)
	}

	// 连接WebSocket
	conn, err := c.ConnectWebSocket(reqCtx, ticket)
	if err != nil {
		return nil, fmt.Errorf("连接WebSocket失败: %v", err)
	}
	defer conn.Close()

	// 创建结果通道和错误通道
	resultChan := make(chan *ImageResult, 1)
	errorChan := make(chan error, 1)

	// 先发送消息，获取submissionID
	promptMessage := fmt.Sprintf("%s", prompt)
	submissionID, err := c.SendMessage(reqCtx, promptMessage, assistantID)
	if err != nil {
		return nil, fmt.Errorf("发送消息失败: %v", err)
	}

	logger.Debug(reqCtx, fmt.Sprintf("获取到submissionID: %s", submissionID))

	// 启动消息接收协程
	go func() {
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Printf("WebSocket读取错误: %v", err)
				errorChan <- fmt.Errorf("WebSocket读取错误: %v", err)
				return
			}

			// 解析消息为通用结构
			var msg map[string]interface{}
			if err := json.Unmarshal(message, &msg); err != nil {
				log.Printf("解析消息错误: %v", err)
				continue
			}

			// 提取数据部分
			data, ok := msg["data"].(map[string]interface{})
			if !ok {
				continue
			}

			// 检查是否是当前提交ID的消息
			msgSubmissionID, _ := data["submissionId"].(string)
			if msgSubmissionID != submissionID {
				continue // 不是当前请求的响应，跳过
			}

			// 检查消息类型
			msgType, _ := data["type"].(string)

			// 如果是最终回复
			if msgType == "MessageFinalResponse" {
				// 提取消息部分
				messageParts, ok := data["message"].([]interface{})
				if ok && len(messageParts) > 0 {
					// 遍历消息部分
					for _, part := range messageParts {
						partMap, ok := part.(map[string]interface{})
						if !ok {
							continue
						}

						partType, _ := partMap["type"].(string)

						// 处理图像内容
						if partType == "Image" {
							imageUrl, ok := partMap["imageUrl"].(string)
							title, titleOk := partMap["title"].(string)

							if ok {
								logger.Debug(reqCtx, fmt.Sprintf("找到图像URL: %s", imageUrl))
								//fmt.Println(fmt.Sprintf("找到图像URL: %s", imageUrl))
								result := &ImageResult{
									URL: imageUrl,
								}

								// 如果有title，将其作为revised_prompt
								if titleOk {
									result.RevisedPrompt = title
								}

								resultChan <- result
								return
							}
						}
					}
				}
			}
		}
	}()

	// 发送活跃状态
	if err := c.SendActiveStatus(reqCtx, conn); err != nil {
		return nil, fmt.Errorf("发送活跃状态失败: %v", err)
	}

	// 检查消息状态
	if err := c.CheckMessageStatus(submissionID); err != nil {
		return nil, fmt.Errorf("检查消息状态失败: %v", err)
	}

	// 等待结果或错误或超时
	select {
	case result := <-resultChan:
		return result, nil
	case err := <-errorChan:
		return nil, err
	case <-time.After(120 * time.Second):
		return nil, fmt.Errorf("请求超时")
	}
}

// SendMessage 发送消息
func (c *BoodleClient) SendMessage(ctx *gin.Context, message string, assistantID string) (string, error) {
	httpURL := fmt.Sprintf("https://box.boodle.ai/api/chat/%s/message", c.ChatID)
	httpRequestBody := map[string]interface{}{
		"mentions": []interface{}{},
		"message": map[string]interface{}{
			"content": message,
			"type":    "PlainText",
		},
		"assistantId":        assistantID,
		"fullTextSearchUsed": false,
	}

	jsonBody, _ := json.Marshal(httpRequestBody)

	req, err := http.NewRequest("POST", httpURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://box.boodle.ai")
	req.Header.Set("Referer", fmt.Sprintf("https://box.boodle.ai/c/%s", c.ChatID))
	req.Header.Set("Cookie", c.Cookie)
	req.Header.Set("Vary", "*")

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("发送HTTP请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %v", err)
	}

	logger.Debug(ctx, fmt.Sprintf("发送消息响应: %s", string(body)))

	var respData map[string]interface{}
	if err := json.Unmarshal(body, &respData); err == nil {
		if id, ok := respData["id"].(string); ok {
			logger.Debug(ctx, fmt.Sprintf("从响应获取到提交ID: %s", id))
			return id, nil
		}
	}

	return "", fmt.Errorf("无法从响应中获取提交ID")
}

// SendActiveStatus 发送活跃状态
func (c *BoodleClient) SendActiveStatus(ctx *gin.Context, conn *websocket.Conn) error {
	activeMsg := map[string]interface{}{
		"chatId": c.ChatID,
		"type":   "ChatMemberActive",
		"userId": c.UserID,
	}
	activeJSON, _ := json.Marshal(activeMsg)
	logger.Debug(ctx, fmt.Sprintf("发送活跃状态"))
	return conn.WriteMessage(websocket.TextMessage, activeJSON)
}

// CheckMessageStatus 检查消息状态
func (c *BoodleClient) CheckMessageStatus(submissionID string) error {
	statusURL := fmt.Sprintf("https://box.boodle.ai/api/chat/%s/message/%s", c.ChatID, submissionID)

	req, err := http.NewRequest("GET", statusURL, nil)
	if err != nil {
		return fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Cookie", c.Cookie)
	req.Header.Set("Referer", fmt.Sprintf("https://box.boodle.ai/c/%s", c.ChatID))

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return fmt.Errorf("检查消息状态失败: %v", err)
	}
	defer resp.Body.Close()

	return nil
}

// GetWSTicket 获取WebSocket票据
func (c *BoodleClient) GetWSTicket(ctx *gin.Context) (string, error) {
	url := "https://box.boodle.ai/api/user/ws-ticket"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Referer", fmt.Sprintf("https://box.boodle.ai/c/%s", c.ChatID))
	req.Header.Set("Cookie", c.Cookie)

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("获取ticket失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %v", err)
	}

	ticket := strings.Trim(string(body), "\"")
	logger.Debug(ctx.Request.Context(), fmt.Sprintf("收到图像生成请求: %+v", req))

	logger.Debug(ctx, fmt.Sprintf("获取到ticket: %s", ticket))
	return ticket, nil
}

// ConnectWebSocket 连接WebSocket
func (c *BoodleClient) ConnectWebSocket(ctx *gin.Context, ticket string) (*websocket.Conn, error) {
	url := fmt.Sprintf("wss://box.boodle.ai/api/v2/parrot/connect/user/%s/ticket/%s", c.UserID, ticket)

	header := http.Header{}
	header.Add("Origin", "https://box.boodle.ai")
	header.Add("Cache-Control", "no-cache")
	header.Add("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	header.Add("Pragma", "no-cache")
	header.Add("User-Agent", c.UserAgent)
	header.Add("Cookie", c.Cookie)

	dialer := websocket.Dialer{
		EnableCompression: true,
	}

	conn, _, err := dialer.Dial(url, header)
	if err != nil {
		return nil, fmt.Errorf("无法连接到WebSocket: %v", err)
	}

	logger.Debug(ctx, fmt.Sprintf("已连接到WebSocket服务器"))
	return conn, nil
}

type ImageResult struct {
	URL           string
	RevisedPrompt string
}
