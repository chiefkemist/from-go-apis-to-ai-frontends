package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
	"gopkg.in/yaml.v3"
)

var (
	apiURL     = "https://api.openai.com/v1/chat/completions" // replace with actual endpoint
	apiKey     = os.Getenv("OPENAI_API_KEY")
	yamlInput  string
	jsonOutput string
	imagePath  string
	prompt     string
	rootCmd    *cobra.Command
)

func init() {
	rootCmd = &cobra.Command{
		Use:   "cli [input.yaml] [output.json]",
		Short: "Utilities for common tasks",
		Long:  "Utilities to streamline common tasks like converting file formats, validating data, and extracting information from images.",
	}

	// Convert Yaml to JSON command
	y2jCmd := &cobra.Command{
		Use:   "y2j [input.yaml] [output.json]",
		Short: "Convert YAML to JSON",
		Args:  cobra.ExactArgs(2),
		RunE:  convertYamlToJson,
	}

	// Get Info from image command
	imgiCmd := &cobra.Command{
		Use:   "imgi [input.webp] [prompt]",
		Short: "Extract info from image",
		Args:  cobra.ExactArgs(2),
		RunE:  getInfoFromImage,
	}

	// Flags
	imgiCmd.Flags().StringVarP(&imagePath, "imagePath", "i", "", "Image file path")
	imgiCmd.Flags().StringVarP(&prompt, "prompt", "p", "", "Prompt for the image")
	y2jCmd.Flags().StringVarP(&yamlInput, "yaml", "y", "", "Yaml input file")
	y2jCmd.Flags().StringVarP(&jsonOutput, "json", "j", "", "Json output file")

	rootCmd.AddCommand(imgiCmd, y2jCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func getInfoFromImage(cmd *cobra.Command, args []string) error {
	imagePath := args[0]
	prompt := args[1]

	fmt.Println(fmt.Sprintf("Extracting Info from: %v\n", imagePath))
	// Load image and encode as base64
	imageBytes, err := os.ReadFile(imagePath)
	if err != nil {
		return fmt.Errorf("Error reading image file: %w", err)
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
		return fmt.Errorf("Error marshalling JSON: %w", err)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("Error creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Error making request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Error reading response body: %w", err)
	}

	responseString := string(body)
	// fmt.Println("Response:", responseString)

	result := gjson.Get(responseString, "choices.0.message.content")

	response := result.Str
	fmt.Println("Response:", response)

	return nil
}

func convertYamlToJson(cmd *cobra.Command, args []string) error {
	inputFile := args[0]
	outputFile := args[1]

	// Read YAML file
	yamlData, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("error reading YAML file: %w", err)
	}

	// Parse YAML
	var data interface{}
	if err := yaml.Unmarshal(yamlData, &data); err != nil {
		return fmt.Errorf("error parsing YAML: %w", err)
	}

	// Convert to JSON
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("error converting to JSON: %w", err)
	}

	// Write JSON file
	if err := os.WriteFile(outputFile, jsonData, 0644); err != nil {
		return fmt.Errorf("error writing JSON file: %w", err)
	}

	fmt.Printf("Successfully converted %s to %s\n", inputFile, outputFile)
	return nil
}
