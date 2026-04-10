package domain

type Attachment struct {
	ID          string
	TaskID      string
	FileName    string
	ContentType string
	StoragePath string
}
