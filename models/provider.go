package models

type Provider struct {
	BaseModel
	BaseUrl         string
	Type            string
	EncryptedAPIKey string `json:"-"`
	Requests        int
	Models          []*LLM
	Interval        int
	Unit            string
}

type ProviderCreate struct {
	// TODO add ressources
	Id       string
	BaseUrl  string
	Type     string
	ApiKey   string
	Requests int
	Interval int
	Unit     string
}

type ProviderUpdate struct {
	ApiKey   string
	BaseUrl  string
	Type     string
	Requests int
	Interval int
	Unit     string
}
