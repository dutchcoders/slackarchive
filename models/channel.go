package models

import "github.com/nlopes/slack"

type Channel struct {
	ID        string `bson:"_id"`
	Name      string `bson:"name"`
	Team      string `bson:"team"`
	IsChannel bool   `bson:"is_channel"`
	//Created    time.Time `bson:"created"`
	Creator    string   `bson:"creator"`
	IsArchived bool     `bson:"is_archived"`
	IsGeneral  bool     `bson:"is_general"`
	IsGroup    bool     `bson:"is_group"`
	IsStarred  bool     `bson:"is_starred"`
	Members    []string `bson:"members"`
	Topic      Topic    `bson:"topic"`
	Purpose    Purpose  `bson:"purpose"`
	IsMember   bool     `bson:"is_member"`
	LastRead   string   `bson:"last_read,omitempty"`
	//Latest             Message        `bson:"latest,omitempty"`
	UnreadCount        int `bson:"unread_count,omitempty"`
	NumMembers         int `bson:"num_members,omitempty"`
	UnreadCountDisplay int `bson:"unread_count_display,omitempty"`
}

// Purpose contains information about the topic
type Purpose struct {
	Value   string         `bson:"value"`
	Creator string         `bson:"creator"`
	LastSet slack.JSONTime `bson:"last_set"`
}

// Topic contains information about the topic
type Topic struct {
	Value   string         `bson:"value"`
	Creator string         `bson:"creator"`
	LastSet slack.JSONTime `bson:"last_set"`
}
