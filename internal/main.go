package main

import (
	"net/http"

	_ "github.com/lib/pq"
)

func main() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("../web/static"))))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, "../web/templates/index.html")
	})

	http.HandleFunc("/admin", JWTAuthMiddleware(adminHandler))

	http.HandleFunc("/admin/get", JWTAuthMiddleware(adminGetHandler))

	http.HandleFunc("/admin/topics", JWTAuthMiddleware(adminTopicsHandler))

	http.HandleFunc("/choise", choiseHandler)

	http.HandleFunc("/test", testHandler)

	http.HandleFunc("/result", resultHandler)

	http.HandleFunc("/admin/login", admin_loginHandler)

	http.ListenAndServe("localhost:8080", nil)
}
