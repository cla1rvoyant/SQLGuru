package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	_ "github.com/lib/pq"
)

var JWTSecretString = []byte("ajsodnjasndojasd")

type TableData struct {
	TableName string
	Headers   []string
	Rows      []map[string]interface{}
	Actions   bool
}

// Для получения списка тем при создании вопроса
type Theme struct {
	ID   int
	Name string
}

type Question struct {
	ID              uint
	Text            string
	CorrectAnswer   string
	WrongAnswer1    string
	WrongAnswer2    string
	WrongAnswer3    string
	ShuffledAnswers []string
}

func testHandler(w http.ResponseWriter, r *http.Request) {
	topic := r.URL.Query().Get("topic")
	fmt.Printf("DEBUG: Topic received: %s\n", topic)

	connStr := "user=postgres password=aboba dbname=tests_db sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Printf("DEBUG: DB connection error: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	if r.Method == "POST" {
		selectedAnswer := r.FormValue("answer")
		questionID := r.FormValue("question_id")

		var correctAnswer string

		err = db.QueryRow(`
		SELECT correct_answer
		FROM questions
		WHERE id = $1
		`, questionID).Scan(&correctAnswer)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		correctAnswerCounter := 0
		cookie, err := r.Cookie("correctAnswerCounter")
		if err == nil {
			correctAnswerCounter, _ = strconv.Atoi(cookie.Value)
		}

		fmt.Println(selectedAnswer)

		if selectedAnswer == correctAnswer {
			fmt.Println("Правильный ответ")
			correctAnswerCounter++
		} else {
			fmt.Println("Неправильный ответ")
		}

		http.SetCookie(w, &http.Cookie{
			Name:   "correctAnswerCounter",
			Value:  strconv.Itoa(correctAnswerCounter),
			Path:   "/",
			MaxAge: 3600,
		})

		var nextQuestion Question
		err = db.QueryRow(`
		SELECT id, question_text, correct_answer, wrong_answer1, wrong_answer2, wrong_answer3
		FROM questions
		WHERE topic_id = (SELECT topic_id FROM questions WHERE id = $1)
		AND id > $1
		ORDER BY id LIMIT 1
		`, questionID).Scan(&nextQuestion.ID, &nextQuestion.Text, &nextQuestion.CorrectAnswer, &nextQuestion.WrongAnswer1, &nextQuestion.WrongAnswer2, &nextQuestion.WrongAnswer3)

		if err == sql.ErrNoRows {
			http.Redirect(w, r, "/result?topic="+url.QueryEscape(topic), http.StatusSeeOther)
			return
		}

		fmt.Println(nextQuestion.Text)

		answers := []string{nextQuestion.CorrectAnswer, nextQuestion.WrongAnswer1, nextQuestion.WrongAnswer2, nextQuestion.WrongAnswer3}
		for i := 2; i < len(answers); i++ {
			if answers[i] == "" {
				answers[i] = answers[len(answers)-1]
				answers = answers[:len(answers)-1]
			}
		}
		rand.Shuffle(len(answers), func(i, j int) { answers[i], answers[j] = answers[j], answers[i] })
		nextQuestion.ShuffledAnswers = answers

		tmpl := template.Must(template.ParseFiles("templates/exercise.html"))
		tmpl.Execute(w, map[string]interface{}{
			"Topic":    topic,
			"Question": nextQuestion,
		})
	}

	if r.Method == "GET" {
		http.SetCookie(w, &http.Cookie{
			Name:   "correctAnswerCounter",
			Value:  "0",
			Path:   "/",
			MaxAge: 3600,
		})

		var firstQuestion Question
		err = db.QueryRow(`
		SELECT id, question_text, correct_answer, wrong_answer1, wrong_answer2, wrong_answer3
		FROM questions 
		WHERE topic_id = (SELECT id FROM topics WHERE name = $1)
		ORDER BY id LIMIT 1
		`, topic).Scan(&firstQuestion.ID, &firstQuestion.Text, &firstQuestion.CorrectAnswer, &firstQuestion.WrongAnswer1, &firstQuestion.WrongAnswer2, &firstQuestion.WrongAnswer3)
		if err != nil {
			http.Error(w, "Questions not found", http.StatusNotFound)
			return
		}
		fmt.Printf("Текст вопроса: %s", firstQuestion.Text)

		answers := []string{firstQuestion.CorrectAnswer, firstQuestion.WrongAnswer1, firstQuestion.WrongAnswer2, firstQuestion.WrongAnswer3}
		for i := 2; i < len(answers); i++ {
			if answers[i] == "" {
				answers[i] = answers[len(answers)-1]
				answers = answers[:len(answers)-1]
			}
		}
		rand.Shuffle(len(answers), func(i, j int) { answers[i], answers[j] = answers[j], answers[i] })
		firstQuestion.ShuffledAnswers = answers

		tmpl := template.Must(template.ParseFiles("templates/exercise.html"))
		tmpl.Execute(w, map[string]interface{}{
			"Topic":    topic,
			"Question": firstQuestion,
		})
	}
}

