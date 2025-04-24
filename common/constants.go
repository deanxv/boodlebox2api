package common

import "time"

var StartTime = time.Now().Unix() // unit: second
var Version = "v1.0.0"            // this hard coding will be replaced automatically when building, no need to manually change

type ModelInfo struct {
	Model     string
	Source    string
	MaxTokens int
}

// 创建映射表（假设用 model 名称作为 key）
var ModelRegistry = map[string]ModelInfo{
	"claude-3-7-sonnet-20250219":          {"claude-3-7-sonnet-20250219", "claude", 128000},
	"claude-3-7-sonnet-20250219-thinking": {"claude-3-7-sonnet-20250219", "claude", 128000},

	"gemini-2.5-pro-preview-03-25": {"google/gemini-2.5-pro-preview-03-25", "openrouter", 65536},
	"gemini-2.5-flash-preview":     {"google/gemini-2.5-flash-preview", "openrouter", 65536},
	"gpt-4.1":                      {"openai/gpt-4.1", "openrouter", 65536},
}

// 通过 model 名称查询的方法
func GetModelInfo(modelName string) (ModelInfo, bool) {
	info, exists := ModelRegistry[modelName]
	return info, exists
}

func GetModelList() []string {
	var modelList []string
	for k := range ModelRegistry {
		modelList = append(modelList, k)
	}
	return modelList
}
