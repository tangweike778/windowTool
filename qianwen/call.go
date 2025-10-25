package qianwen

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
	"windowTool/entity"

	"github.com/imroc/req"
)

// TextToImageRequest 定义API请求结构
type TextToImageRequest struct {
	Model  string `json:"model"`
	Input  Input  `json:"input"`
	Params Params `json:"parameters"`
}

type Input struct {
	Messages []Message `json:"messages"`
}

type Message struct {
	Role    string    `json:"role"`
	Content []Content `json:"content"`
}

type Content struct {
	Text string `json:"text"`
}

type Params struct {
	NegativePrompt string `json:"negative_prompt,omitempty"`
	PromptExtend   bool   `json:"prompt_extend,omitempty"`
	Watermark      bool   `json:"watermark,omitempty"`
	Size           string `json:"size,omitempty"`
}

// TextToImageResponse 定义API响应结构
type TextToImageResponse struct {
	Output struct {
		Choices []struct {
			FinishReason string `json:"finish_reason"`
			Message      struct {
				Content []struct {
					Image string `json:"image"`
				} `json:"content"`
				Role string `json:"role"`
			} `json:"message"`
		} `json:"choices"`
		TaskMetric struct {
			Failed    int `json:"FAILED"`
			Succeeded int `json:"SUCCEEDED"`
			Total     int `json:"TOTAL"`
		} `json:"task_metric"`
	} `json:"output"`
	Usage struct {
		Height     int `json:"height"`
		ImageCount int `json:"image_count"`
		Width      int `json:"width"`
	} `json:"usage"`
	RequestID string `json:"request_id"`
}

// Text2Image 文生图
func Text2Image(ctx context.Context, userReq entity.UserRequest, config *entity.Config) string {
	apiKey := config.QianwenAPIKey
	apiEndpoint := config.QianwenEndpoint // 示例端点，请替换为实际端点
	// 构建请求数据
	requestData := TextToImageRequest{
		Model: "qwen-image-plus",
		Input: Input{
			Messages: []Message{
				{
					Role: "user",
					Content: []Content{
						{
							Text: userReq.Input,
						},
					},
				},
			},
		},
		Params: Params{
			Size:         userReq.ScreenResolution, // 默认大小
			Watermark:    true,
			PromptExtend: true,
		},
	}

	// 发送请求
	headers := req.Header{
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + apiKey,
	}

	resp, err := req.Post(apiEndpoint, headers, req.BodyJSON(&requestData))
	if err != nil {
		log.Fatal("API请求失败:", err)
	}

	if resp.Response().StatusCode != 200 {
		body, _ := resp.ToString()
		log.Fatalf("API返回错误: %d, 响应: %s", resp.Response().StatusCode, body)
	}

	// 解析响应
	var response TextToImageResponse
	err = resp.ToJSON(&response)
	if err != nil {
		log.Fatal("解析响应失败:", err)
	}

	// 检查是否有结果
	if len(response.Output.Choices) == 0 || len(response.Output.Choices[0].Message.Content) == 0 {
		log.Fatal("API未返回任何图像")
	}

	// 获取第一张图片的URL
	imageURL := response.Output.Choices[0].Message.Content[0].Image
	if imageURL == "" {
		log.Fatal("API返回的图像URL为空")
	}

	// 下载并保存图片
	filename := fmt.Sprintf("generated_image_%d.png", time.Now().Unix())
	err = downloadImage(imageURL, filename)
	if err != nil {
		log.Fatal("下载图片失败:", err)
	}

	fmt.Printf("图片已生成: %s\n", filename)
	return imageURL

}

// 下载图片并保存到本地
func downloadImage(url, filename string) error {
	// 发送HTTP请求获取图片
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载图片失败，状态码: %d", resp.StatusCode)
	}

	// 创建文件
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// 将图片内容写入文件
	_, err = io.Copy(file, resp.Body)
	return err
}
