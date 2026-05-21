package main

import(
	"fmt"
	"encoding/json"
	"net/http"
	"log"
	"strings"
)


func respondWithError(w http.ResponseWriter, code int, msg string) {

	respBody := chirpError{Error: fmt.Sprintf("%s", msg)}
	data, err := json.Marshal(respBody)
	if err != nil {
		err_msg := fmt.Sprintf("Error marshalling JSON: %s", err)
		log.Printf(err_msg)
		w.WriteHeader(500)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		w.Write(data)
	}

}

func respondWithJSON(w http.ResponseWriter, code int, payload chirpToValidate)  {
	cleaned_body := censorChirp(payload.Body)

	respBody := chirpValid{Valid: true, CleanedBody: cleaned_body}
	data, err := json.Marshal(respBody)
	if err != nil {
		err_msg := fmt.Sprintf("Error marshalling JSON: %s", err)
		log.Printf(err_msg)
		w.WriteHeader(500)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		w.Write(data)
	}


}

func censorChirp(s string) string {
	dirtyWords := []string{"kerfuffle", "sharbert", "fornax"}
	words := strings.Split(s, " ")
	for i, word := range words{
		for _, dirtyWord := range dirtyWords {
			if strings.ToLower(word) == dirtyWord {
				words[i] = "****"
			}
		}
	}

	return strings.Join(words, " ")
}