package repository

import (
	"database/sql"

	"sqlguru/internal/domain"
)

type postgresTopicRepository struct {
	db *sql.DB
}

func NewTopicRepository(db *sql.DB) TopicRepository {
	return &postgresTopicRepository{db: db}
}

func (r *postgresTopicRepository) GetAll() ([]domain.Topic, error) {
	rows, err := r.db.Query(`SELECT id, name FROM topics ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var topics []domain.Topic
	for rows.Next() {
		var t domain.Topic
		if err := rows.Scan(&t.ID, &t.Name); err != nil {
			return nil, err
		}
		topics = append(topics, t)
	}
	return topics, nil
}

func (r *postgresTopicRepository) GetAllWithID() ([]map[string]interface{}, error) {
	rows, err := r.db.Query(`SELECT id, name FROM topics ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []map[string]interface{}
	for rows.Next() {
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, err
		}
		result = append(result, map[string]interface{}{"id": id, "name": name})
	}
	return result, nil
}

func (r *postgresTopicRepository) GetNameByID(id string) (string, error) {
	var name string
	err := r.db.QueryRow(`SELECT name FROM topics WHERE id = $1`, id).Scan(&name)
	return name, err
}

func (r *postgresTopicRepository) Create(name string) error {
	_, err := r.db.Exec(`INSERT INTO topics (name) VALUES ($1)`, name)
	return err
}

func (r *postgresTopicRepository) Update(id, name string) error {
	_, err := r.db.Exec(`UPDATE topics SET name = $1 WHERE id = $2`, name, id)
	return err
}

func (r *postgresTopicRepository) Delete(id string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err = tx.Exec(`DELETE FROM topics WHERE id = $1`, id); err != nil {
		return err
	}
	if _, err = tx.Exec(`SELECT setval('topics_id_seq', COALESCE((SELECT MAX(id) FROM topics), 0) + 1, false)`); err != nil {
		return err
	}
	return tx.Commit()
}
