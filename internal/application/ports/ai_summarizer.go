package ports

import "context"

type AISummarizer interface {
	Summarize(ctx context.Context, prompt string) (string, error)
}
