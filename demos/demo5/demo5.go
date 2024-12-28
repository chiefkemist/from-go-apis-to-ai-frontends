package main

import (
	"bufio"
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"github.com/tidwall/gjson"
)

var (
	apiURL = "https://api.openai.com/v1/chat/completions" // replace with actual endpoint
	apiKey = os.Getenv("OPENAI_API_KEY")
)

//go:embed *
var content embed.FS

type ImageUpload struct {
	ID     string `json:"id"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
	Blob   string `json:"blob"`
}

type ImageInfo struct {
	Info string `json:"info"`
}

const schema = `

import "strings"

// Image upload contract
#ImageUpload: {
	// Unique identifier
	id: string & =~"^[0-9a-zA-Z -]{36}$"

	// Image prompt
	prompt: string & =~"^.{3,100}$" & =~"^[A-Za-z0-9 -_.]+$"

	// Stream enabled
	stream: bool

	// Base64 encoded image
	blob: string & strings.MinRunes(3) & strings.MaxRunes(13_900_000) & =~"^data:image/(jpeg|png|gif|webp);base64,[A-Za-z0-9+/]+=*$"
}

// Image info contract
#ImageInfo: {
	// Image info
	info: string
}

`

var ctx = cuecontext.New()
var compiledSchema = ctx.CompileString(schema)

func validateImageUpload(p ImageUpload) error {
	val := ctx.Encode(p)
	return val.Unify(compiledSchema.LookupPath(cue.ParsePath("#ImageUpload"))).Err()
}

func validateImageInfoStatus(p ImageInfo) error {
	val := ctx.Encode(p)
	return val.Unify(compiledSchema.LookupPath(cue.ParsePath("#ImageUploadStatus"))).Err()
}

func processImageUploadHandler(w http.ResponseWriter, r *http.Request) {

	var (
		image  ImageUpload
		status ImageInfo
	)
	if err := json.NewDecoder(r.Body).Decode(&image); err != nil {
		fmt.Println(fmt.Errorf("BAD_PAYLOAD:::: +%v", err))
		status = ImageInfo{
			Info: err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(status)
		return
	}

	if err := validateImageUpload(image); err != nil {
		fmt.Println(fmt.Errorf("INVALID_PAYLOAD:::: +%v with image size: %d", err, len(image.Blob)))
		status = ImageInfo{
			Info: err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(status)
		return
	}

	if image.Stream {
		// Set CORS headers to allow all origins. You may want to restrict this to specific origins in a production environment.
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		//w.Header().Set("Access-Control-Max-Age", "3600")
		// Set the content type to text/event-stream
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		if err := getInfoFromImageStreaming(w, image.Blob, image.Prompt); err != nil {
			fmt.Println(fmt.Errorf("INFO_RETRIEVAL_ERROR:::: +%v with image size: %d", err, len(image.Blob)))
			fmt.Fprintf(w, "data: %s\n\n", err.Error())
			w.(http.Flusher).Flush()
			fmt.Fprintf(w, "data: %s\n\n", "[DONE]")
			w.(http.Flusher).Flush()
			return
		}
	} else {
		if err, info := getInfoFromImage(image.Blob, image.Prompt); err != nil {
			fmt.Println(fmt.Errorf("INFO_RETRIEVAL_ERROR:::: +%v with image size: %d", err, len(image.Blob)))
			status = ImageInfo{
				Info: err.Error(),
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(status)
			return
		} else {
			w.Header().Set("Content-Type", "application/json")
			status = ImageInfo{
				Info: info,
			}
			fmt.Println(fmt.Sprintf("Extracted Image Data; +%v", status))
			json.NewEncoder(w).Encode(status)
		}
	}
}

func main() {
	// API endpoints
	http.HandleFunc("POST /extract-image-info", processImageUploadHandler)

	// Serve OpenAPI spec
	http.HandleFunc("GET /openapi.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		spec, _ := content.ReadFile("openapi.json")
		w.Write(spec)
	})

	// Serve Swagger UI
	http.HandleFunc("GET /docs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		tmpl := template.Must(template.New("swagger").Parse(swaggerTemplate))
		tmpl.Execute(w, nil)
	})

	log.Printf("Server starting on http://localhost:8080")
	log.Printf("API documentation available at http://localhost:8080/docs")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

const swaggerTemplate = `
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8">
    <title>API Documentation</title>
    <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5.10.5/swagger-ui.css">
  </head>
  <body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5.10.5/swagger-ui-bundle.js"></script>
    <script>
      window.onload = function() {
        window.ui = SwaggerUIBundle({
          url: '/openapi.json',
          dom_id: '#swagger-ui',
          deepLinking: true,
          presets: [
            SwaggerUIBundle.presets.apis,
            SwaggerUIBundle.SwaggerUIStandalonePreset
          ],
        });
      };
    </script>
  </body>
</html>
`

func getInfoFromImageStreaming(w http.ResponseWriter, imageUrl, prompt string) error {
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
							"url": imageUrl,
						},
					},
				},
			},
		},
		"stream": true, // Enable streaming
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

	// Check for non-OK status code
	if resp.StatusCode != http.StatusOK {
		respBody, _ := bufio.NewReader(resp.Body).ReadString('\n')
		fmt.Printf("Non-OK status code: %d\nBody: %s\n", resp.StatusCode, respBody)
		return fmt.Errorf("Non-OK status code: %d", resp.StatusCode)
	}

	// Use bufio to read response line by line
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()

		// OpenAI streams each chunk prefixed by "data:"
		if !strings.HasPrefix(line, "data:") {
			continue
		}

		// "data: [DONE]" indicates the end of the stream
		jsonData := strings.TrimPrefix(line, "data: ")
		if jsonData == "[DONE]" {
			fmt.Println("\n--- Stream finished ---")
			fmt.Fprintf(w, "data: %s\n\n", "[DONE]")
			w.(http.Flusher).Flush()
			break
		}

		// Parse each chunk using gjson
		result := gjson.Parse(jsonData)

		// The relevant content is in "choices" -> array -> "delta" -> "content"
		// e.g., result.Get("choices.0.delta.content")
		content := ""
		for _, choice := range result.Get("choices").Array() {
			contentDelta := choice.Get("delta.content").String()
			content += contentDelta
		}
		// Print the chunk of content (if any)
		if content != "" {
			fmt.Print(content)
			fmt.Fprintf(w, "data: %s\n\n", content)
			w.(http.Flusher).Flush()
		}

		// Sleep for a bit to simulate processing time
		time.Sleep(100 * time.Millisecond)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("Error reading streamed response body: %w", err)
	}

	return nil
}

func getInfoFromImage(imageUrl, prompt string) (error, string) {
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
							"url": imageUrl,
						},
					},
				},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("Error marshalling JSON: %w", err), ""
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("Error creating request: %w", err), ""
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Error making request: %w", err), ""
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Error reading response body: %w", err), ""
	}

	responseString := string(body)
	// fmt.Println("Response:", responseString)

	result := gjson.Get(responseString, "choices.0.message.content")

	response := result.Str
	fmt.Println("Response:", response)

	return nil, response
}
