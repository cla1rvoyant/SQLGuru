package service

import (
	"database/sql"
	"errors"
	"math/rand/v2"

	"sqlguru/internal/domain"
	"sqlguru/internal/repository"
)

var ErrNoMoreQuestions = errors.New("no more questions")

type QuizService struct {
	questions repository.QuestionRepository
	topics    repository.TopicRepository
}

func NewQuizService(q repository.QuestionRepository, t repository.TopicRepository) *QuizService {
	return &QuizService{questions: q, topics: t}
}

func (s *QuizService) GetTopics() ([]domain.Topic, error) {
	return s.topics.GetAll()
}

func (s *QuizService) StartQuiz(topicName string) (*domain.Question, error) {
	q, err := s.questions.GetFirstByTopic(topicName)
	if err != nil {
		return nil, err
	}
	shuffleAnswers(q)
	return q, nil
}

func (s *QuizService) GetQuestion(questionID string) (*domain.Question, error) {
	q, err := s.questions.GetByID(questionID)
	if err != nil {
		return nil, err
	}
	shuffleAnswers(q)
	return q, nil
}

// CheckAndNext checks the selected answer and returns the next question.
// Returns (correct, nil, ErrNoMoreQuestions) when the quiz is finished.
func (s *QuizService) CheckAndNext(questionID, selectedAnswer string) (bool, *domain.Question, error) {
	correctAnswer, err := s.questions.GetCorrectAnswer(questionID)
	if err != nil {
		return false, nil, err
	}

	isCorrect := selectedAnswer == correctAnswer

	next, err := s.questions.GetNextAfter(questionID)
	if errors.Is(err, sql.ErrNoRows) {
		return isCorrect, nil, ErrNoMoreQuestions
	}
	if err != nil {
		return false, nil, err
	}

	shuffleAnswers(next)
	return isCorrect, next, nil
}

func shuffleAnswers(q *domain.Question) {
	answers := []string{q.CorrectAnswer, q.WrongAnswer1, q.WrongAnswer2, q.WrongAnswer3}

	end := len(answers)
	for end > 1 && answers[end-1] == "" {
		end--
	}
	answers = answers[:end]

	rand.Shuffle(len(answers), func(i, j int) { answers[i], answers[j] = answers[j], answers[i] })
	q.ShuffledAnswers = answers
}
