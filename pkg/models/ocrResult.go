package models

type OCRResult struct {
	Filename  string   `json:"filename"`
	Timestamp string   `json:"timestamp"`
	Data      []string `json:"data"`
}
