package apihandlers

import (
	"context"
	"reflect"
	"script_validation/models"
	"testing"
)

func TestSetMessageEmbeddings(t *testing.T) {
	type args struct {
		messages *[]models.Message
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetMessageEmbeddings(tt.args.messages)
		})
	}
}

func TestCreateConversation(t *testing.T) {
	type args struct {
		input *models.CreateConversationInput
	}
	tests := []struct {
		name    string
		args    args
		want    *models.Conversation
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreateConversation(tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateConversation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateConversation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateConversationAPI(t *testing.T) {
	type args struct {
		ctx   context.Context
		input *models.CreateConversationInput
	}
	tests := []struct {
		name    string
		args    args
		want    *models.CreateConversationOutput
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreateConversationAPI(tt.args.ctx, tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateConversationAPI() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateConversationAPI() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetConversation(t *testing.T) {
	type args struct {
		id  string
	}
	tests := []struct {
		name    string
		args    args
		want    *models.Conversation
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetConversation(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetConversation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetConversation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetConversationAPI(t *testing.T) {
	type args struct {
		ctx   context.Context
		input *models.GetConversationInput
	}
	tests := []struct {
		name    string
		args    args
		want    *models.CreateConversationOutput
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetConversationAPI(tt.args.ctx, tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetConversationAPI() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetConversationAPI() = %v, want %v", got, tt.want)
			}
		})
	}
}
