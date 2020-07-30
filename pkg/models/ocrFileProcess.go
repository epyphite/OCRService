package models

//OCRFileProcess will hold the status for each job and their path
type OCRFileProcess struct {
	FileHash  string `json:"FileHash"`
	TimeStamp string `json:"TimeStamp"`
	Status    string `json:"Status"`
	FileName  string `json:"FileName"`
}