func resultHandler(w http.ResponseWriter, r *http.Request) {
	topic := r.URL.Query().Get("topic")

	correctAnswerCounter := 0

	cookie, err := r.Cookie("correctAnswerCounter")
	if err == nil {
		correctAnswerCounter, _ = strconv.Atoi(cookie.Value)
	}

	tmpl := template.Must(template.ParseFiles("templates/result.html"))
	tmpl.Execute(w, map[string]interface{}{
		"Topic":                topic,
		"correctAnswerCounter": correctAnswerCounter,
	})
}

func choiseHandler(w http.ResponseWriter, r *http.Request) {
	connStr := "user=postgres password=aboba dbname=tests_db sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Printf("DEBUG: DB connection error: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var topics []string

	rows, err := db.Query(`SELECT name FROM topics`)
	if err != nil {
		http.Error(w, "Topics not found", http.StatusNotFound)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var topic string
		rows.Scan(&topic)
		topics = append(topics, topic)
	}

	tmpl := template.Must(template.ParseFiles("templates/choise.html"))
	tmpl.Execute(w, map[string]interface{}{
		"Topics": topics,
	})
}

func generateJWT(adminLogin string) (string, error) {
	claims := jwt.MapClaims{
		"adm": adminLogin,
		"exp": time.Now().Add(24 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(JWTSecretString)
}

func admin_loginHandler(w http.ResponseWriter, r *http.Request) {
	connStr := "user=postgres password=aboba dbname=tests_db sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Printf("DEBUG: DB connection error: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	if r.Method == "POST" {
		adminLogin := r.FormValue("login")
		adminPassword := r.FormValue("password")

		var searchPassword string
		err := db.QueryRow(`SELECT password FROM admins WHERE login = $1`, adminLogin).Scan(&searchPassword)
		if err != nil {
			fmt.Println("Admin not found")
			tmpl := template.Must(template.ParseFiles("templates/admin_login.html"))
			tmpl.Execute(w, map[string]interface{}{"Error": "Неверный логин или пароль"})
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(searchPassword), []byte(adminPassword))
		if err == nil {
			fmt.Println("Success")
			token, err := generateJWT(adminLogin)
			if err != nil {
				http.Error(w, "Ошибка создания токена", http.StatusInternalServerError)
				return
			}

			http.SetCookie(w, &http.Cookie{
				Name:   "adminJWT",
				Value:  token,
				Path:   "/admin",
				MaxAge: 86400,
			})
			http.Redirect(w, r, "/admin", http.StatusSeeOther)
			return
		} else {
			fmt.Println("Wrong")
			tmpl := template.Must(template.ParseFiles("templates/admin_login.html"))
			tmpl.Execute(w, map[string]interface{}{"Error": "Неверный логин или пароль"})
			return
		}
	}

	http.ServeFile(w, r, "templates/admin_login.html")
}

func JWTAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("adminJWT")
		if err != nil {
			http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
			return
		}

		token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("неожиданный метод подписи: %v", token.Header["alg"])
			}
			return JWTSecretString, nil
		})

		if err != nil || !token.Valid {
			http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
			return
		}

		next(w, r)
	}
}

