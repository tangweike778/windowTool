package entity

// Config 定义配置文件的结构体
type Config struct {
	APIKey          string  `json:"apiKey"`
	Model           string  `json:"model"`
	Temperature     float64 `json:"temperature"`
	APIEndpoint     string  `json:"apiEndpoint"`
	QianwenAPIKey   string  `json:"qianwenAPIKey"`
	QianwenEndpoint string  `json:"qianwenEndpoint"`
}

// LLMRequest 定义请求LLM的结构体
type LLMRequest struct {
	Model    string `json:"model"`
	Messages []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"messages"`
	Temperature float64 `json:"temperature"`
}

// LLMResponse 定义LLM响应的结构体
type LLMResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

type softwareInfo struct {
	Name string         `json:"name"`
	Type string         `json:"type"`
	Docs []softwareInfo `json:"docs"`
}

// UserRequest 定义用户请求的结构体
type UserRequest struct {
	Input            string         `json:"input"`
	Type             int            `json:"type"`             // 0-自然语言聊天 1-桌面整理 2-生成壁纸
	SoftwareLayout   string         `json:"softwareLayout"`   // 软件布局，表示可以放n*m个软件
	ScreenResolution string         `json:"screenResolution"` // 屏幕分辨率
	SoftwareInfo     []softwareInfo `json:"softwareInfo"`
	Prompt           string         `json:"prompt"`
}

// UserResponse 定义返回给用户的结构体
type UserResponse struct {
	Output string `json:"output"`
}
