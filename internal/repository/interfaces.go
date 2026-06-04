package repository

import "sqlguru/internal/domain"

type QuestionRepository interface {
	GetFirstByTopic(topicName string) (*domain.Question, error)
	GetNextAfter(questionID string) (*domain.Question, error)
	GetByID(id string) (*domain.Question, error)
	GetCorrectAnswer(id string) (string, error)
	GetAllWithTopic() ([]map[string]interface{}, error)
	Create(topicID, text, correct, w1 string, w2, w3 interface{}) error
	Update(id, topicID, text, correct, w1 string, w2, w3 interface{}) error
	Delete(id string) error
}

type TopicRepository interface {
	GetAll() ([]domain.Topic, error)
	GetAllWithID() ([]map[string]interface{}, error)
	GetNameByID(id string) (string, error)
	Create(name string) error
	Update(id, name string) error
	Delete(id string) error
}

type ArticleRepository interface {
	GetByID(id string) (*domain.Article, error)
	GetAll() ([]domain.Article, error)
	Create(topicName, title, content string) error
	Update(id, topicName, title, content string) error
	Delete(id string) error
}

type AdminRepository interface {
	GetPasswordByLogin(login string) (string, error)
	GetLoginByID(id string) (string, error)
	GetAll() ([]map[string]interface{}, error)
	Create(login, hashedPassword string) error
	Update(id, login string, hashedPassword *string) error
	Delete(id string) error
}
