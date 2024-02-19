package models

type Msg struct {
	Role           string `db:"role" json:"role"`
	Content        string `db:"content" json:"content"`
	MessageIndex   uint   `db:"message_index" json:"message_index"`
	ConversationId string   `db:"conversation_id" json:"conversation_id" hidden:"true"`
}

type MessageModel struct {
	BaseModel
	Msg
}
