package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"

	_ "github.com/lib/pq"
)

type Question struct {
	ID              uint
	Text            string
	CorrectAnswer   string
	WrongAnswer1    string
	WrongAnswer2    string
	WrongAnswer3    string
	ShuffledAnswers []string
}

// func getNextTestID(r *http.Request) string {
// 	maxID := 0
// 	for _, cookie := range r.Cookies() {
// 		if strings.HasPrefix(cookie.Name, "test_") {
// 			parts := strings.Split(cookie.Name, "_")
// 			if len(parts) == 2 {
// 				id, err := strconv.Atoi(parts[1])
// 				if err == nil {
// 					if id > maxID {
// 						maxID = id
// 					}
// 				}
// 			}
// 		}
// 	}

// 	return strconv.Itoa(maxID + 1)
// }

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

func main() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, "templates/index.html")
	})

	http.HandleFunc("/choise", choiseHandler)

	http.HandleFunc("/test", testHandler)

	http.HandleFunc("/result", resultHandler)

	http.ListenAndServe("localhost:8080", nil)
}
