package server

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"mime"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"github.com/google/uuid"
	"github.com/zfusx/mbti/internal/models"
	"github.com/zfusx/mbti/internal/scoring"
	"github.com/zfusx/mbti/internal/storage"
)

// Server wires HTTP routing to domain services.
type Server struct {
	store         *storage.Storage
	scorer        *scoring.Service
	allowedOrigin string
	questionCount int
	spaFS         http.FileSystem
	router        chi.Router
}

// New constructs a Server.
func New(store *storage.Storage, scorer *scoring.Service, allowedOrigin string, questionCount int, spaFS http.FileSystem) *Server {
	s := &Server{
		store:         store,
		scorer:        scorer,
		allowedOrigin: allowedOrigin,
		questionCount: questionCount,
		spaFS:         spaFS,
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(s.cors)
	r.Use(httprate.LimitByIP(300, time.Minute))

	r.Get("/healthz", s.handleHealth)
	r.Route("/api", func(api chi.Router) {
		api.Get("/questions", s.handleGetQuestions)
		api.With(httprate.LimitByIP(30, time.Minute)).Post("/sessions", s.handleCreateSession)
		api.With(httprate.LimitByIP(30, time.Minute)).Post("/sessions/{id}/answers", s.handleSubmitAnswers)
		api.Get("/results/{type}", s.handleGetResult)
	})

	if s.spaFS != nil {
		r.NotFound(s.spaHandler())
	}

	s.router = r
	return s
}

// Handler exposes the HTTP handler for the API.
func (s *Server) Handler() http.Handler {
	return s.router
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := s.store.Ping(ctx); err != nil {
		writeInternalError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleGetQuestions(w http.ResponseWriter, r *http.Request) {
	limit := s.questionCount
	if v := r.URL.Query().Get("limit"); v != "" {
		parsed, err := strconv.Atoi(v)
		if err != nil || !s.supportedQuizSize(parsed) {
			writeError(w, http.StatusBadRequest, errors.New("limit must be 20, 40, or the configured full question count"))
			return
		}
		limit = parsed
	}

	randomize := strings.EqualFold(r.URL.Query().Get("random"), "true")

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	items, err := s.store.GetQuestions(ctx, limit, randomize)
	if err != nil {
		writeInternalError(w, err)
		return
	}
	if len(items) != limit {
		writeInternalError(w, errors.New("question bank does not contain the requested balanced quiz"))
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"count":     len(items),
		"questions": items,
	})
}

func (s *Server) handleCreateSession(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	id, err := s.store.CreateSession(ctx)
	if err != nil {
		writeInternalError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{
		"sessionId": id.String(),
	})
}

func (s *Server) handleSubmitAnswers(w http.ResponseWriter, r *http.Request) {
	sessionIDStr := chi.URLParam(r, "id")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, errors.New("invalid session id"))
		return
	}

	var payload struct {
		Answers []models.Answer `json:"answers"`
	}

	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, errors.New("invalid JSON payload"))
		return
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		writeError(w, http.StatusBadRequest, errors.New("request body must contain one JSON object"))
		return
	}

	if len(payload.Answers) == 0 {
		writeError(w, http.StatusBadRequest, errors.New("answers array cannot be empty"))
		return
	}
	if !s.supportedQuizSize(len(payload.Answers)) {
		writeError(w, http.StatusBadRequest, errors.New("answer count must match a supported quiz size"))
		return
	}

	ids := make([]int32, 0, len(payload.Answers))
	seen := make(map[int]struct{}, len(payload.Answers))
	for _, answer := range payload.Answers {
		if answer.QuestionID <= 0 {
			writeError(w, http.StatusBadRequest, errors.New("question ids must be positive"))
			return
		}
		if _, duplicate := seen[answer.QuestionID]; duplicate {
			writeError(w, http.StatusBadRequest, errors.New("duplicate question id"))
			return
		}
		seen[answer.QuestionID] = struct{}{}
		ids = append(ids, int32(answer.QuestionID))
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	questions, err := s.store.GetQuestionsByIDs(ctx, ids)
	if err != nil {
		writeInternalError(w, err)
		return
	}
	if err := validateAnswers(payload.Answers, questions); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	result, err := s.scorer.Score(payload.Answers, len(payload.Answers))
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if err := s.store.SaveSessionAnswers(ctx, sessionID, payload.Answers, result); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			writeError(w, http.StatusNotFound, errors.New("session not found"))
			return
		}
		writeInternalError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"sessionId": sessionID.String(),
		"result":    result,
	})
}

