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


//go:embed demo3/*
var content embed.FS

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

const schema = `
#User: {
    id:   string
    name: string & =~"^[A-Za-z ]+$"
}
`

var ctx = cuecontext.New()
var userSchema = ctx.CompileString(schema)

func validateUser(u User) error {
	val := ctx.Encode(u)
	return val.Unify(userSchema.LookupPath(cue.ParsePath("#User"))).Err()
}

func userHandler(w http.ResponseWriter, r *http.Request) {

	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := validateUser(user); err != nil {
		fmt.Println(fmt.Errorf("INVALID_PAYLOAD:::: +%v", err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func main() {
    // API endpoints
	http.HandleFunc("POST /users", userHandler)

    // Serve OpenAPI spec
    http.HandleFunc("GET /openapi.json", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        spec, _ := content.ReadFile("demo3/openapi.json")
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
