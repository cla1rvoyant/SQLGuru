package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

// Обработчик для получения данных записи
func adminGetHandler(w http.ResponseWriter, r *http.Request) {
	connStr := "user=postgres password=aboba dbname=tests_db sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	table := r.URL.Query().Get("table")
	id := r.URL.Query().Get("id")

	var data map[string]interface{}

	switch table {
	case "admins":
		var login string
		err := db.QueryRow("SELECT login FROM admins WHERE id = $1", id).Scan(&login)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		data = map[string]interface{}{
			"login": login,
		}

	case "topics":
		var name string
		err := db.QueryRow("SELECT name FROM topics WHERE id = $1", id).Scan(&name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		data = map[string]interface{}{
			"name": name,
		}

	case "questions":
		var topicID int
		var questionText, correctAnswer, wrong1, wrong2, wrong3 string

		err := db.QueryRow(`
            SELECT topic_id, question_text, correct_answer, 
                   wrong_answer1, 
                   COALESCE(wrong_answer2, ''),
                   COALESCE(wrong_answer3, '')
            FROM questions WHERE id = $1
        `, id).Scan(&topicID, &questionText, &correctAnswer, &wrong1, &wrong2, &wrong3)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		data = map[string]interface{}{
			"topic_id":       topicID,
			"question_text":  questionText,
			"correct_answer": correctAnswer,
			"wrong_answer1":  wrong1,
			"wrong_answer2":  wrong2,
			"wrong_answer3":  wrong3,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// Обработчик для получения списка тем
func adminTopicsHandler(w http.ResponseWriter, r *http.Request) {
	connStr := "user=postgres password=aboba dbname=tests_db sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, name FROM topics ORDER BY name")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var topics []map[string]interface{}
	for rows.Next() {
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		topics = append(topics, map[string]interface{}{
			"id":   id,
			"name": name,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(topics)
}

func admin_loginHandler(w http.ResponseWriter, r *http.Request) {
	connStr := "user=postgres password=aboba dbname=tests_db sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
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
			tmpl := template.Must(template.ParseFiles("../web/templates/admin_login.html"))
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
			tmpl := template.Must(template.ParseFiles("../web/templates/admin_login.html"))
			tmpl.Execute(w, map[string]interface{}{"Error": "Неверный логин или пароль"})
			return
		}
	}

	http.ServeFile(w, r, "../web/templates/admin_login.html")
}

func adminHandler(w http.ResponseWriter, r *http.Request) {
	connStr := "user=postgres password=aboba dbname=tests_db sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Обработка POST запросов
	if r.Method == "POST" {
		action := r.FormValue("action")
		table := r.FormValue("table")

		switch action {
		case "delete":
			id := r.FormValue("id")
			switch table {
			case "admins":
				var adminLogin string
				err := db.QueryRow("SELECT login FROM admins WHERE id = $1", id).Scan(&adminLogin)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				// Проверка на главного администратора
				if adminLogin == "main" {
					http.Error(w, "Невозможно удалить главного администратора", http.StatusForbidden)
					return
				}

				tx, err := db.Begin()
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				defer tx.Rollback()

				_, err = db.Exec("DELETE FROM admins WHERE id = $1", id)

				_, err = tx.Exec("SELECT setval('admins_id_seq', COALESCE((SELECT MAX(id) FROM admins), 0) + 1, false)")
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			case "topics":
				tx, err := db.Begin()
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				defer tx.Rollback()

				_, err = db.Exec("DELETE FROM topics WHERE id = $1", id)

				_, err = tx.Exec("SELECT setval('topics_id_seq', COALESCE((SELECT MAX(id) FROM topics), 0) + 1, false)")
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			case "questions":
				tx, err := db.Begin()
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				defer tx.Rollback()

				_, err = db.Exec("DELETE FROM questions WHERE id = $1", id)

				_, err = tx.Exec("SELECT setval('questions_id_seq', COALESCE((SELECT MAX(id) FROM questions), 0) + 1, false)")
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
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
		case "update":
			switch table {
			case "admins":
				id := r.FormValue("id")
				login := r.FormValue("login")
				password := r.FormValue("password")

				if password != "" {
					hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
					_, err = db.Exec("UPDATE admins SET login = $1, password = $2 WHERE id = $3",
						login, string(hashedPassword), id)
				} else {
					_, err = db.Exec("UPDATE admins SET login = $1 WHERE id = $2", login, id)
				}

			case "topics":
				id := r.FormValue("id")
				name := r.FormValue("name")
				_, err = db.Exec("UPDATE topics SET name = $1 WHERE id = $2", name, id)

			case "questions":
				id := r.FormValue("id")
				topicID := r.FormValue("topic_id")
				questionText := r.FormValue("question_text")
				correctAnswer := r.FormValue("correct_answer")
				wrongAnswer1 := r.FormValue("wrong_answer1")
				wrongAnswer2 := r.FormValue("wrong_answer2")
				wrongAnswer3 := r.FormValue("wrong_answer3")

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

				_, err = db.Exec(`
                    UPDATE questions 
                    SET topic_id = $1, question_text = $2, correct_answer = $3, 
                        wrong_answer1 = $4, wrong_answer2 = $5, wrong_answer3 = $6
                    WHERE id = $7
                `, topicID, questionText, correctAnswer, wrongAnswer1, wrong2, wrong3, id)
			}

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Определение активную таблицу
	table := r.FormValue("table")
	if table == "" {
		table = "admins"
	}

	var tableData TableData
	tableData.TableName = table

	// Список тем для формы создания или редактирования вопроса
	var topics []Topic
	if table == "questions" {
		topicRows, err := db.Query("SELECT id, name FROM topics")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer topicRows.Close()

		for topicRows.Next() {
			var topic Topic
			if err := topicRows.Scan(&topic.ID, &topic.Name); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			topics = append(topics, topic)
		}
	}

	// Загрузка данных
	switch table {
	case "admins":
		tableData.Headers = []string{"ID", "Логин"}
		rows, err := db.Query("SELECT id, login FROM admins ORDER BY id")
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
		rows, err := db.Query("SELECT id, name FROM topics ORDER BY id")
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
			ORDER BY q.id
        `)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var id int
			var topicName, questionText, correctAnswer, wrong1, wrong2, wrong3 string
			if err := rows.Scan(&id, &topicName, &questionText, &correctAnswer, &wrong1, &wrong2, &wrong3); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			row := map[string]interface{}{
				"ID":               id,
				"Тема":             topicName,
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
		Topics []Topic
	}{
		TableData: tableData,
		Topics:    topics,
	}

	tmpl, err := template.ParseFiles("../web/templates/admin.html")
	if err != nil {
		http.Error(w, "Ошибка загрузки шаблона: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Ошибка выполнения шаблона: "+err.Error(), http.StatusInternalServerError)
		return
	}
}
