package apihandlers

import (
	"script_validation/models"

	"github.com/gofiber/fiber/v2"
)

func (app *App) CreateProvider(req *CreateConversationRequest) (*models.Provider, error) {
	encryptedAPIKey, err := app.Encrypt(req.APIKey)
	if err != nil {
		return nil, err
	}
	provider := &models.Provider{
		ID:              req.ID,
		BaseUrl:         req.BaseUrl,
		Requests:        req.Requests,
		Interval:        req.Interval,
		Unit:            req.Unit,
		EncryptedAPIKey: encryptedAPIKey,
	}
	tx := app.Db.Create(&provider)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return provider, nil
}

type CreateConversationRequest struct {
	ID       string `json:"id" form:"id" validate:"required" `
	BaseUrl  string `json:"base_url" validate:"required" form:"base_url"`
	Requests int    `json:"requests" validate:"required" form:"requests"`
	Interval int    `json:"interval" validate:"required" form:"interval"`
	Unit     string `json:"unit" validate:"required" form:"unit"`
	APIKey   string `json:"api_key" validate:"required" form:"api_key"`
}

func (app *App) CreateProviderAPI(ctx *fiber.Ctx) error {
	req := &CreateConversationRequest{}
	err := app.ValidateStruct(ctx, req)
	if err != nil {
		return err
	}

	provider, err := app.CreateProvider(req)
	if err != nil {
		return err
	}
	// Check the Accept header to send the response in the correct format
	if ctx.Get("Accept") == "application/json" {
		return ctx.JSON(provider)
	} else { // assume it's HTML
		return app.Render(ctx, app.Pages.ProviderRow(provider)) // replace "provider" with your template name
	}
}

func (app *App) GetProviders() ([]*models.Provider, error) {
	providers := []*models.Provider{}
	tx := app.Db.Find(&providers)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return providers, nil
}

func (app *App) GetProvidersAPI(ctx *fiber.Ctx) error {
	providers, err := app.GetProviders()
	if err != nil {
		return err
	}
	// Check the Accept header to send the response in the correct format
	if ctx.Get("Accept") == "application/json" {
		return ctx.JSON(providers)
	} else { // assume it's HTML
		return app.Render(ctx, app.Pages.ProvidersPage(providers)) // replace "providers" with your template name
	}
}

func (app *App) DeleteProvider(id string) (error) {
	tx := app.Db.Delete(&models.Provider{ID: id})
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

func (app *App) DeleteProviderAPI(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	err := app.DeleteProvider(id)
	if err != nil {
		return err
	}
	return nil
}