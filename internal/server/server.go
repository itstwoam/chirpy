package server

import (
	"net/http"
	"fmt"
	//"strconv"
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
)

type newchirp struct {
	Body string `json:"body"`
	ID uuid.UUID `json:"user_id"`
}

type badchirp struct {
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
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	response, err := json.Marshal(allChirps)
	if err != nil {
		log.Printf("Error marshalling chirps: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Write(response)
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
	response, err := json.Marshal(aChirp)
	if err != nil {
		log.Printf("Error marshalling chirp: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(200)
	w.Write(response)
	return
}
func (s *state) serveChirp(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	chirps := newchirp{}
	err := decoder.Decode(&chirps)
	if err != nil {
		type errResponse struct {
			Error string `json:"error"`
		}
		respBody := errResponse{
			Error: "Malformed Post request",
		}
		dat, err := json.Marshal(respBody)
		if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(500)
			return
		}
		//process error here
		w.WriteHeader(422)
		w.Header().Set("Content-Type", "application/json")
		w.Write(dat)
		return
	}
	if len(chirps.Body) > 140{
		type errResponse struct {
			Error string `json:"error"`
		}
		respBody := errResponse {
			Error: "Chirp is too long",
		}
		w.WriteHeader(400)
		dat, err := json.Marshal(respBody)
		if err != nil {
			log.Printf("error marshalling JSON: %s", err)
			w.WriteHeader(500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(dat)
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
	w.WriteHeader(201)
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
		errResponse := badchirp{}
		errResponse.Error = "Unable to create chirp"
		dat, _ := json.Marshal(errResponse)
		w.WriteHeader(401)
		w.Header().Set("Content-Type", "application/json")
		w.Write(dat)
		return
	}
	dat, err := json.Marshal(Chirp)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(dat)
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
	//db, err := sql.Open("postgres", curState.cfg.DB_url)
	db, err := sql.Open("postgres", dbURL)
	curState.db = database.New(db)
	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("./internal/app"))
	//mux.Handle("/app/", http.StripPrefix("/app", fs))
	mux.Handle("/app/", curState.middlewareInc(http.StripPrefix("/app", fs)))
	mux.HandleFunc("POST /api/users", curState.serveUsers)
	mux.HandleFunc("GET /api/healthz", serveStatus)
	mux.HandleFunc("GET /admin/metrics", curState.serveHits)
	mux.HandleFunc("POST /admin/reset", curState.serveReset)
	mux.HandleFunc("POST /api/chirps", curState.serveChirp)
	mux.HandleFunc("GET /api/chirps", curState.getChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", curState.getChirp)
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

func (s *state) serveUsers(w http.ResponseWriter, r *http.Request) {
	type chirp struct {
		Email string `json:"email"`
	}
	decoder := json.NewDecoder(r.Body)
	chirps := chirp{}
	err := decoder.Decode(&chirps)
	if err != nil {
		type errResponse struct {
			Error string `json:"error"`
		}
		respBody := errResponse{
			Error: "Malformed Post request",
		}
		dat, err := json.Marshal(respBody)
		if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(500)
			return
		}
		//process error here
		w.WriteHeader(422)
		w.Header().Set("Content-Type", "application/json")
		w.Write(dat)
		return
	}
	
	response, err:= s.db.CreateUser(r.Context(), database.CreateUserParams{ ID: uuid.New(), CreatedAt: time.Now(), UpdatedAt: time.Now(), Email:chirps.Email,})
	if err != nil {
		fmt.Printf("error registering new user: %v\n", err)
		return
	}
	
	dat, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(201)
	w.Header().Set("Content-Type", "application/json")
	w.Write(dat)
	return
}

func GetUUID(s string) (uuid.UUID, error) {
	theUUID, err := uuid.Parse(s)
	if err != nil {
		return uuid.UUID{}, errors.New("in GetUUID, could not parse UUID")
	}
	return theUUID, nil
}
