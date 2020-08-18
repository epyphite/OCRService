package models

//OCRFileProcess will hold the status for each job and their path
type OCRFileProcess struct {
	FileHash                 string `json:"FileHash"`
	TimeStamp                string `json:"TimeStamp"`
	Status                   string `json:"Status"`
	FileName                 string `json:"FileName"`
	CloudDestinationProvider string `json:"CloudDestinationProvider"` // if Enable Cloud you should specify AWS |GCLOUD | AZURE
	CloudDestinationRegion   string `json:"CloudDestinationRegion"`   // If required, default is us-east-2
	CloudDestinationStorage  string `json:"CloudDestinationStorage"`  //Bucket or Root Key folder
	CloudDestinationKey      string `json:"CloudDestinationKey"`      //Folder or path to key
}