func (s *Server) handleGetResult(w http.ResponseWriter, r *http.Request) {
	typ := strings.ToUpper(chi.URLParam(r, "type"))
	if !validMBTIType(typ) {
		writeError(w, http.StatusBadRequest, errors.New("invalid MBTI type"))
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	doc, err := s.store.GetResult(ctx, typ)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			writeError(w, http.StatusNotFound, errors.New("result not found"))
			return
		}
		writeInternalError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": doc,
	})
}

func (s *Server) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.allowedOrigin != "" {
			origin := r.Header.Get("Origin")
			if s.allowedOrigin == "*" || origin == s.allowedOrigin {
				w.Header().Set("Access-Control-Allow-Origin", s.allowedOrigin)
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
				w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
				w.Header().Add("Vary", "Origin")
			}
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{
		"error": err.Error(),
	})
}

func writeInternalError(w http.ResponseWriter, err error) {
	log.Printf("internal request error: %v", err)
	writeError(w, http.StatusInternalServerError, errors.New("internal server error"))
}

func (s *Server) supportedQuizSize(size int) bool {
	return size > 0 && size <= s.questionCount && (size == 20 || size == 40 || size == s.questionCount)
}

func validateAnswers(answers []models.Answer, questions []models.Question) error {
	if len(questions) != len(answers) {
		return errors.New("one or more question ids are invalid")
	}

	byID := make(map[int]models.Question, len(questions))
	for _, question := range questions {
		byID[question.ID] = question
	}

	dimensionCounts := map[string]int{"EI": 0, "SN": 0, "TF": 0, "JP": 0}
	for _, answer := range answers {
		question, ok := byID[answer.QuestionID]
		if !ok {
			return errors.New("one or more question ids are invalid")
		}
		if _, ok := dimensionCounts[question.Dimension]; !ok {
			return errors.New("question has an invalid MBTI dimension")
		}
		validOption := false
		for _, option := range question.Options {
			if answer.Value == option.Value {
				validOption = true
				break
			}
		}
		if !validOption {
			return errors.New("answer value is not valid for its question")
		}
		dimensionCounts[question.Dimension]++
	}

	expectedPerDimension := len(answers) / len(dimensionCounts)
	for _, count := range dimensionCounts {
		if count != expectedPerDimension {
			return errors.New("answers must contain a balanced set of MBTI dimensions")
		}
	}
	return nil
}

func validMBTIType(value string) bool {
	return len(value) == 4 &&
		strings.ContainsRune("EI", rune(value[0])) &&
		strings.ContainsRune("SN", rune(value[1])) &&
		strings.ContainsRune("TF", rune(value[2])) &&
		strings.ContainsRune("JP", rune(value[3]))
}

func (s *Server) spaHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		target := r.URL.Path
		if target == "/" || target == "" {
			target = "/index.html"
		}

		if !s.assetExists(target) {
			target = "/index.html"
		}

		s.serveFile(w, r, target)
	}
}

func (s *Server) assetExists(p string) bool {
	if s.spaFS == nil {
		return false
	}
	clean := strings.TrimPrefix(path.Clean(p), "/")
	f, err := s.spaFS.Open(clean)
	if err != nil {
		return false
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func (s *Server) serveFile(w http.ResponseWriter, r *http.Request, p string) {
	if s.spaFS == nil {
		http.NotFound(w, r)
		return
	}
	clean := strings.TrimPrefix(path.Clean(p), "/")
	f, err := s.spaFS.Open(clean)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || info.IsDir() {
		http.NotFound(w, r)
		return
	}
	data, err := io.ReadAll(f)
	if err != nil {
		http.Error(w, "failed to read asset", http.StatusInternalServerError)
		return
	}
	ctype := mime.TypeByExtension(path.Ext(info.Name()))
	if ctype == "" {
		ctype = http.DetectContentType(data)
	}
	w.Header().Set("Content-Type", ctype)
	if r.Method == http.MethodHead {
		return
	}
	_, _ = w.Write(data)
}
