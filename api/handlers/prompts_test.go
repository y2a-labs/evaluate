package apihandlers

import (
	"context"
	"reflect"
	"script_validation/models"
	"testing"
)

func TestGetPromptById(t *testing.T) {
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		args    args
		want    *models.Prompt
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetPromptById(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPromptById() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetPromptById() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetPromptByIdAPI(t *testing.T) {
	type args struct {
		context context.Context
		request *models.CreatePromptRequest
	}
	tests := []struct {
		name    string
		args    args
		want    *models.FindOrCreatePromptResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetPromptByIdAPI(tt.args.context, tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPromptByIdAPI() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetPromptByIdAPI() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindOrCreatePrompt(t *testing.T) {
	type args struct {
		text string
	}
	tests := []struct {
		name    string
		args    args
		want    *models.Prompt
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FindOrCreatePrompt(tt.args.text)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindOrCreatePrompt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FindOrCreatePrompt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindOrCreatePromptAPI(t *testing.T) {
	type args struct {
		context context.Context
		request *models.FindOrCreatePromptRequest
	}
	tests := []struct {
		name    string
		args    args
		want    *models.FindOrCreatePromptResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FindOrCreatePromptAPI(tt.args.context, tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindOrCreatePromptAPI() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FindOrCreatePromptAPI() = %v, want %v", got, tt.want)
			}
		})
	}
}
