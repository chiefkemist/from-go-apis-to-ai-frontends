package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
)

//go:embed *
var content embed.FS

type ImageUpload struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Blob  string `json:"blob"`
}

type ImageUploadStatus struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Status string `json:"status"`
}

const schema = `

import "strings"

// Image upload contract
#ImageUpload: {
	// Unique identifier
	id: string & =~"^[0-9a-zA-Z -]{36}$"

	// Image title
	title: string & =~"^.{3,100}$" & =~"^[A-Za-z0-9 -_.]+$"

	// Base64 encoded image
	blob: string & strings.MinRunes(3) & strings.MaxRunes(13_900_000) & =~"^data:image/(jpeg|png|gif|webp);base64,[A-Za-z0-9+/]+=*$"
}

// Image upload status
#ImageUploadStatus: {
	// Unique identifier
	id: string & =~"^[0-9a-zA-Z -]{36}$"

	// Image title
	title: string & strings.MinRunes(3) & strings.MaxRunes(100) & =~"^[A-Za-z0-9 -_.]+$"

	// Image upload status
	status: string & strings.MinRunes(3) & strings.MaxRunes(300) & =~"^[A-Za-z0-9 -_.]+$"
}

`

var ctx = cuecontext.New()
var compiledSchema = ctx.CompileString(schema)

func validateImageUpload(p ImageUpload) error {
	val := ctx.Encode(p)
	return val.Unify(compiledSchema.LookupPath(cue.ParsePath("#ImageUpload"))).Err()
}

func validateImageUploadStatus(p ImageUploadStatus) error {
	val := ctx.Encode(p)
	return val.Unify(compiledSchema.LookupPath(cue.ParsePath("#ImageUploadStatus"))).Err()
}

func processImageUploadHandler(w http.ResponseWriter, r *http.Request) {

	var (
		image  ImageUpload
		status ImageUploadStatus
	)
	if err := json.NewDecoder(r.Body).Decode(&image); err != nil {
		fmt.Println(fmt.Errorf("BAD_PAYLOAD:::: +%v", err))
		status = ImageUploadStatus{
			ID:     "bad-id",
			Title:  "bad-title",
			Status: err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(status)
		return
	}

	if err := validateImageUpload(image); err != nil {
		fmt.Println(fmt.Errorf("INVALID_PAYLOAD:::: +%v with image size: %d", err, len(image.Blob)))
		status = ImageUploadStatus{
			ID:     "bad-id",
			Title:  "bad-title",
			Status: err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	status = ImageUploadStatus{
		ID:     image.ID,
		Title:  image.Title,
		Status: "Successfully processed",
	}
	json.NewEncoder(w).Encode(status)
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
