// ABOUTME: Qdrant vector database client management
// ABOUTME: Provides connection setup and health checks for vector operations
package database

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type Qdrant struct {
	baseURL    string
	httpClient *http.Client
}

func NewQdrant(url string) *Qdrant {
	return &Qdrant{
		baseURL: url,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (q *Qdrant) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", q.baseURL+"/", nil)
	if err != nil {
		return err
	}

	resp, err := q.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			// Log the error but don't return it as it's in a defer
			fmt.Printf("Error closing response body: %v\n", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("qdrant health check failed with status: %d", resp.StatusCode)
	}

	return nil
}

func (q *Qdrant) BaseURL() string {
	return q.baseURL
}