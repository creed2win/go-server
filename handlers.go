package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func respWriter(writer http.ResponseWriter, req *http.Request) {
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	writer.WriteHeader(200)
	writer.Write([]byte("OK"))
}

func (cfg *apiConfig) handlerMetrics(writer http.ResponseWriter, req *http.Request) {
	writer.Header().Add("Content-Type", "text/html; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte(fmt.Sprintf(`<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>
`, cfg.fileserverHits.Load())))
	fmt.Println("hits logged: ", cfg.fileserverHits.Load())
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		cfg.fileserverHits.Add(1)
		fmt.Println("hits added: ", cfg.fileserverHits.Load())
		next.ServeHTTP(writer, req)
	})
}
func (cfg *apiConfig) handleReset(writer http.ResponseWriter, req *http.Request) {
	cfg.fileserverHits.Store(0)
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte("Hits reset to 0."))
	if os.Getenv("PLATFORM") == "dev" {
		//TODO - call query to delete all users from table, but not the table. Need to write the query also
	} else {
		writer.WriteHeader(http.StatusForbidden)
	}

}

func (cfg *apiConfig) handleValidation(writer http.ResponseWriter, req *http.Request) {
	data, err := io.ReadAll(req.Body)
	if err != nil {
		fmt.Println("error reading data", err)
	}
	type respBody struct {
		Valid       bool   `json:"valid"`
		Error       string `json:"error"`
		CleanedBody string `json:"cleaned_body"`
	}
	if len(string(data)) > 140 {
		body := respBody{
			Valid: false,
			Error: "",
		}
		dat, err := json.Marshal(body)
		if err != nil {
			fmt.Println("error marshaling data while invalid", err)
		}

		writer.WriteHeader(400)
		writer.Write(dat)
		return
	}
	type responseBody struct {
		Body string `json:"Body"`
	}

	bodyUnmarsh := responseBody{}

	err = json.Unmarshal(data, &bodyUnmarsh)
	if err != nil {
		fmt.Println("error unmarshaling data", err)
	}

	cleanBody := replaceBadWords(bodyUnmarsh.Body)

	body := respBody{
		CleanedBody: cleanBody,
		Valid:       true,
	}

	dat, err := json.Marshal(body)
	if err != nil {
		fmt.Println("error marshaling data while valid", err)
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(200)
	writer.Write(dat)

}

type User struct {
	Email      string `json:"email"`
	Created_at string `json:"created_at"`
	Updated_at string `json:"updated_at"`
	Id         string `json:"id"`
}

func (cfg *apiConfig) handleUsers(writer http.ResponseWriter, req *http.Request) {
	data, err := io.ReadAll(req.Body)
	if err != nil {
		fmt.Println("error reading req.body", err)
	}
	var userInput struct {
		Email string `json:"email"`
	}
	err = json.Unmarshal(data, &userInput)
	if err != nil {
		fmt.Println("error unmarshaling data", err)
	}
	user, err := cfg.dbQ.CreateUser(req.Context(), userInput.Email)
	if err != nil {
		fmt.Println("error creating user in db", err)
	}

	userOut := User{
		Email:      user.Email,
		Id:         user.ID.String(),
		Created_at: user.CreatedAt.String(),
		Updated_at: user.UpdatedAt.String(),
	}
	userjson, err := json.Marshal(userOut)
	if err != nil {
		fmt.Println("error marshaling json to userjson", err)
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusCreated)
	writer.Write(userjson)
}
