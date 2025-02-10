package test_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/go-faker/faker/v4"
)

const SEED_LENGTH = 1000

type FakerKV struct {
	Key   string
	Value string
}

type BatchRequestBody []struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// SeedDatabase seeds the database with random key-value pairs
func Test_SeedDatabase(t *testing.T) {

	var tags []FakerKV
	for i := 0; i < SEED_LENGTH; i++ {
		tag := FakerKV{}
		err := faker.FakeData(&tag)
		if err != nil {
			t.Fatalf("Seed Failed failed: %v", err)
		}
		tags = append(tags, tag)
	}

	// marshal tags to JSON body
	body := BatchRequestBody{}
	for _, tag := range tags {
		body = append(body, struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		}{
			Key:   tag.Key,
			Value: tag.Value,
		})
	}

	// convert to bytes
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("Seed Failed failed: %v", err)
	}
	bodyReader := bytes.NewReader(bodyBytes)

	// request 8080 batch/set
	req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/batch/set", bodyReader)

	if err != nil {
		t.Fatalf("Seed Failed failed: %v", err)
	}

	// set content type
	req.Header.Set("Content-Type", "application/json")

	_, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Seed Failed failed: %v", err)
	}

	println("Database seeded with", SEED_LENGTH, "key-value pairs")
}
