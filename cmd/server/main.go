package main

import (
	"database/sql"
	"log"
	"net/http"
	"path/filepath"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"sqlguru/internal/config"
	"sqlguru/internal/handler"
	"sqlguru/internal/middleware"
	"sqlguru/internal/repository"
	"sqlguru/internal/service"
)

func main() {
	_ = godotenv.Load()

	cfg := config.Load()

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("cannot open database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("cannot connect to database: %v", err)
	}

	questionRepo := repository.NewQuestionRepository(db)
	topicRepo := repository.NewTopicRepository(db)
	adminRepo := repository.NewAdminRepository(db)
	articleRepo := repository.NewArticleRepository(db)

	quizSvc := service.NewQuizService(questionRepo, topicRepo)
	adminSvc := service.NewAdminService(adminRepo, topicRepo, questionRepo)
	articleSvc := service.NewArticleService(articleRepo)

	quizH := handler.NewQuizHandler(quizSvc, cfg.TemplatesDir)
	adminH := handler.NewAdminHandler(adminSvc, cfg.TemplatesDir, cfg.JWTSecret, cfg.JWTTokenTTL)
	articleH := handler.NewArticleHandler(articleSvc, cfg.TemplatesDir)

	authMw := middleware.NewJWTMiddleware(cfg.JWTSecret)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(cfg.StaticDir))))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, filepath.Join(cfg.TemplatesDir, "index.html"))
	})

	http.HandleFunc("/admin", authMw(adminH.Panel))
	http.HandleFunc("/admin/get", authMw(adminH.GetRecord))
	http.HandleFunc("/admin/topics", authMw(adminH.GetTopics))
	http.HandleFunc("/admin/login", adminH.Login)
	http.HandleFunc("/article", articleH.Show)
	http.HandleFunc("/choise", quizH.Choice)
	http.HandleFunc("/test", quizH.Test)
	http.HandleFunc("/result", quizH.Result)

	log.Printf("Server starting on %s", cfg.Addr)
	if err := http.ListenAndServe(cfg.Addr, nil); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
