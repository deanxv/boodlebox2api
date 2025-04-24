package common

import "time"

var StartTime = time.Now().Unix() // unit: second
var Version = "v1.0.0"            // this hard coding will be replaced automatically when building, no need to manually change

type ModelInfo struct {
	Model string
}

// 创建映射表（假设用 model 名称作为 key）
var ModelRegistry = map[string]ModelInfo{
	"dall-e-3":             {"dall-e-3"},
	"flux-pro":             {"flux-pro"},
	"ideogram-v2":          {"ideogram-v2"},
	"stable-diffusion-3.5": {"stable-diffusion-3.5"},
	"stable-diffusion-xl":  {"stable-diffusion-xl"},
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
