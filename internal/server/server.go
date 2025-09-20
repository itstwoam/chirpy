package server

import (
	"net/http"
	"fmt"
	"sync/atomic"
	"encoding/json"
	"log"
	"strings"
	"github.com/itstwoam/chirpy/internal/database"
	"github.com/itstwoam/chirpy/internal/config"
	"database/sql"
	"github.com/google/uuid"
	"time"
	"errors"
	"github.com/itstwoam/chirpy/internal/auth"
)
type CleanUser struct {
	ID uuid.UUID `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email string `json:email"`
}

type UserLoginPost struct {
		Password string `json:"password"`
		Email string `json:"email"`
}

type Newchirp struct {
	Body string `json:"body"`
	ID uuid.UUID `json:"user_id"`
}

type BadChirp struct {
	Error string `json:"error"`
}

type state struct {
	db *database.Queries `json:"db"`
	cfg *config.Config `json:"cfg"`
	fileserverHits atomic.Int32 `json:"fileserverHits"`
	environType string `json:"environtype"`
}
	
var blacklist = []string {"kerfuffle", "sharbert", "fornax"}

func (s *state) middlewareInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		s.fileserverHits.Add(1)	
		next.ServeHTTP(w, r) 
	})
}

func (s *state) serveHits(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	hits := s.fileserverHits.Load()
	w.Write([]byte(fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", hits)))
}

func (s *state) serveReset(w http.ResponseWriter, r *http.Request) {
	if s.environType != "dev" {
		fmt.Println("Environment isn't = \"dev\" so users cannot be reset")
		return
	}
	numDeleted, err := s.db.DeleteAllUsers(r.Context())
	if err != nil {
		fmt.Printf("error when clearing users database %v\nUsers deleted: %v\n", err, numDeleted)
		return
	}
	fmt.Printf("Deleted %v users.", numDeleted)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	s.fileserverHits.Store(0)
}

func (s *state) getChirps(w http.ResponseWriter, r *http.Request) {
	allChirps, err := s.db.GetChirps(r.Context())
	if err != nil {
		fmt.Println("error while retrieving chirps")
		return
	}
	WriteJSONResponse(w, allChirps, 200, -1)
	return
}

func (s *state) getChirp(w http.ResponseWriter, r *http.Request) {
	cID, err := GetUUID(r.PathValue("chirpID"))
	if err != nil {
		log.Printf("Error getting UUID: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}	
	aChirp, err := s.db.GetChirp(r.Context(), cID)
	if err != nil {
		log.Printf("Error in retrieving chirp: %v", err)
		w.WriteHeader(404)
		return
	}
	WriteJSONResponse(w, aChirp, 200, -1)
	return
}

func (s *state) serveChirp(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	chirps := Newchirp{}
	err := decoder.Decode(&chirps)
	if err != nil {
		type errResponse struct {
			Error string `json:"error"`
		}
		respBody := errResponse{
			Error: "Malformed Post request",
		}
		WriteJSONResponse(w, respBody, 422, 500)
		return
	}
	if len(chirps.Body) > 140{
		type errResponse struct {
			Error string `json:"error"`
		}
		respBody := errResponse {
			Error: "Chirp is too long",
		}
		WriteJSONResponse(w, respBody, 400, 500)
		return
	}
	final := ""
	words := strings.Fields(chirps.Body)
	for idx, word := range words {
		for _, bad := range blacklist {
			if strings.EqualFold(word, bad){
				words[idx] = "****"
			}
		}
		if idx > 0 {
			final += " "
		}
		final += words[idx]
	}
	chirps.Body = final
	timenow := time.Now()
	Chirp, err := s.db.CreateChirp(r.Context(), database.CreateChirpParams {
		ID: uuid.New(),
		CreatedAt: timenow,
		UpdatedAt: timenow,
		Body: chirps.Body,
		UserID: chirps.ID,
	})
	if err != nil {
		fmt.Printf("Unable to create chirp: %v", err)
		errResponse := BadChirp{}
		errResponse.Error = "Unable to create chirp"
		WriteJSONResponse(w, errResponse, 401, -1)
		return
	}
	WriteJSONResponse(w, Chirp, 201, 500)
	return
}



func StartServer(dbURL string, devEnv string) {
	var curState state
	curState.environType = devEnv
	var err error
	curState.cfg, err = config.Read()
	if err != nil{
		fmt.Println("could not retrieve config file")
		return
	}
	db, err := sql.Open("postgres", dbURL)
	curState.db = database.New(db)
	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("./internal/app"))
	mux.Handle("/app/", curState.middlewareInc(http.StripPrefix("/app", fs)))
	mux.HandleFunc("POST /api/users", curState.serveUsers)
	mux.HandleFunc("GET /api/healthz", serveStatus)
	mux.HandleFunc("GET /admin/metrics", curState.serveHits)
	mux.HandleFunc("POST /admin/reset", curState.serveReset)
	mux.HandleFunc("POST /api/chirps", curState.serveChirp)
	mux.HandleFunc("GET /api/chirps", curState.getChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", curState.getChirp)
	mux.HandleFunc("POST /api/login", curState.serveLogin)
	server := http.Server{}
	server.Handler = mux
	server.Addr = ":8085"
	err = server.ListenAndServe()
	if err != nil {
		fmt.Println("There was an error, but wtf?")
	}
	fmt.Println("Am I still running?")
}

func serveStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

func (s *state) serveLogin(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	newUser := UserLoginPost{} //non hashed password
	err := decoder.Decode(&newUser)
	if err != nil {
		respBody := BadChirp{
			Error: "Malformed login request",
		}
		WriteJSONResponse(w, respBody, 422, 500)
		return
	}
	user, err := s.db.GetUserByEmail(r.Context(), newUser.Email)// hashed password 
	if err != nil {
		respBody := BadChirp{
			Error: "Couldn't find user",
		}
		WriteJSONResponse(w, respBody, 422, 500)
		return
	}
	isValid := auth.CheckPasswordHash(newUser.Password, user.HashedPassword)
	if isValid != nil {
		respBody := BadChirp{
			Error: "Incorrect email or password",
		}
		WriteJSONResponse(w, respBody, 401, 500)
	} else {
		respBody := CleanUser {
			ID: user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email: user.Email,
		}
		WriteJSONResponse(w, respBody, 200, -1)
	}
}

func (s *state) serveUsers(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	newUser := UserLoginPost{}
	err := decoder.Decode(&newUser)
	if err != nil {
		respBody := BadChirp{
			Error: "Malformed Post request",
		}
		WriteJSONResponse(w, respBody, 422, 500)
		return
	}
	if len(newUser.Password) < 5 {
		respBody := BadChirp{
			Error: "Password length below minimum",
		}
		WriteJSONResponse(w, respBody, 422, 500)
		return
	}
	hashWord, err := auth.HashPassword(newUser.Password)
	if err != nil {
		respBody := BadChirp{
			Error: "Error processing password",
		}
		WriteJSONResponse(w, respBody, 422, 500)
		return
	}
	response, err:= s.db.CreateUser(r.Context(), database.CreateUserParams{ ID: uuid.New(), CreatedAt: time.Now(), UpdatedAt: time.Now(), Email:newUser.Email, HashedPassword: hashWord})
	if err != nil {
		fmt.Printf("error registering new user: %v\n", err)
		return
	}
	
	WriteJSONResponse(w, response, 201, -1)
	if err != nil {
		fmt.Errorf("failed to write reponse: %v", err)
		return
	}
	return
}

func GetUUID(s string) (uuid.UUID, error) {
	theUUID, err := uuid.Parse(s)
	if err != nil {
		return uuid.UUID{}, errors.New("in GetUUID, could not parse UUID")
	}
	return theUUID, nil
}

func WriteJSONResponse(w http.ResponseWriter, t any, code int, ecode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	data, err := json.Marshal(t)
	if err != nil {
		if ecode != -1 {
			w.WriteHeader(ecode)
		}
		fmt.Errorf("error in marshalling JSON: %w", err)
		return
	}
	_, writeErr := w.Write(data)
	if writeErr != nil {
		errMsg := "error writing response: %v"
		log.Printf(errMsg, writeErr)
		if ecode != -1 {
			w.WriteHeader(ecode)
		}
		fmt.Errorf(errMsg, writeErr)
		return
	}
	return
}
