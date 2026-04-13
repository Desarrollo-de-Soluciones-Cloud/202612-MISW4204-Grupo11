package domain

type Attachment struct {
	ID          int
	TaskID      int
	FileName    string
	ContentType string
	StoragePath string
}
