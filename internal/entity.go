package main

type TableData struct {
	TableName string
	Headers   []string
	Rows      []map[string]interface{}
}

type Topic struct {
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
