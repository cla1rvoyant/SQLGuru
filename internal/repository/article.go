package repository

import (
	"database/sql"

	"sqlguru/internal/domain"
)

type postgresArticleRepository struct {
	db *sql.DB
}

func NewArticleRepository(db *sql.DB) ArticleRepository {
	return &postgresArticleRepository{db: db}
}

func (r *postgresArticleRepository) GetByID(id string) (*domain.Article, error) {
	a := &domain.Article{}
	err := r.db.QueryRow(
		`SELECT id, topic_name, title, content FROM articles WHERE id = $1`, id,
	).Scan(&a.ID, &a.TopicName, &a.Title, &a.Content)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (r *postgresArticleRepository) GetAll() ([]domain.Article, error) {
	rows, err := r.db.Query(`SELECT id, topic_name, title, content FROM articles ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var articles []domain.Article
	for rows.Next() {
		var a domain.Article
		if err := rows.Scan(&a.ID, &a.TopicName, &a.Title, &a.Content); err != nil {
			return nil, err
		}
		articles = append(articles, a)
	}
	return articles, nil
}
