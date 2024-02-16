package pocketbase

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Record represents the structure of the data returned by the PocketBase API.
type Record struct {
	Archived       bool      `json:"archived"`
	Character      string    `json:"character"`
	CollectionId   string    `json:"collectionId"`
	CollectionName string    `json:"collectionName"`
	Created        string `json:"created"`
	CreatedBy      string    `json:"created_by"`
	Description    string    `json:"description"`
	Id             string    `json:"id"`
	Intro          string    `json:"intro"`
	Model          string    `json:"model"`
	Name           string    `json:"name"`
	Public         bool      `json:"public"`
	Script         string    `json:"script"`
	Sella          bool      `json:"sella"`
	Tasks          []Task    `json:"tasks"`
	Updated        string `json:"updated"`
	VoiceId        string    `json:"voice_id"`
}

// Task represents the structure of tasks within a Record.
type Task struct {
	Character string `json:"character"`
	Id        int    `json:"id"`
	Intro     string `json:"intro"`
	Task      string `json:"task"`
	Title     string `json:"title"`
}

// Client struct to hold base configuration for PocketBase API.
type Client struct {
	BaseURL string
}

// NewClient creates a new instance of the PocketBase client.
func NewClient(baseUrl string) *Client {
	return &Client{BaseURL: baseUrl}
}

// GetRecord fetches a record from a PocketBase collection by ID using the client's base URL and returns a structured Record.
func (c *Client) GetRecord(collectionName, recordId string) (*Record, error) {
	url := fmt.Sprintf("%s/api/collections/%s/records/%s", c.BaseURL, collectionName, recordId)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-200 status code: %d", resp.StatusCode)
	}

	var record Record
	if err := json.NewDecoder(resp.Body).Decode(&record); err != nil {
		return nil, fmt.Errorf("error decoding response body: %w", err)
	}

	return &record, nil
}