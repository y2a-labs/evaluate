package apihandlers

import (
	"reflect"
	"script_validation/models"
	"testing"
)

/*
func TestProcessMessage(t *testing.T) {
	godotenv.Load("../../.env")
	database.ConnectDB(":memory:")
	conversation, err := CreateConversation("Test Conversation", "user: test content\nassistant: Hello!\nuser: test content\nassistant: Hello!")
	if err != nil {
		log.Fatal(err)
	}
	conversation.EvalTestCount = 2

	llm_model := models.LLM{ID: "openchat/openchat-7b"}
	prompt := "You are a helpful assistant"
	database.DB.Save(&llm_model)
	database.DB.Save(&prompt)

	routine := EvaluationRoutine{
		Conversation: conversation,
		MessageIndex: 2,
		LLM:          llm_model,
		Prompt:       prompt,
	}
	provider := &models.Provider{
		ID:           "groq",
		BaseUrl:      "https://api.groq.com/openai/v1",
		EnvKey: 	 "GROQ_API_KEY",
	}
	llm_client := llm.GetLLMClient(provider)

	embedding_client := nomicai.NewClient(os.Getenv("NOMICAI_API_KEY"))

	evaluation, err := processMessage(
		routine,
		llm_client,
		embedding_client,
	)
	assert.Nil(t, err, "Should not return an error")

		want := &models.MessageEvaluation{
			MessageID:      conversation.Messages[routine.MessageIndex].ID,
			LLMID:          llm_model.ID,
			LLM:			llm_model,
			PromptID:       prompt.ID,
			Prompt: 	   prompt,
			MessageEvaluationResults: []models.MessageEvaluationResult{
				models.MessageEvaluationResult{
					Content:    "test content",
					Similarity: 0.0,
					MessageEvaluationID: evaluation.ID,
				},
			},
		}

	assert.Equal(t, len(evaluation.MessageEvaluationResults), 2, "Should return 2 message evaluations results")
	assert.NotEmpty(t, evaluation.MessageEvaluationResults[0].Content, "Should have content in the message evaluation results")
	assert.NotEmpty(t, evaluation.MessageEvaluationResults[0].Similarity, "Should have similarity in the message evaluation results")
	//assert.Greater(t, evaluation.AverageSimilarity, 0.0, "Should return a positive average similarity")
}
*/

func TestConvertToChatMessages(t *testing.T) {
	type args struct {
		message []models.Message
	}
	tests := []struct {
		name string
		args args
		want []models.ChatMessage
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ConvertToChatMessages(tt.args.message); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ConvertToChatMessages() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetSystemPrompt(t *testing.T) {
	type args struct {
		messages *[]models.Message
		prompt   string
	}
	tests := []struct {
		name    string
		args    args
		want    *[]models.Message
		wantErr bool
	}{
		{
			name: "Prepending a system prompt",
			args: args{
				messages: &[]models.Message{
					{
						ChatMessage: models.ChatMessage{
							Role:    "user",
							Content: "test content",
						},
					},
				},
				prompt: "test prompt",
			},
			want: &[]models.Message{
				{
					ChatMessage: models.ChatMessage{
						Role:    "system",
						Content: "test prompt",
					},
				},
				{
					ChatMessage: models.ChatMessage{
						Role:    "user",
						Content: "test content",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Test updating the system prompt",
			args: args{
				messages: &[]models.Message{
					{
						ChatMessage: models.ChatMessage{
							Role:    "system",
							Content: "test content",
						},
					},
				},
				prompt: "A new prompt",
			},
			want: &[]models.Message{
				{
					ChatMessage: models.ChatMessage{
						Role:    "system",
						Content: "A new prompt",
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := SetSystemPrompt(tt.args.messages, tt.args.prompt); (err != nil) != tt.wantErr {
				t.Errorf("SetSystemPrompt() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(tt.args.messages, tt.want) {
				t.Errorf("SetSystemPrompt() = %v, want %v", tt.args.messages, tt.want)
			}
		})
	}
}

/*
func TestGetEvaluationRoutines(t *testing.T) {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	database.ConnectDB(":memory:")
	conversation := models.Conversation{
		Name: "Test Conversation",
		Messages: []models.Message{
			{ChatMessage: models.ChatMessage{Role: "user", Content: "test content"}},
			{ChatMessage: models.ChatMessage{Role: "assistant", Content: "Hello!"}},
		},
	}
	req := &models.CreateLLMEvaluationRequestBody{
		Models:    []string{"openchat/openchat-7b", "google/gemma-7b-it"},
		TestCount: 1,
		Prompt:    "You are a helpful assistant.",
	}
	// Using a system level prompt
	routines, err := GetEvaluationRoutines(&conversation, req)
	assert.Nil(t, err, "Should not return an error")
	assert.Equal(t, 2, len(routines), "Should return 2 routines")

	// Using a prompt per message
	conversation = models.Conversation{
		Name: "Test Conversation",
		Messages: []models.Message{
			{ChatMessage: models.ChatMessage{Role: "user", Content: "test content"}, BaseModel: models.BaseModel{ID: "1"}},
			{ChatMessage: models.ChatMessage{Role: "assistant", Content: "Hello!"}, BaseModel: models.BaseModel{ID: "2"}},
			{ChatMessage: models.ChatMessage{Role: "user", Content: "test content"}, BaseModel: models.BaseModel{ID: "3"}},
			{ChatMessage: models.ChatMessage{Role: "assistant", Content: "Hello!"}, BaseModel: models.BaseModel{ID: "4"}},
		},
	}
	req = &models.CreateLLMEvaluationRequestBody{
		Models:    []string{"openchat/openchat-7b"},
		TestCount: 1,
		Prompt:    "You are a helpful assistant.",
		Messages: []models.CreateLLMEvaluationMessage{
			{ID: "2", Prompt: "You are a helpful assistant."},
			{ID: "2", Prompt: "What's going on?"},
		},
	}
	routines, err = GetEvaluationRoutines(&conversation, req)
	assert.Nil(t, err, "Should not return an error")
	assert.Equal(t, len(routines), 2, "Should return 2 routines")
}
*/
