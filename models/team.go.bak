package models

type Team struct {
	ID     string `bson:"_id"`
	Name   string `bson:"name"`
	Domain string `bson:"domain"`
	Token  string `bson:"token"`

	IsDisabled bool `bson:"is_disabled"`
	IsHidden   bool `bson:"is_hidden"`

	Plan string                 `bson:"plan"`
	Icon map[string]interface{} `bson:"icon"`
}
