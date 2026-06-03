package repository

import (
	"database/sql"

	"sqlguru/internal/domain"
)

type postgresQuestionRepository struct {
	db *sql.DB
}

func NewQuestionRepository(db *sql.DB) QuestionRepository {
	return &postgresQuestionRepository{db: db}
}

func (r *postgresQuestionRepository) GetFirstByTopic(topicName string) (*domain.Question, error) {
	q := &domain.Question{}
	err := r.db.QueryRow(`
		SELECT id, topic_id, question_text, correct_answer, wrong_answer1,
		       COALESCE(wrong_answer2, ''), COALESCE(wrong_answer3, '')
		FROM questions
		WHERE topic_id = (SELECT id FROM topics WHERE name = $1)
		ORDER BY id LIMIT 1
	`, topicName).Scan(&q.ID, &q.TopicID, &q.Text, &q.CorrectAnswer,
		&q.WrongAnswer1, &q.WrongAnswer2, &q.WrongAnswer3)
	if err != nil {
		return nil, err
	}
	return q, nil
}

func (r *postgresQuestionRepository) GetNextAfter(questionID string) (*domain.Question, error) {
	q := &domain.Question{}
	err := r.db.QueryRow(`
		SELECT id, topic_id, question_text, correct_answer, wrong_answer1,
		       COALESCE(wrong_answer2, ''), COALESCE(wrong_answer3, '')
		FROM questions
		WHERE topic_id = (SELECT topic_id FROM questions WHERE id = $1)
		  AND id > $1
		ORDER BY id LIMIT 1
	`, questionID).Scan(&q.ID, &q.TopicID, &q.Text, &q.CorrectAnswer,
		&q.WrongAnswer1, &q.WrongAnswer2, &q.WrongAnswer3)
	if err != nil {
		return nil, err
	}
	return q, nil
}

func (r *postgresQuestionRepository) GetByID(id string) (*domain.Question, error) {
	q := &domain.Question{}
	err := r.db.QueryRow(`
		SELECT id, topic_id, question_text, correct_answer, wrong_answer1,
		       COALESCE(wrong_answer2, ''), COALESCE(wrong_answer3, '')
		FROM questions WHERE id = $1
	`, id).Scan(&q.ID, &q.TopicID, &q.Text, &q.CorrectAnswer,
		&q.WrongAnswer1, &q.WrongAnswer2, &q.WrongAnswer3)
	if err != nil {
		return nil, err
	}
	return q, nil
}

func (r *postgresQuestionRepository) GetCorrectAnswer(id string) (string, error) {
	var answer string
	err := r.db.QueryRow(`SELECT correct_answer FROM questions WHERE id = $1`, id).Scan(&answer)
	return answer, err
}

func (r *postgresQuestionRepository) GetAllWithTopic() ([]map[string]interface{}, error) {
	rows, err := r.db.Query(`
		SELECT q.id, t.name, q.question_text, q.correct_answer,
		       COALESCE(q.wrong_answer1, ''), COALESCE(q.wrong_answer2, ''), COALESCE(q.wrong_answer3, '')
		FROM questions q
		JOIN topics t ON q.topic_id = t.id
		ORDER BY q.id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []map[string]interface{}
	for rows.Next() {
		var id int
		var topicName, text, correct, w1, w2, w3 string
		if err := rows.Scan(&id, &topicName, &text, &correct, &w1, &w2, &w3); err != nil {
			return nil, err
		}
		result = append(result, map[string]interface{}{
			"ID":               id,
			"Тема":             topicName,
			"Вопрос":           text,
			"Правильный ответ": correct,
			"Неправильный 1":   w1,
			"Неправильный 2":   w2,
			"Неправильный 3":   w3,
		})
	}
	return result, nil
}

func (r *postgresQuestionRepository) Create(topicID, text, correct, w1 string, w2, w3 interface{}) error {
	_, err := r.db.Exec(`
		INSERT INTO questions (topic_id, question_text, correct_answer, wrong_answer1, wrong_answer2, wrong_answer3)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, topicID, text, correct, w1, w2, w3)
	return err
}

func (r *postgresQuestionRepository) Update(id, topicID, text, correct, w1 string, w2, w3 interface{}) error {
	_, err := r.db.Exec(`
		UPDATE questions
		SET topic_id = $1, question_text = $2, correct_answer = $3,
		    wrong_answer1 = $4, wrong_answer2 = $5, wrong_answer3 = $6
		WHERE id = $7
	`, topicID, text, correct, w1, w2, w3, id)
	return err
}

func (r *postgresQuestionRepository) Delete(id string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err = tx.Exec(`DELETE FROM questions WHERE id = $1`, id); err != nil {
		return err
	}
	if _, err = tx.Exec(`SELECT setval('questions_id_seq', COALESCE((SELECT MAX(id) FROM questions), 0) + 1, false)`); err != nil {
		return err
	}
	return tx.Commit()
}
