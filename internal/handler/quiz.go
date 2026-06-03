package handler

import (
	"errors"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"

	"sqlguru/internal/service"
)

type QuizHandler struct {
	svc          *service.QuizService
	templatesDir string
}

func NewQuizHandler(svc *service.QuizService, templatesDir string) *QuizHandler {
	return &QuizHandler{svc: svc, templatesDir: templatesDir}
}

func (h *QuizHandler) Choice(w http.ResponseWriter, r *http.Request) {
	topics, err := h.svc.GetTopics()
	if err != nil {
		http.Error(w, "Темы не найдены", http.StatusInternalServerError)
		return
	}

	names := make([]string, len(topics))
	for i, t := range topics {
		names[i] = t.Name
	}

	h.render(w, "choise.html", map[string]interface{}{"Topics": names})
}

func (h *QuizHandler) Test(w http.ResponseWriter, r *http.Request) {
	topic := r.URL.Query().Get("topic")
	if r.Method == http.MethodGet {
		h.startQuiz(w, r, topic)
		return
	}
	if r.Method == http.MethodPost {
		h.processAnswer(w, r, topic)
	}
}

func (h *QuizHandler) Result(w http.ResponseWriter, r *http.Request) {
	topic := r.URL.Query().Get("topic")
	counter := 0
	if cookie, err := r.Cookie("correctAnswerCounter"); err == nil {
		counter, _ = strconv.Atoi(cookie.Value)
	}
	h.render(w, "result.html", map[string]interface{}{
		"Topic":                topic,
		"correctAnswerCounter": counter,
	})
}

func (h *QuizHandler) startQuiz(w http.ResponseWriter, r *http.Request, topic string) {
	http.SetCookie(w, &http.Cookie{
		Name:   "correctAnswerCounter",
		Value:  "0",
		Path:   "/",
		MaxAge: int(cookieTTL.Seconds()),
	})
	q, err := h.svc.StartQuiz(topic)
	if err != nil {
		http.Error(w, "Вопросы не найдены", http.StatusNotFound)
		return
	}
	h.render(w, "exercise.html", map[string]interface{}{"Topic": topic, "Question": q})
}

func (h *QuizHandler) processAnswer(w http.ResponseWriter, r *http.Request, topic string) {
	selectedAnswer := r.FormValue("answer")
	questionID := r.FormValue("question_id")

	if selectedAnswer == "" {
		q, err := h.svc.GetQuestion(questionID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		h.render(w, "exercise.html", map[string]interface{}{
			"Topic":    topic,
			"Question": q,
		})
		return
	}

	isCorrect, next, err := h.svc.CheckAndNext(questionID, selectedAnswer)
	if err != nil && !errors.Is(err, service.ErrNoMoreQuestions) {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	counter := 0
	if cookie, cerr := r.Cookie("correctAnswerCounter"); cerr == nil {
		counter, _ = strconv.Atoi(cookie.Value)
	}
	if isCorrect {
		counter++
	}

	http.SetCookie(w, &http.Cookie{
		Name:   "correctAnswerCounter",
		Value:  strconv.Itoa(counter),
		Path:   "/",
		MaxAge: int(cookieTTL.Seconds()),
	})

	if errors.Is(err, service.ErrNoMoreQuestions) {
		http.Redirect(w, r, "/result?topic="+url.QueryEscape(topic), http.StatusSeeOther)
		return
	}
	h.render(w, "exercise.html", map[string]interface{}{"Topic": topic, "Question": next})
}

func (h *QuizHandler) render(w http.ResponseWriter, tmplName string, data interface{}) {
	tmpl, err := template.ParseFiles(filepath.Join(h.templatesDir, tmplName))
	if err != nil {
		log.Printf("template parse error (%s): %v", tmplName, err)
		http.Error(w, "Ошибка загрузки страницы", http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("template execute error (%s): %v", tmplName, err)
	}
}
