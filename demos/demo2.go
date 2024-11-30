package main


// basic_cueapi.go

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
)

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
	http.HandleFunc("POST /users", userHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
