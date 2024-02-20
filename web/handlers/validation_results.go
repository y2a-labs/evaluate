package webhandlers

import (
	"fmt"
	"io"
	"os"
	apihandlers "script_validation/api/handlers"

	"gopkg.in/yaml.v2"
)

func GetPayloadFromYAML(filename string) (*apihandlers.ScriptChatInput, error) {
	// Open the file
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Read the file
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Unmarshal the YAML data into the Body field of a ScriptChatInput struct
	payload := &apihandlers.ScriptChatInput{}
	err = yaml.Unmarshal(data, &payload.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	return payload, nil
}
