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


// func validateChirp(chirp string) {
// 	decoder := json.NewDecoder(r.Body)
//     params := chirpToValidate{}
//     err := decoder.Decode(&params)
//     if err != nil {
// 		log.Printf("Error decoding parameters: %s", err)
// 		code := 500
// 		msg := fmt.Sprintf("Error decoding parameters: %s", err)
// 		respondWithError(w, code, msg) 
//     }

// 	if len(chirp) <= 140 {
// 		code := 200
// 		respondWithJSON(w, code, params)
// 	} else if len(chirp) > 140 {
// 		code := 400
// 		msg := fmt.Sprintf("Chirp is too long")
// 		respondWithError(w, code, msg) 
// 	} else {
// 		code := 500
// 		msg := fmt.Sprintf("Unknow Error")
// 		respondWithError(w, code , msg) 
// 	}
// }


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