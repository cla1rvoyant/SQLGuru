package handler

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"sqlguru/internal/service"
)

type ArticleHandler struct {
	svc          *service.ArticleService
	templatesDir string
}

func NewArticleHandler(svc *service.ArticleService, templatesDir string) *ArticleHandler {
	return &ArticleHandler{svc: svc, templatesDir: templatesDir}
}

func (h *ArticleHandler) Show(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		id = "1"
	}

	article, err := h.svc.GetArticle(id)
	if err != nil {
		http.Error(w, "Статья не найдена", http.StatusNotFound)
		return
	}

	all, _ := h.svc.GetAllArticles()

	tmpl, err := template.ParseFiles(filepath.Join(h.templatesDir, "article.html"))
	if err != nil {
		log.Printf("template parse error (article.html): %v", err)
		http.Error(w, "Ошибка загрузки страницы", http.StatusInternalServerError)
		return
	}

	data := struct {
		Article  interface{}
		Content  template.HTML
		Articles interface{}
	}{
		Article:  article,
		Content:  template.HTML(article.Content),
		Articles: all,
	}

	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("template execute error (article.html): %v", err)
	}
}