func adminHandler(w http.ResponseWriter, r *http.Request) {
	connStr := "user=postgres password=aboba dbname=tests_db sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Обрабатываем POST запросы
	if r.Method == "POST" {
		action := r.FormValue("action")
		table := r.FormValue("table")

		switch action {
		case "delete":
			id := r.FormValue("id")
			switch table {
			case "admins":
				_, err = db.Exec("DELETE FROM admins WHERE id = $1", id)
			case "topics":
				_, err = db.Exec("DELETE FROM topics WHERE id = $1", id)
			case "questions":
				_, err = db.Exec("DELETE FROM questions WHERE id = $1", id)
			}

		case "create":
			switch table {
			case "admins":
				login := r.FormValue("login")
				password := r.FormValue("password")
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
				_, err = db.Exec("INSERT INTO admins (login, password) VALUES ($1, $2)", login, string(hashedPassword))

			case "topics":
				name := r.FormValue("name")
				_, err = db.Exec("INSERT INTO topics (name) VALUES ($1)", name)

			case "questions":
				topicID := r.FormValue("topic_id")
				questionText := r.FormValue("question_text")
				correctAnswer := r.FormValue("correct_answer")
				wrongAnswer1 := r.FormValue("wrong_answer1")
				wrongAnswer2 := r.FormValue("wrong_answer2")
				wrongAnswer3 := r.FormValue("wrong_answer3")
				fmt.Printf("%T", topicID)
				// Для необязательных полей используем NULL если пусто
				var wrong2, wrong3 interface{}
				if wrongAnswer2 == "" {
					wrong2 = nil
				} else {
					wrong2 = wrongAnswer2
				}
				if wrongAnswer3 == "" {
					wrong3 = nil
				} else {
					wrong3 = wrongAnswer3
				}

				_, err = db.Exec("INSERT INTO questions (id, topic_id, question_text, correct_answer, wrong_answer1, wrong_answer2, wrong_answer3) VALUES (DEFAULT, $1, $2, $3, $4, $5, $6)",
					topicID, questionText, correctAnswer, wrongAnswer1, wrong2, wrong3)
			}

		case "switch":
			// Просто переключаем таблицу
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Определяем активную таблицу
	table := r.FormValue("table")
	if table == "" {
		table = "admins"
	}

	var tableData TableData
	tableData.TableName = table
	tableData.Actions = true

	// Для формы создания вопроса нужен список тем
	var topics []Theme
	if table == "questions" {
		themeRows, err := db.Query("SELECT id, name FROM topics")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer themeRows.Close()

		for themeRows.Next() {
			var theme Theme
			if err := themeRows.Scan(&theme.ID, &theme.Name); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			topics = append(topics, theme)
		}
	}

	// Загружаем данные в универсальном формате
	switch table {
	case "admins":
		tableData.Headers = []string{"ID", "Логин"}
		rows, err := db.Query("SELECT id, login FROM admins")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var id int
			var login string
			if err := rows.Scan(&id, &login); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			row := map[string]interface{}{
				"ID":    id,
				"Логин": login,
			}
			tableData.Rows = append(tableData.Rows, row)
		}

	case "topics":
		tableData.Headers = []string{"ID", "Название"}
		rows, err := db.Query("SELECT id, name FROM topics")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var id int
			var name string
			if err := rows.Scan(&id, &name); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			row := map[string]interface{}{
				"ID":       id,
				"Название": name,
			}
			tableData.Rows = append(tableData.Rows, row)
		}

	case "questions":
		tableData.Headers = []string{"ID", "Тема", "Вопрос", "Правильный ответ", "Неправильный 1", "Неправильный 2", "Неправильный 3"}
		rows, err := db.Query(`
            SELECT q.id, t.name, q.question_text, q.correct_answer, 
                COALESCE(q.wrong_answer1, ''), 
                COALESCE(q.wrong_answer2, ''), 
                COALESCE(q.wrong_answer3, '')
            FROM questions q 
            JOIN topics t ON q.topic_id = t.id
        `)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var id int
			var themeName, questionText, correctAnswer, wrong1, wrong2, wrong3 string
			if err := rows.Scan(&id, &themeName, &questionText, &correctAnswer, &wrong1, &wrong2, &wrong3); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			row := map[string]interface{}{
				"ID":               id,
				"Тема":             themeName,
				"Вопрос":           questionText,
				"Правильный ответ": correctAnswer,
				"Неправильный 1":   wrong1,
				"Неправильный 2":   wrong2,
				"Неправильный 3":   wrong3,
			}
			tableData.Rows = append(tableData.Rows, row)
		}
	}

	// Подготовка данных для шаблона
	data := struct {
		TableData
		Topics []Theme
	}{
		TableData: tableData,
		Topics:    topics,
	}

	// Отладочный вывод
	fmt.Printf("DEBUG: TableName: %s\n", tableData.TableName)
	fmt.Printf("DEBUG: Headers: %v\n", tableData.Headers)
	fmt.Printf("DEBUG: Rows count: %d\n", len(tableData.Rows))
	for i, row := range tableData.Rows {
		fmt.Printf("DEBUG: Row %d: %v\n", i, row)
	}

	tmpl, err := template.ParseFiles("templates/admin.html")
	if err != nil {
		http.Error(w, "Ошибка загрузки шаблона: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Выполнение шаблона с обработкой ошибок
	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Ошибка выполнения шаблона: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func main() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, "templates/index.html")
	})

	http.HandleFunc("/admin", JWTAuthMiddleware(adminHandler))

	http.HandleFunc("/choise", choiseHandler)

	http.HandleFunc("/test", testHandler)

	http.HandleFunc("/result", resultHandler)

	http.HandleFunc("/admin/login", admin_loginHandler)

	http.ListenAndServe("localhost:8080", nil)
}
