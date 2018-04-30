package models

// UserProfile contains all the information details of a given user
type UserProfile struct {
	FirstName          string `bson:"first_name"`
	LastName           string `bson:"last_name"`
	RealName           string `bson:"real_name"`
	RealNameNormalized string `bson:"real_name_normalized"`
	Email              string `bson:"email"`
	Skype              string `bson:"skype"`
	Phone              string `bson:"phone"`
	Image24            string `bson:"image_24"`
	Image32            string `bson:"image_32"`
	Image48            string `bson:"image_48"`
	Image72            string `bson:"image_72"`
	Image192           string `bson:"image_192"`
	ImageOriginal      string `bson:"image_original"`
	Title              string `bson:"title"`
}

// User contains all the information of a user
type User struct {
	ID                string      `bson:"_id"`
	Name              string      `bson:"name"`
	Team              string      `bson:"team"`
	Deleted           bool        `bson:"deleted"`
	Color             string      `bson:"color"`
	Profile           UserProfile `bson:"profile"`
	IsBot             bool        `bson:"is_bot"`
	IsAdmin           bool        `bson:"is_admin"`
	IsOwner           bool        `bson:"is_owner"`
	IsPrimaryOwner    bool        `bson:"is_primary_owner"`
	IsRestricted      bool        `bson:"is_restricted"`
	IsUltraRestricted bool        `bson:"is_ultra_restricted"`
	HasFiles          bool        `bson:"has_files"`
	Presence          string      `bson:"presence"`
}
