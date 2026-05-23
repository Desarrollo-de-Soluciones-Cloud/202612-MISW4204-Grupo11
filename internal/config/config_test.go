package config_test

import (
	"testing"

	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/config"
)

func TestLoad_Defaults(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("HTTP_ADDR", "")
	t.Setenv("GROQ_API_KEY", "")
	t.Setenv("GROQ_MODEL", "")
	t.Setenv("GCP_PROJECT_ID", "")
	t.Setenv("PUBSUB_TOPIC_ID", "")
	t.Setenv("PUBSUB_SUBSCRIPTION_ID", "")
	t.Setenv("STORAGE_PROVIDER", "")
	t.Setenv("STORAGE_LOCAL_DIR", "")
	t.Setenv("GCS_BUCKET", "")
	t.Setenv("GCS_REPORTS_PREFIX", "")

	c, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if c.HTTPAddr != ":8080" {
		t.Fatalf("HTTPAddr %q", c.HTTPAddr)
	}
	if c.DBURL == "" {
		t.Fatal("expected default DBURL")
	}
	if c.GroqAPIKey != "" {
		t.Fatalf("expected empty GroqAPIKey, got %q", c.GroqAPIKey)
	}
	if c.GroqModel != "llama-3.1-8b-instant" {
		t.Fatalf("groq model default: %q", c.GroqModel)
	}
	if c.PubSubProjectID == "" || c.PubSubTopicID == "" || c.PubSubSubscriptionID == "" {
		t.Fatalf("pubsub defaults: %+v", c)
	}
	if c.StorageProvider != "local" || c.StorageLocalDir == "" || c.GCSReportsPrefix == "" {
		t.Fatalf("storage defaults: %+v", c)
	}
}

func TestLoad_CustomEnv(t *testing.T) {
	t.Setenv("JWT_SECRET", "x")
	t.Setenv("HTTP_ADDR", ":9090")
	t.Setenv("DATABASE_URL", "postgres://custom")
	t.Setenv("GROQ_API_KEY", "gsk_test123")
	t.Setenv("GROQ_MODEL", "llama-3.3-70b-versatile")
	t.Setenv("GCP_PROJECT_ID", "my-gcp-project")
	t.Setenv("PUBSUB_TOPIC_ID", "reports-topic")
	t.Setenv("PUBSUB_SUBSCRIPTION_ID", "reports-sub")
	t.Setenv("STORAGE_PROVIDER", "gcs")
	t.Setenv("STORAGE_LOCAL_DIR", "/tmp/uploads")
	t.Setenv("GCS_BUCKET", "bucket-test")
	t.Setenv("GCS_REPORTS_PREFIX", "reports-prod")

	c, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if c.HTTPAddr != ":9090" || c.DBURL != "postgres://custom" {
		t.Fatalf("%+v", c)
	}
	if c.GroqAPIKey != "gsk_test123" || c.GroqModel != "llama-3.3-70b-versatile" {
		t.Fatalf("groq %+v", c)
	}
	if c.PubSubProjectID != "my-gcp-project" ||
		c.PubSubTopicID != "reports-topic" ||
		c.PubSubSubscriptionID != "reports-sub" {
		t.Fatalf("pubsub %+v", c)
	}
	if c.StorageProvider != "gcs" ||
		c.StorageLocalDir != "/tmp/uploads" ||
		c.GCSBucket != "bucket-test" ||
		c.GCSReportsPrefix != "reports-prod" {
		t.Fatalf("storage %+v", c)
	}
}
