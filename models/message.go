package models

import "encoding/json"

// Msg contains information about a slack message
type Message struct {
	// Basic Message
	ID              string `json:"id,omitempty" bson:"_id"`
	Type            string `json:"type,omitempty" bson:"type,omitempty"`
	Channel         string `json:"channel,omitempty" bson:"channel,omitempty"`
	User            string `json:"user,omitempty" bson:"user,omitempty"`
	Text            string `json:"text,omitempty" bson:"text,omitempty"`
	Timestamp       string `json:"ts,omitempty" bson:"ts,omitempty"`
	ThreadTimestamp string `json:"thread_ts,omitempty" bson:"thread_ts,omitempty"`

	IsStarred   bool         `json:"is_starred,omitempty" bson:"is_starred,omitempty"`
	PinnedTo    []string     `json:"pinned_to,omitempty" bson:"pinned_to,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty" bson:"attachments,omitempty"`
	Edited      *Edited      `json:"edited,omitempty" bson:"edited,omitempty"`

	// Message Subtypes
	SubType string `json:"subtype,omitempty" bson:"subtype,omitempty"`

	// Hidden Subtypes
	Hidden           bool   `json:"hidden,omitempty" bson:"hidden,omitempty"`         // message_changed, message_deleted, unpinned_item
	DeletedTimestamp string `json:"deleted_ts,omitempty" bson:"deleted_ts,omitempty"` // message_deleted
	EventTimestamp   string `json:"event_ts,omitempty" bson:"event_ts,omitempty"`

	// bot_message (https://api.slack.com/events/message/bot_message)
	BotID    string `json:"bot_id,omitempty" bson:"bot_id,omitempty"`
	Username string `json:"username,omitempty" bson:"username,omitempty"`
	Icons    *Icon  `json:"icons,omitempty" bson:"icons,omitempty"`

	// channel_join, group_join
	Inviter string `json:"inviter,omitempty" bson:"inviter,omitempty"`

	// channel_topic, group_topic
	Topic string `json:"topic,omitempty" bson:"topic,omitempty"`

	// channel_purpose, group_purpose
	Purpose string `json:"purpose,omitempty" bson:"purpose,omitempty"`

	// channel_name, group_name
	Name    string `json:"name,omitempty" bson:"name,omitempty"`
	OldName string `json:"old_name,omitempty" bson:"old_name,omitempty"`

	// channel_archive, group_archive
	Members []string `json:"members,omitempty" bson:"members,omitempty"`

	// file_share, file_comment, file_mention
	File *File `json:"file,omitempty" bson:"file,omitempty"`

	// file_share
	Upload bool `json:"upload,omitempty" bson:"upload,omitempty"`

	// file_comment
	Comment *Comment `json:"comment,omitempty" bson:"comment,omitempty"`

	// pinned_item
	ItemType string `json:"item_type,omitempty" bson:"item_type,omitempty"`

	// https://api.slack.com/rtm
	ReplyTo int    `json:"reply_to,omitempty" bson:"reply_to,omitempty"`
	Team    string `json:"team,omitempty" bson:"team,omitempty"`

	IsDeleted bool `json:"is_deleted,omitempty" bson:"is_deleted,omitempty"`
}

// Icon is used for bot messages
type Icon struct {
	IconURL   string `json:"icon_url,omitempty" bson:"icon_url,omitempty"`
	IconEmoji string `json:"icon_emoji,omitempty" bson:"icon_emoji,omitempty"`
}

// Edited indicates that a message has been edited.
type Edited struct {
	User      string `json:"user,omitempty" bson:"user,omitempty"`
	Timestamp string `json:"ts,omitempty" bson:"ts,omitempty"`
}

// AttachmentField contains information for an attachment field
// An Attachment can contain multiple of these
type AttachmentField struct {
	Title string `json:"title,omitempty" bson:"title"`
	Value string `json:"value,omitempty" bson:"value"`
	Short bool   `json:"short,omitempty" bson:"short"`
}

type AttachmentAction struct {
	Name  string `json:"name,omitempty" bson:"name"`             // Required.
	Text  string `json:"text,omitempty" bson:"text"`             // Required.
	Style string `json:"style,omitempty" bson:"style,omitempty"` // Optional. Allowed values: "default", "primary", "danger"
	Type  string `json:"type,omitempty" bson:"type"`             // Required. Must be set to "button"
	Value string `json:"value,omitempty" bson:"value,omitempty"` // Optional.
}

