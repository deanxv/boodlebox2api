package common

import "time"

var StartTime = time.Now().Unix() // unit: second
var Version = "v1.1.0"            // this hard coding will be replaced automatically when building, no need to manually change

type ModelInfo struct {
	Model string
	Id    string
}

// 创建映射表（假设用 model 名称作为 key）
var ModelRegistry = map[string]ModelInfo{
	"dall-e-3":             {"dall-e-3", "ec252a5c-cd59-4ca5-b92b-6ee6e6864ebc"},
	"flux-pro":             {"flux-pro", "fabc04cf-662f-4af0-9b55-2fece45a51e7"},
	"ideogram-v2":          {"ideogram-v2", "1e678939-395d-4921-b6ce-d4be3d2e72c4"},
	"stable-diffusion-3.5": {"stable-diffusion-3.5", "9f382632-43b1-41a4-b85f-9a599ea3caf5"},
	"stable-diffusion-xl":  {"stable-diffusion-xl", "9fa7e69d-ee00-471c-bb9b-2f553588325a"},
}

var ImageModelList = []string{
	"dall-e-3",
	"flux-pro",
	"ideogram-v2",
	"stable-diffusion-3.5",
	"stable-diffusion-xl",
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
