package handler

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"sqlguru/internal/middleware"
	"sqlguru/internal/service"
)

const cookieTTL = time.Hour

type AdminHandler struct {
	svc          *service.AdminService
	templatesDir string
	jwtSecret    []byte
	jwtTTL       time.Duration
}

func NewAdminHandler(svc *service.AdminService, templatesDir string, jwtSecret []byte, jwtTTL time.Duration) *AdminHandler {
	return &AdminHandler{
		svc:          svc,
		templatesDir: templatesDir,
		jwtSecret:    jwtSecret,
		jwtTTL:       jwtTTL,
	}
}

func (h *AdminHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		http.ServeFile(w, r, filepath.Join(h.templatesDir, "admin_login.html"))
		return
	}

	login := r.FormValue("login")
	password := r.FormValue("password")

	if err := h.svc.Authenticate(login, password); err != nil {
		h.renderLogin(w, "Неверный логин или пароль")
		return
	}

	token, err := middleware.GenerateJWT(login, h.jwtSecret, h.jwtTTL)
	if err != nil {
		http.Error(w, "Ошибка создания токена", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:   "adminJWT",
		Value:  token,
		Path:   "/admin",
		MaxAge: int(h.jwtTTL.Seconds()),
	})
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (h *AdminHandler) Panel(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		action := r.FormValue("action")
		table := r.FormValue("table")
		id := r.FormValue("id")

		var err error
		switch action {
		case "delete":
			err = h.svc.DeleteRecord(table, id)
		case "create":
			err = h.svc.CreateRecord(table, h.collectFields(r))
		case "update":
			err = h.svc.UpdateRecord(table, id, h.collectFields(r))
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	table := r.FormValue("table")
	if table == "" {
		table = r.URL.Query().Get("table")
	}
	if table == "" {
		table = "admins"
	}

	tableData, topics, err := h.svc.GetTableData(table)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		TableName string
		Headers   []string
		Rows      []map[string]interface{}
		Topics    interface{}
	}{
		TableName: tableData.TableName,
		Headers:   tableData.Headers,
		Rows:      tableData.Rows,
		Topics:    topics,
	}

	h.render(w, "admin.html", data)
}

func (h *AdminHandler) GetRecord(w http.ResponseWriter, r *http.Request) {
	table := r.URL.Query().Get("table")
	id := r.URL.Query().Get("id")

	data, err := h.svc.GetRecord(table, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (h *AdminHandler) GetTopics(w http.ResponseWriter, r *http.Request) {
	topics, err := h.svc.GetTopicsWithID()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(topics)
}

func (h *AdminHandler) collectFields(r *http.Request) map[string]string {
	return map[string]string{
		"login":          r.FormValue("login"),
		"password":       r.FormValue("password"),
		"name":           r.FormValue("name"),
		"topic_id":       r.FormValue("topic_id"),
		"question_text":  r.FormValue("question_text"),
		"correct_answer": r.FormValue("correct_answer"),
		"wrong_answer1":  r.FormValue("wrong_answer1"),
		"wrong_answer2":  r.FormValue("wrong_answer2"),
		"wrong_answer3":  r.FormValue("wrong_answer3"),
		"topic_name":     r.FormValue("topic_name"),
		"title":          r.FormValue("title"),
		"content":        r.FormValue("content"),
	}
}

func (h *AdminHandler) renderLogin(w http.ResponseWriter, errMsg string) {
	h.render(w, "admin_login.html", map[string]interface{}{"Error": errMsg})
}

func (h *AdminHandler) render(w http.ResponseWriter, tmplName string, data interface{}) {
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
