package apihandlers

import (
	"context"
	"log"
	"reflect"
	database "script_validation"
	"script_validation/models"
	"testing"

	"github.com/joho/godotenv"
)

func Test_processMessage(t *testing.T) {
	type args struct {
		p ProcessMessage
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processMessage(tt.args.p)
		})
	}
}

func TestCreateLLMEvaluation(t *testing.T) {
	type args struct {
		ctx   context.Context
		input *models.CreateLLMEvaluationRequest
	}
	tests := []struct {
		name    string
		args    args
		want    *LLMEvaluationOutput
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreateLLMEvaluationAPI(tt.args.ctx, tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateLLMEvaluation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateLLMEvaluation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindOrCreateLLMs(t *testing.T) {
	godotenv.Load()
	database.ConnectDB(":memory:")
	llms := []models.LLM{
		{
			Name: "model1",
		},
		{
			Name: "model2",
		},
	}
	database.DB.Create(&llms)

	type args struct {
		model_names []string
	}
	tests := []struct {
		name    string
		args    args
		want    []models.LLM
		wantErr bool
	}{
		{
			name: "Create new models",
			args: args{
				model_names: []string{"model1", "model2"},
			},
			want:    llms,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FindOrCreateLLMs(tt.args.model_names)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindOrCreateLLMs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for i, gotLLM := range got {
				if gotLLM.Name != tt.want[i].Name {
					t.Errorf("FindOrCreateLLMs() = %v, want %v", gotLLM.Name, tt.want[i].Name)
				}
			}
		})
	}
}

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

func TestGetEvaluationRoutines(t *testing.T) {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	database.ConnectDB(":memory:")
	conversation := models.Conversation{
		Name: "Test Conversation",
		Messages: []models.Message{
			{
				ChatMessage: models.ChatMessage{
					Role:    "user",
					Content: "test content",
				},
			},
			{
				ChatMessage: models.ChatMessage{
					Role:    "assistant",
					Content: "Hello!",
				},
			},
		},
	}
	req := &models.CreateLLMEvaluationRequest{
		Body: struct {
			Models    []string "json:\"models\" default:\"openchat/openchat-7b\" required:\"true\""
			TestCount int      "json:\"test_count\" minimum:\"1\" maximum:\"10\" default:\"1\" required:\"true\""
			Prompt    string   "json:\"prompt\" example:\"You are a helpful assistant.\""
			Messages  []struct {
				ID     string "json:\"id\""
				Prompt string "json:\"prompt\""
			} "json:\"messages,omitempty\""
		}{
			Models:    []string{"openchat/openchat-7b"},
			TestCount: 1,
			Prompt:    "You are a helpful assistant.",
			//Messages: []struct{ID string "json:\"id\""; Prompt string "json:\"prompt\""}{},
		},
	}
	type args struct {
		conversation *models.Conversation
		req          *models.CreateLLMEvaluationRequest
	}
	tests := []struct {
		name    string
		args    args
		want    []EvaluationRoutine
		wantErr bool
	}{
		{
			name: "Test GetEvaluationRoutines",
			args: args{
				conversation: &conversation,
				req:          req,
			},
			want: []EvaluationRoutine{
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetEvaluationRoutines(tt.args.conversation, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetEvaluationRoutines() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != 1 {
				t.Errorf("GetEvaluationRoutines() = %v, want %v", len(got), len(conversation.Messages)-1)
			}
		})
	}
}

func TestCreateLLMEvaluationAPI(t *testing.T) {
	type args struct {
		ctx   context.Context
		input *models.CreateLLMEvaluationRequest
	}
	tests := []struct {
		name    string
		args    args
		want    *models.CreateLLMEvaluationResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreateLLMEvaluationAPI(tt.args.ctx, tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateLLMEvaluationAPI() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateLLMEvaluationAPI() = %v, want %v", got, tt.want)
			}
		})
	}
}
