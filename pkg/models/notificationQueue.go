package models

//NotificationMessage send the notification whena particular folder has been process
type NotificationMessage struct {
	Key       string   `json:"key"`
	Timestamp string   `json:"timestamp"`
	FileCount []string `json:"filecount"`
	Finished  bool     `json:"Finfshed"`
}
