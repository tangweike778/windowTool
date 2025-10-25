package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"windowTool/entity"
	"windowTool/qianwen"
)

const (
	typeSmallTalk       = 0 // 闲聊
	typeGenerateImage   = 2 // 生成壁纸
	typeOrganizeDesktop = 1 // 整理桌面
)

func main() {
	// 设置HTTP路由
	http.HandleFunc("/process", processHandler)

	// 启动HTTP服务器
	port := ":8080"
	fmt.Printf("Server is running on port %s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

func loadConfig(path string) (*entity.Config, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %v", err)
	}

	var config entity.Config
	if err := json.Unmarshal(file, &config); err != nil {
		return nil, fmt.Errorf("error parsing config file: %v", err)
	}

	return &config, nil
}

func processHandler(w http.ResponseWriter, r *http.Request) {
	// 加载配置文件
	ctx := r.Context()
	config, err := loadConfig("config.json")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error loading config: %v", err), http.StatusInternalServerError)
		return
	}

	// 只接受POST请求
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 读取请求体
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// 记录原始请求体以便调试
	log.Printf("Raw request body: %s", body)

	// 解析用户请求
	var userReq entity.UserRequest
	if err := json.Unmarshal(body, &userReq); err != nil {
		http.Error(w, fmt.Sprintf("Error parsing request: %v", err), http.StatusBadRequest)
		return
	}

	// 检查用户输入是否包含关键词
	keywords := []string{"分类", "整理桌面", "整理", "收拾"}
	for _, keyword := range keywords {
		if strings.Contains(userReq.Input, keyword) {
			userReq.Type = typeOrganizeDesktop
		}
	}
	keywords = []string{"壁纸", "图片"}
	for _, keyword := range keywords {
		if strings.Contains(userReq.Input, keyword) {
			userReq.Type = typeGenerateImage
		}
	}

	var output string
	if userReq.Type == typeSmallTalk {
		msg := fmt.Sprintf("Input: %s\nSoftwareInfo: %s\nSorfwareLayout:%s", userReq.Input, userReq.SoftwareInfo, userReq.SoftwareLayout)
		// 调用LLM处理用户输入
		output, err = callLLM(msg, userReq.Prompt, config)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error calling LLM: %v", err), http.StatusInternalServerError)
			return
		}
		processOutput(w, output)
	} else if userReq.Type == typeGenerateImage {
		// 调用千问API生成图片
		output = qianwen.Text2Image(ctx, userReq, config)
	} else if userReq.Type == typeOrganizeDesktop {
		// 整理桌面
		promptContent, err := os.ReadFile("classify_prompt.txt")
		if err != nil {
			http.Error(w, fmt.Sprintf("Error reading prompt file: %v", err), http.StatusInternalServerError)
			return
		}
		userReq.Prompt = string(promptContent)
		msg := fmt.Sprintf("Input: %s\nSoftwareInfo: %s\nSorfwareLayout:%s", userReq.Input, userReq.SoftwareInfo, userReq.SoftwareLayout)
		// 调用LLM处理用户输入
		output, err = callLLM(msg, userReq.Prompt, config)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error calling LLM: %v", err), http.StatusInternalServerError)
			return
		}
	}
	processOutput(w, output)
}

func processOutput(w http.ResponseWriter, output string) {
	// 准备响应
	response := entity.UserResponse{
		Output: output,
	}
	// 设置响应头并返回JSON
	w.Header().Set("Content-Type", "application/json")
	res, err := json.Marshal(response)
	if err != nil {
		log.Printf("json marshal fail, err:%v", err)
		http.Error(w, fmt.Sprintf("Error calling LLM: %v", err), http.StatusInternalServerError)
		return
	}
	w.Write(res)
}

func callLLM(input string, prompt string, config *entity.Config) (string, error) {
	// 准备请求体
	llmReq := entity.LLMRequest{
		Model: config.Model,
		Messages: []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		}{
			{
				Role:    "user",
				Content: input,
			},
			{
				Role:    "system",
				Content: prompt,
			},
		},
		Temperature: config.Temperature,
	}

	reqBody, err := json.Marshal(llmReq)
	if err != nil {
		return "", fmt.Errorf("error marshaling LLM request: %v", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequest("POST", config.APIEndpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.APIKey))

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request to LLM: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading LLM response: %v", err)
	}
	log.Printf("LLM response: %s", respBody)

	// 解析响应
	var llmResp entity.LLMResponse
	if err := json.Unmarshal(respBody, &llmResp); err != nil {
		return "", fmt.Errorf("error parsing LLM response: %v", err)
	}

	// 检查是否有响应内容
	if len(llmResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in LLM response")
	}

	return llmResp.Choices[0].Message.Content, nil
}
