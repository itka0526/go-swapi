package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"reflect"
	"time"
	"unicode"

	"github.com/gorilla/mux"
)

const API_URL = "https://swapi.dev/api/"

type Server struct {
	listAddr string
}

func NewServer(listAddr string) *Server {
	return &Server{
		listAddr: listAddr,
	}
}

func WriteJSON(w http.ResponseWriter, statusCode int, v any) error {
	// Makes sense now.
	h := w.Header()
	// Initialize header and then set the header values :D
	h.Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	return json.NewEncoder(w).Encode(v)
}

type MyAPIHandleFunc func(http.ResponseWriter, *http.Request) error

func makeHTTPHandleFunc(f MyAPIHandleFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			WriteJSON(w, http.StatusInternalServerError, struct{ message string }{message: "FUCK YOU!"})
		}
	}
}

func (s *Server) Start() {
	r := mux.NewRouter()

	r.HandleFunc("/test", makeHTTPHandleFunc(s.Characters))

	http.ListenAndServe(s.listAddr, r)
}

type ResponseBackFromSlashPeople struct {
	Count    int         `json:"count"`
	Next     string      `json:"next"`
	Previous any         `json:"previous"`
	Results  []Character `json:"results"`
}

type Character struct {
	Name      string    `json:"name"`
	Height    string    `json:"height"`
	Mass      string    `json:"mass"`
	HairColor string    `json:"hair_color"`
	SkinColor string    `json:"skin_color"`
	EyeColor  string    `json:"eye_color"`
	BirthYear string    `json:"birth_year"`
	Gender    string    `json:"gender"`
	Homeworld string    `json:"homeworld"`
	Films     []string  `json:"films"`
	Species   []any     `json:"species"`
	Vehicles  []string  `json:"vehicles"`
	Starships []string  `json:"starships"`
	Created   time.Time `json:"created"`
	Edited    time.Time `json:"edited"`
	URL       string    `json:"url"`
}

func (s *Server) Characters(w http.ResponseWriter, r *http.Request) error {
	res, err := http.Get(API_URL + "/people")
	if err != nil {
		return err
	}

	wantedFields := r.URL.Query()

	msg, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	var p ResponseBackFromSlashPeople

	if err := json.Unmarshal(msg, &p); err != nil {
		return err
	}

	var characters []Character

	characters = append(characters, p.Results...)

	if len(wantedFields) == 0 {
		return WriteJSON(w, http.StatusOK, characters)
	}

	var field404 error

	askedFields := make([]map[string]interface{}, len(characters))

o:
	for fieldKey := range wantedFields {
		for idx, chr := range characters {
			val := reflect.ValueOf(chr).FieldByName(goCase(fieldKey))

			if val.IsValid() {
				if askedFields[idx] == nil {
					askedFields[idx] = make(map[string]interface{})
				}
				askedFields[idx][fieldKey] = val.Interface()
			} else {
				field404 = errors.New("unknown field name")
				break o
			}
		}
	}

	if field404 != nil {
		return field404
	}

	return WriteJSON(w, http.StatusOK, askedFields)
}

func goCase(s string) string {
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

func main() {
	s := NewServer(":4000")
	s.Start()
}
