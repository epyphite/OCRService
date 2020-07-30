package models

type TextractBody struct {
	JobID            string `json:"JobId"`
	Status           string
	API              string
	Timestamp        string
	DocumentLocation DocumentLocation
}

type DocumentLocation struct {
	S3ObjectName string
	S3BucketName string
}