// Attachment contains all the information for an attachment
type Attachment struct {
	Color    string `json:"color,omitempty,omitempty" bson:"color,omitempty"`
	Fallback string `json:"fallback,omitempty" bson:"fallback"`

	AuthorName    string `json:"author_name,omitempty" bson:"author_name,omitempty"`
	AuthorSubname string `json:"author_subname,omitempty" bson:"author_subname,omitempty"`
	AuthorLink    string `json:"author_link,omitempty" bson:"author_link,omitempty"`
	AuthorIcon    string `json:"author_icon,omitempty" bson:"author_icon,omitempty"`

	Title     string `json:"title,omitempty" bson:"title,omitempty"`
	TitleLink string `json:"title_link,omitempty" bson:"title_link,omitempty"`
	Pretext   string `json:"pretext,omitempty" bson:"pretext,omitempty"`
	Text      string `json:"text,omitempty" bson:"text"`

	ImageURL string `json:"image_url,omitempty" bson:"image_url,omitempty"`
	ThumbURL string `json:"thumb_url,omitempty" bson:"thumb_url,omitempty"`

	Fields     []AttachmentField  `json:"fields,omitempty" bson:"fields,omitempty"`
	Actions    []AttachmentAction `json:"actions,omitempty" bson:"actions,omitempty"`
	MarkdownIn []string           `json:"mrkdwn_in,omitempty" bson:"mrkdwn_in,omitempty"`

	Footer     string `json:"footer,omitempty" bson:"footer,omitempty"`
	FooterIcon string `json:"footer_icon,omitempty" bson:"footer_icon,omitempty"`

	Ts json.Number `json:"ts,omitempty" bson:"ts,omitempty"`
}

// Comment contains all the information relative to a comment
type Comment struct {
	ID string `json:"id,omitempty" bson:"id,omitempty"`
	//Created   time.Time `bson:"created,omitempty"`
	//Timestamp time.Time `bson:"timestamp,omitempty"`
	User    string `json:"user,omitempty" bson:"user,omitempty"`
	Comment string `json:"comment,omitempty" bson:"comment,omitempty"`
}

// File contains all the information for a file
type File struct {
	ID string `json:"id,omitempty" bson:"id"`
	//Created   time.Time `bson:"created"`
	//Timestamp time.Time `bson:"timestamp"`

	Name       string `json:"name,omitempty" bson:"name"`
	Title      string `json:"title,omitempty" bson:"title"`
	Mimetype   string `json:"mimetype,omitempty" bson:"mimetype"`
	Filetype   string `json:"filetype,omitempty" bson:"filetype"`
	PrettyType string `json:"pretty_type,omitempty" bson:"pretty_type"`
	User       string `json:"user,omitempty" bson:"user"`

	Mode         string `json:"mode,omitempty" bson:"mode"`
	Editable     bool   `json:"editable,omitempty" bson:"editable"`
	IsExternal   bool   `json:"is_external,omitempty" bson:"is_external"`
	ExternalType string `json:"external_type,omitempty" bson:"external_type"`

	Size int `json:"size,omitempty" bson:"size"`

	URL                string `json:"url,omitempty" bson:"url"`
	URLDownload        string `json:"url_download,omitempty" bson:"url_download"`
	URLPrivate         string `json:"url_private,omitempty" bson:"url_private"`
	URLPrivateDownload string `json:"url_private_download,omitempty" bson:"url_private_download"`

	Thumb64     string `json:"thumb_64,omitempty" bson:"thumb_64"`
	Thumb80     string `json:"thumb_80,omitempty" bson:"thumb_80"`
	Thumb360    string `json:"thumb_360,omitempty" bson:"thumb_360"`
	Thumb360Gif string `json:"thumb_360_gif,omitempty" bson:"thumb_360_gif"`
	Thumb360W   int    `json:"thumb_360_w,omitempty" bson:"thumb_360_w"`
	Thumb360H   int    `json:"thumb_360_h,omitempty" bson:"thumb_360_h"`

	Permalink        string `json:"permalink,omitempty" bson:"permalink"`
	EditLink         string `json:"edit_link,omitempty" bson:"edit_link"`
	Preview          string `json:"preview,omitempty" bson:"preview"`
	PreviewHighlight string `json:"preview_highlight,omitempty" bson:"preview_highlight"`
	Lines            int    `json:"lines,omitempty" bson:"lines"`
	LinesMore        int    `json:"lines_more,omitempty" bson:"lines_more"`

	IsPublic        bool     `json:"is_public,omitempty" bson:"is_public"`
	PublicURLShared bool     `json:"public_url_shared,omitempty" bson:"public_url_shared"`
	Channels        []string `json:"channels,omitempty" bson:"channels"`
	Groups          []string `json:"groups,omitempty" bson:"groups"`
	InitialComment  Comment  `json:"initial_comment,omitempty" bson:"initial_comment"`
	NumStars        int      `json:"num_stars,omitempty" bson:"num_stars"`
	IsStarred       bool     `json:"is_starred,omitempty" bson:"is_starred"`
}
