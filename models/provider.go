package models

type Provider struct {
	BaseModel
	BaseUrl         string
	Type            string
	EncryptedAPIKey string `json:"-"`
	Requests        int
	Interval        int
	Unit            string
}

type ProviderCreate struct {
	// TODO add ressources
	ID string `json:"id"`
}

type ProviderUpdate struct {
	// TODO add ressources
	ID string `json:"id"`
}
