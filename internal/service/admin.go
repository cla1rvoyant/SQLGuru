package service

import (
	"errors"

	"sqlguru/internal/domain"
	"sqlguru/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

var errProtectedAdmin = errors.New("cannot delete the main administrator")

type AdminService struct {
	admins    repository.AdminRepository
	topics    repository.TopicRepository
	questions repository.QuestionRepository
}

func NewAdminService(a repository.AdminRepository, t repository.TopicRepository, q repository.QuestionRepository) *AdminService {
	return &AdminService{admins: a, topics: t, questions: q}
}

func (s *AdminService) Authenticate(login, password string) error {
	hashed, err := s.admins.GetPasswordByLogin(login)
	if err != nil {
		return err
	}
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))
}

func (s *AdminService) GetTopicsWithID() ([]map[string]interface{}, error) {
	return s.topics.GetAllWithID()
}

func (s *AdminService) GetTableData(table string) (*domain.TableData, []domain.Topic, error) {
	td := &domain.TableData{TableName: table}

	switch table {
	case "admins":
		td.Headers = []string{"ID", "Логин"}
		rows, err := s.admins.GetAll()
		if err != nil {
			return nil, nil, err
		}
		td.Rows = rows

	case "topics":
		td.Headers = []string{"ID", "Название"}
		raw, err := s.topics.GetAllWithID()
		if err != nil {
			return nil, nil, err
		}
		for _, row := range raw {
			td.Rows = append(td.Rows, map[string]interface{}{
				"ID":       row["id"],
				"Название": row["name"],
			})
		}

	case "questions":
		td.Headers = []string{"ID", "Тема", "Вопрос", "Правильный ответ", "Неправильный 1", "Неправильный 2", "Неправильный 3"}
		rows, err := s.questions.GetAllWithTopic()
		if err != nil {
			return nil, nil, err
		}
		td.Rows = rows
	}

	var topics []domain.Topic
	if table == "questions" {
		var err error
		topics, err = s.topics.GetAll()
		if err != nil {
			return nil, nil, err
		}
	}

	return td, topics, nil
}

func (s *AdminService) GetRecord(table, id string) (map[string]interface{}, error) {
	switch table {
	case "admins":
		login, err := s.admins.GetLoginByID(id)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"login": login}, nil

	case "topics":
		name, err := s.topics.GetNameByID(id)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"name": name}, nil

	case "questions":
		q, err := s.questions.GetByID(id)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"topic_id":       q.TopicID,
			"question_text":  q.Text,
			"correct_answer": q.CorrectAnswer,
			"wrong_answer1":  q.WrongAnswer1,
			"wrong_answer2":  q.WrongAnswer2,
			"wrong_answer3":  q.WrongAnswer3,
		}, nil
	}
	return nil, nil
}

func (s *AdminService) DeleteRecord(table, id string) error {
	switch table {
	case "admins":
		login, err := s.admins.GetLoginByID(id)
		if err != nil {
			return err
		}
		if login == "main" {
			return errProtectedAdmin
		}
		return s.admins.Delete(id)
	case "topics":
		return s.topics.Delete(id)
	case "questions":
		return s.questions.Delete(id)
	}
	return nil
}

func (s *AdminService) CreateRecord(table string, fields map[string]string) error {
	switch table {
	case "admins":
		hashed, err := bcrypt.GenerateFromPassword([]byte(fields["password"]), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		return s.admins.Create(fields["login"], string(hashed))

	case "topics":
		return s.topics.Create(fields["name"])

	case "questions":
		w2, w3 := nullableField(fields["wrong_answer2"]), nullableField(fields["wrong_answer3"])
		return s.questions.Create(
			fields["topic_id"], fields["question_text"],
			fields["correct_answer"], fields["wrong_answer1"], w2, w3,
		)
	}
	return nil
}

func (s *AdminService) UpdateRecord(table, id string, fields map[string]string) error {
	switch table {
	case "admins":
		var hashed *string
		if p := fields["password"]; p != "" {
			h, err := bcrypt.GenerateFromPassword([]byte(p), bcrypt.DefaultCost)
			if err != nil {
				return err
			}
			hs := string(h)
			hashed = &hs
		}
		return s.admins.Update(id, fields["login"], hashed)

	case "topics":
		return s.topics.Update(id, fields["name"])

	case "questions":
		w2, w3 := nullableField(fields["wrong_answer2"]), nullableField(fields["wrong_answer3"])
		return s.questions.Update(
			id, fields["topic_id"], fields["question_text"],
			fields["correct_answer"], fields["wrong_answer1"], w2, w3,
		)
	}
	return nil
}

func nullableField(v string) interface{} {
	if v == "" {
		return nil
	}
	return v
}
