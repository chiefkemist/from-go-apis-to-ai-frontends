package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/tidwall/gjson"
)

var (
	apiURL = "https://api.openai.com/v1/chat/completions" // replace with actual endpoint
	apiKey = os.Getenv("OPENAI_API_KEY")
)

func main() {
	fmt.Println("Hello, World!\n")
	imgInfo := getInfoFromImage(
		"from-go-apis-to-ai-enhanced-frontends.webp",
		"Describe the scenery in the image.",
	)
	fmt.Println(imgInfo)
}

func getInfoFromImage(imagePath, prompt string) (response string) {
	fmt.Println(fmt.Sprintf("Extracting Info from: %v\n", imagePath))
	// Load image and encode as base64
	imageBytes, err := os.ReadFile(imagePath)
	if err != nil {
		fmt.Println("Error reading image file:", err)
		return
	}
	imageBase64 := base64.StdEncoding.EncodeToString(imageBytes)

	requestBody, err := json.Marshal(map[string]any{
		"model":      "gpt-4o",
		"max_tokens": 4096,
		"messages": []map[string]any{
			{
				"role": "user",
				"content": []map[string]any{
					{
						"type": "text",
						"text": prompt,
					},
					{
						"type": "image_url",
						"image_url": map[string]any{
							"url": fmt.Sprintf("data:image/webp;base64, %s", imageBase64),
						},
					},
				},
			},
		},
	})
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(requestBody))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}

	responseString := string(body)
	// fmt.Println("Response:", responseString)

	result := gjson.Get(responseString, "choices.0.message.content")

	response = result.Str

	return response
}
