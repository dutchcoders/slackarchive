package api

import (
	"crypto/sha1"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"regexp"
	"strconv"
	"sync"
	"time"

	"context"

	autocert "golang.org/x/crypto/acme/autocert"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	config "github.com/dutchcoders/slackarchive/config"
	models "github.com/dutchcoders/slackarchive/models"
	utils "github.com/dutchcoders/slackarchive/utils"

	handlers "github.com/dutchcoders/slackarchive/api/handlers"

	elastic "gopkg.in/olivere/elastic.v5"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	// "github.com/olivere/elastic"

	logging "github.com/op/go-logging"
	// "github.com/mattbaird/elastigo/lib"
	"net"
	"strings"
)

var log = logging.MustGetLogger("slackarchive-api")

type api struct {
	session *mgo.Session
	es      *elastic.Client
	config  *config.Config
	store   *sessions.CookieStore

	wg sync.WaitGroup

	indexChan chan (Message)

	// Registered connections.
	connections map[*connection]bool

	// Register requests from the connections.
	register chan *connection

	// Unregister requests from connections.
	unregister chan *connection
}

func New(config *config.Config) *api {
	session, err := mgo.Dial(config.Database.DSN)
	if err != nil {
		panic(err)
	}

	session.SetMode(mgo.Monotonic, true)

	es, err := elastic.NewClient(elastic.SetURL(config.ElasticSearch.URL), elastic.SetSniff(true))
	if err != nil {
		panic(err)
	}

	var store = sessions.NewCookieStore(
		[]byte(config.Cookies.AuthenticationKey),
		[]byte(config.Cookies.EncryptionKey),
	)

	return &api{
		session:     session,
		es:          es,
		config:      config,
		store:       store,
		indexChan:   make(chan Message),
		connections: map[*connection]bool{},
		register:    make(chan *connection),
		unregister:  make(chan *connection),
	}
}

func (api *api) indexer() {

	bulk := api.es.Bulk()

	count := 0

	start := time.Now()

	flush := func() {
		if response, err := bulk.Do(context.Background()); err != nil {
			log.Error("Error indexing: ", err.Error())
		} else {
			indexed := response.Indexed()
			count += len(indexed)

			rate := float64(count) / time.Now().Sub(start).Minutes()
			log.Infof("Bulk indexing: %d total %d (%f messages per minute).", len(indexed), rate, count)
		}
	}

	api.wg.Add(1)

	defer flush()
	defer api.wg.Done()

	// do we want to have buffered channels here? in case we cannot connect to mongo
	for {
		select {
		case msg, ok := <-api.indexChan:
			if !ok {
				return
			}

			if msg.Category == "message" {
				message := models.Message{}
				if err := json.Unmarshal(msg.Body, &message); err != nil {
					log.Errorf("Error unmarshaling message: %s\n%s", err.Error(), string(msg.Body))
					continue
				}

				session := api.session.Copy()

				db := Database(session)

				if _, err := db.Messages.UpsertId(message.ID, &message); err != nil {
					log.Error("Error upserting: %s", err.Error())
				}

				session.Close()

				bulk = bulk.Add(elastic.NewBulkIndexRequest().
					Index("slackarchive").
					Type("message").
					Id(message.ID).
					Doc(message),
				)

				if bulk.NumberOfActions() < 100 {
					continue
				}
			}

		case <-time.After(time.Second * 10):
		}

		if bulk.NumberOfActions() == 0 {
			continue
		}

		flush()
	}
}

func (api *api) run() {
	for {
		select {
		case c := <-api.register:
			api.connections[c] = true
		case c := <-api.unregister:
			if _, ok := api.connections[c]; ok {
				delete(api.connections, c)
				close(c.send)
			}
		}
	}
}

func (api *api) teamHandler(ctx *Context) error {
	type TeamResponse struct {
		ID         string `json:"team_id"`
		Domain     string `json:"domain"`
		Name       string `json:"name"`
		IsDisabled bool   `json:"is_disabled"`
		IsHidden   bool   `json:"is_hidden"`

		Plan string                 `json:"plan"`
		Icon map[string]interface{} `json:"icon"`
	}

	response := struct {
		Teams  []TeamResponse `json:"team"`
		Status string         `json:"status"`
	}{}

	//	domain := ctx.r.FormValue("domain")
	// _ = Team(ctx.r)

	qry := bson.M{
		"is_disabled": bson.M{
			"$not": bson.M{"$eq": true},
		},
	}

	if r := ctx.r.Referer(); r == "" {
	} else if _, err := Host(ctx.r); err != nil {
	} else if t, err := api.Team(ctx); err == nil {
		qry["_id"] = t.ID
	} else {
	}

	iter := ctx.db.Teams.Find(qry).Iter()
	defer iter.Close()

	if err := iter.Err(); err != nil {
		fmt.Printf("Error: %s\n", err.Error())
	}

	team := models.Team{}
	for iter.Next(&team) {
		tr := TeamResponse{}
		if err := utils.Merge(&tr, team); err != nil {
			return err
		}

		response.Teams = append(response.Teams, tr)
	}

	return ctx.Write(response)
}

type domainFn func(*models.Team, *Context) error

// TODO: only for auth users send more info
type UserResponse struct {
	ID      string `json:"user_id"`
	Name    string `json:"name"`
	Team    string `json:"team"`
	Deleted bool   `json:"deleted"`
	Color   string `json:"color"`
	Profile struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		RealName  string `json:"real_name"`
		// RealNameNormalized string `json:"real_name_normalized"`
		// Email              string `json:"email"`
		// Skype         string `json:"skype"`
		// Phone         string `json:"phone"`
		Image24       string `json:"image_24"`
		Image32       string `json:"image_32"`
		Image48       string `json:"image_48"`
		Image72       string `json:"image_72"`
		Image192      string `json:"image_192"`
		ImageOriginal string `json:"image_original"`
		Title         string `json:"title"`
	} `json:"profile"`
	// IsBot             bool   `json:"is_bot"`
	// IsAdmin           bool   `json:"is_admin"`
	// IsOwner           bool   `json:"is_owner"`
	// IsPrimaryOwner    bool   `json:"is_primary_owner"`
	// IsRestricted      bool   `json:"is_restricted"`
	// IsUltraRestricted bool   `json:"is_ultra_restricted"`
	// HasFiles          bool   `json:"has_files"`
	// Presence          string `json:"presence"`
}

func (api *api) usersHandler(ctx *Context) error {
	response := struct {
		Users      []UserResponse `json:"users"`
		TotalCount int64          `json:"total"`
	}{}

	var team *models.Team
	if t, err := api.Team(ctx); err == nil {
		team = t
	} else {
		return err
	}

	team_id := team.ID

	offset := int(0)
	if val, err := strconv.Atoi(ctx.r.FormValue("offset")); err == nil {
		offset = val
	}

	size := int(1000)
	if val, err := strconv.Atoi(ctx.r.FormValue("size")); err == nil {
		size = val
	}

	qry := ctx.db.Users.Find(bson.M{"team": team_id})

	if count, err := qry.Count(); err != nil {
		response.TotalCount = int64(count)
	}

	iter := qry.Skip(offset).Limit(size).Iter()
	defer iter.Close()

	user := models.User{}
	for iter.Next(&user) {
		usr := UserResponse{}
		if err := utils.Merge(&usr, user); err != nil {
			log.Error(err.Error())
		}

		response.Users = append(response.Users, usr)
	}

	return ctx.Write(response)
}

func (api *api) channelsHandler(ctx *Context) error {
	type ChannelResponse struct {
		ID         string `json:"channel_id"`
		Name       string `json:"name"`
		Team       string `json:"team"`
		IsChannel  bool   `json:"is_channel"`
		IsArchived bool   `json:"is_archived"`
		IsGeneral  bool   `json:"is_general"`
		IsGroup    bool   `json:"is_group"`
		IsStarred  bool   `json:"is_starred"`
		IsMember   bool   `json:"is_member"`
		Purpose    struct {
			Value string `json:"value"`
		} `json:"purpose"`
		NumMembers int `json:"num_members"`
	}

	response := struct {
		Channels   []ChannelResponse `json:"channels"`
		TotalCount int64             `json:"total"`
	}{}

	var team *models.Team
	if t, err := api.Team(ctx); err == nil {
		team = t
	} else {
		return err
	}

	team_id := team.ID

	offset := int(0)
	if val, err := strconv.Atoi(ctx.r.FormValue("offset")); err == nil {
		offset = val
	}

	size := int(100)
	if val, err := strconv.Atoi(ctx.r.FormValue("size")); err == nil {
		size = val
	}

	qry := ctx.db.Channels.Find(
		bson.M{
			"$and": []bson.M{
				bson.M{"team": team_id},
				bson.M{"is_member": true},
			},
		})

	if count, err := qry.Count(); err != nil {
		fmt.Printf("Error: %#v\n", err.Error())
	} else {
		response.TotalCount = int64(count)
	}

	iter := qry.
		Skip(offset).
		Limit(size).
		Iter()

	channels := []models.Channel{}
	if err := iter.All(&channels); err != nil {
		return err
	}

	for _, channel := range channels {
		chnl := ChannelResponse{}
		if err := utils.Merge(&chnl, channel); err != nil {
			log.Error(err.Error())
		}

		response.Channels = append(response.Channels, chnl)
	}

	return ctx.Write(response)
}

func All(items []interface{}, fn func(interface{}) error) error {
	for _, item := range items {
		err := fn(item)
		if err == nil {
			continue
		}

		return err
	}

	return nil
}

// AttachmentField contains information for an attachment field
// An Attachment can contain multiple of these
type AttachmentField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

// Attachment contains all the information for an attachment
type Attachment struct {
	Color    string `json:"color,omitempty"`
	Fallback string `json:"fallback"`

	AuthorName    string `json:"author_name,omitempty"`
	AuthorSubname string `json:"author_subname,omitempty"`
	AuthorLink    string `json:"author_link,omitempty"`
	AuthorIcon    string `json:"author_icon,omitempty"`

	Title     string `json:"title,omitempty"`
	TitleLink string `json:"title_link,omitempty"`
	Pretext   string `json:"pretext,omitempty"`
	Text      string `json:"text"`

	ImageURL string `json:"image_url,omitempty"`
	ThumbURL string `json:"thumb_url,omitempty"`

	Fields     []AttachmentField `json:"fields,omitempty"`
	MarkdownIn []string          `json:"mrkdwn_in,omitempty"`
}

func DebugQuery(q elastic.Query) error {
	ss := elastic.NewSearchSource().Query(q)

	src, err := ss.Source()
	if err != nil {
		return err
	}

	out, err := json.MarshalIndent(src, "", "  ")
	if err != nil {
		return err
	}

	log.Info("%s", string(out))
	return err
}

func Host(r *http.Request) (string, error) {
	if h := r.FormValue("host"); h != "" {
		return h, nil
	}

	referer := r.Referer()

	if v := r.Header.Get("X-Alt-Referer"); v != "" {
		referer = v
	}

	if v := r.Header.Get("x-alt-referer"); v != "" {
		referer = v
	}

	if u, err := url.Parse(referer); err != nil {
		return "", err
	} else if h, _, err := net.SplitHostPort(u.Host); err == nil {
		return h, nil
	} else {
		return u.Host, nil
	}

}

func (api *api) Team(ctx *Context) (*models.Team, error) {
	host := ""

	r := ctx.r

	if h, err := Host(r); err == nil {
		host = h
	} else {
		return nil, err
	}

	qry := bson.M{
		"custom_domain": host,
	}

	team := models.Team{}
	if err := ctx.db.Teams.Find(qry).One(&team); err == nil {
		return &team, nil
	} else if err := ctx.db.Teams.Find(
		bson.M{
			"is_disabled": bson.M{
				"$not": bson.M{"$eq": true},
			},
			"domain": api.config.Team,
		}).One(&team); err == nil {
		return &team, nil
	} else {
		log.Errorf("Error: %#v\n", err.Error())
		return nil, fmt.Errorf("Team is disabled or does not exist")

	}

	return &team, nil

}

func (api *api) messagesHandler(ctx *Context) error {
	type MessageResponse struct {
		Text            string `json:"text"`
		Channel         string `json:"channel"`
		User            string `json:"user"`
		Type            string `json:"type"`
		Timestamp       string `json:"ts"`
		ThreadTimestamp string `json:"thread_ts,omitempty"`

		IsStarred   bool         `json:"is_starred,omitempty"`
		PinnedTo    []string     `json:"pinned_to,omitempty"`
		Attachments []Attachment `json:"attachments,omitempty"`
		// Edited      *Edited      `json:"edited,omitempty"`

		// Message Subtypes
		SubType string `json:"subtype,omitempty"`

		// Hidden Subtypes
		Hidden           bool   `json:"hidden,omitempty"`     // message_changed, message_deleted, unpinned_item
		DeletedTimestamp string `json:"deleted_ts,omitempty"` // message_deleted
		EventTimestamp   string `json:"event_ts,omitempty"`

		// bot_message (https://api.slack.com/events/message/bot_message)
		BotID    string `json:"bot_id,omitempty"`
		Username string `json:"username,omitempty"`

		Icons struct {
			IconURL   string `json:"icon_url,omitempty"`
			IconEmoji string `json:"icon_emoji,omitempty"`
		} `json:"icons,omitempty"`

		// channel_join, group_join
		Inviter string `json:"inviter,omitempty"`

		// channel_topic, group_topic
		Topic string `json:"topic,omitempty"`

		// channel_purpose, group_purpose
		Purpose string `json:"purpose,omitempty"`

		// channel_name, group_name
		Name    string `json:"name,omitempty"`
		OldName string `json:"old_name,omitempty"`

		// channel_archive, group_archive
		Members []string `json:"members,omitempty"`

		// file_share, file_comment, file_mention
		// File *File `json:"file,omitempty"`

		// file_share
		Upload bool `json:"upload,omitempty"`

		// file_comment
		// Comment *Comment `json:"comment,omitempty"`

		// pinned_item
		ItemType string `json:"item_type,omitempty"`

		// https://api.slack.com/rtm
		ReplyTo int    `json:"reply_to,omitempty"`
		Team    string `json:"team,omitempty"`
	}

	response := struct {
		Messages   []MessageResponse `json:"messages"`
		TotalCount int64             `json:"total"`
		Aggs       struct {
			Buckets map[string]int64 `json:"buckets"`
		} `json:"aggs"`
		Related struct {
			Users map[string]UserResponse `json:"users"`
		} `json:"related"`
	}{
		Messages: []MessageResponse{},
		Aggs: struct {
			Buckets map[string]int64 `json:"buckets"`
		}{
			Buckets: map[string]int64{},
		},
		Related: struct {
			Users map[string]UserResponse `json:"users"`
		}{
			Users: map[string]UserResponse{},
		},
	}

	_ = response

	var team *models.Team
	if t, err := api.Team(ctx); err == nil {
		team = t
	} else {
		return err
	}

	qs := elastic.NewBoolQuery()

	qs = qs.Must(elastic.NewMatchAllQuery())
	if val := ctx.r.FormValue("q"); val != "" {
		qs = qs.Must(elastic.NewQueryStringQuery(val).DefaultOperator("AND"))
	}

	var pf = elastic.NewBoolQuery()

	var fq = elastic.NewBoolQuery()

	fq = fq.Must(elastic.NewTermsQuery("team.raw", team.ID))

	channels := []models.Channel{}

	// check if bot have been removed from the channel
	if channel := ctx.r.FormValue("channel"); channel != "" {
		//pf = pf.Must(elastic.NewTermQuery("Channel.raw", channel))
		if err := ctx.db.Channels.Find(
			bson.M{
				"$and": []bson.M{
					bson.M{"_id": channel},
					bson.M{"team": team.ID},
					bson.M{"is_member": true},
				},
			}).All(&channels); err != nil {
			return err
		}
	} else {
		// only channels where the bot is still archiving
		if err := ctx.db.Channels.Find(
			bson.M{
				"$and": []bson.M{
					bson.M{"team": team.ID},
					bson.M{"is_member": true},
				},
			}).All(&channels); err != nil {
			return err
		}
	}

	channelsStr := []interface{}{}
	for _, channel := range channels {
		channelsStr = append(channelsStr, channel.ID)
	}

	pf = pf.Must(elastic.NewTermsQuery("channel.raw", channelsStr...))

	if val := ctx.r.FormValue("qfrom"); val == "" {
	} else if qfrom, err := strconv.ParseFloat(ctx.r.FormValue("qfrom"), 64); err != nil {
	} else if val := ctx.r.FormValue("qto"); val == "" {
	} else if qto, err := strconv.ParseFloat(ctx.r.FormValue("qto"), 64); err != nil {
	} else {
		log.Info("Using qfrom and qto")
		qs = qs.Must(elastic.NewRangeQuery("ts.float").Gte(qfrom).Lt(qto))
	}

	// TODO: check if channel is public
	/*
		if count, err := ctx.db.Teams.Find(bson.M{
			"is_disabled": bson.M{
				"$not": bson.M{"$eq": true},
			},
			"_id": team_id,
		}).Count(); err != nil {
			fmt.Printf("Error: %#v\n", err.Error())
			return err
		} else if count == 0 {
			return fmt.Errorf("Team is disabled or does not exist")
		}
	*/

	if val := ctx.r.FormValue("thread"); val == "" {
	} else if val, err := strconv.ParseFloat(val, 64); err != nil {
	} else {
		fq = fq.Must(elastic.NewTermQuery("thread_ts.float", val))
	}

	from := float64(0)
	if val, err := strconv.ParseFloat(ctx.r.FormValue("from"), 64); err == nil {
		from = val
	}

	to := float64(time.Now().Unix())
	if val, err := strconv.ParseFloat(ctx.r.FormValue("to"), 64); err == nil {
		to = val
	}

	sortOrder := false
	if val := ctx.r.FormValue("sort"); val == "asc" {
		sortOrder = true
	}

	fq = fq.MustNot(elastic.NewTermQuery("sub_type.raw", "message_changed"))
	fq = fq.MustNot(elastic.NewTermQuery("sub_type.raw", "message_deleted"))
	fq = fq.MustNot(elastic.NewTermQuery("sub_type.raw", "channel_join"))
	fq = fq.MustNot(elastic.NewTermQuery("sub_type.raw", "channel_leave"))
	fq = fq.MustNot(elastic.NewTermQuery("sub_type.raw", "pinned_item"))
	fq = fq.MustNot(elastic.NewTermQuery("hidden", true))
	fq = fq.Must(elastic.NewRangeQuery("ts.float").Gte(from).Lt(to))

	qs = qs.Filter(fq)

	offset := int(0)
	if val, err := strconv.Atoi(ctx.r.FormValue("offset")); err == nil {
		offset = val
	}

	size := int(100)
	if val, err := strconv.Atoi(ctx.r.FormValue("size")); err != nil {
	} else if size < 500 {
		size = val
	}

	hl := elastic.
		NewHighlight().
		PreTags("[hl]").
		PostTags("[/hl]").
		RequireFieldMatch(false).
		NumOfFragments(0).
		Fields(
			elastic.NewHighlighterField("text"),
			elastic.NewHighlighterField("attachments.text"),
		)

	ss := api.es.Search().
		Index("slackarchive").
		Type("message").
		Query(qs).
		PostFilter(pf).
		Highlight(hl)

	if val := ctx.r.FormValue("aggs"); val == "1" {
		channelAgg := elastic.NewTermsAggregation().Field("Channel.raw").Size(100).OrderByCountDesc()
		ss = ss.Aggregation("channel", channelAgg)
	}

	ss = ss.Sort("ts.float", sortOrder).
		From(offset).
		Size(size)

	func() {
		src, err := qs.Source()
		if err != nil {
			log.Error(err.Error())
			return
		}

		data, err := json.Marshal(src)
		if err != nil {
			log.Error(err.Error())
			return
		}

		s := string(data)
		log.Debug(s)
	}()

	searchResult, err := ss.Do(context.Background())
	if err == nil {
	} else if ee, ok := err.(*elastic.Error); ok {
		json.NewEncoder(os.Stdout).Encode(ee)

		log.Error("Error search: %s %s", ee.Details.Type, ee.Details.Reason)
		log.Error("Error search: %s", ee.Error())
		return ee
	} else {
		return err
	}

	response.TotalCount = searchResult.Hits.TotalHits

	if aggs, ok := searchResult.Aggregations.Terms("channel"); ok {
		for _, bucket := range aggs.Buckets {
			response.Aggs.Buckets[bucket.Key.(string)] = bucket.DocCount
		}
	}

	for _, hit := range searchResult.Hits.Hits {
		var message models.Message
		if err := json.Unmarshal(*hit.Source, &message); err != nil {
			continue
		}

		msg := MessageResponse{}
		if err := utils.Merge(&msg, message); err != nil {
			log.Error(err.Error())
		}

		// update highlight output
		if hit.Highlight != nil {
			if hl, ok := hit.Highlight["text"]; ok {
				msg.Text = hl[0]
			}

			if hl, ok := hit.Highlight["attachments.text"]; ok {
				for i, _ := range hl {
					msg.Attachments[i].Text = hl[i]
				}
			}
		}

		response.Messages = append(response.Messages, msg)
	}

	r := regexp.MustCompile(`\<\@(.+?)\>`)

	userids := []string{}
	for _, message := range response.Messages {
		// extract matches from message text
		func() {
			var matches [][]string
			if matches = r.FindAllStringSubmatch(message.Text, -1); matches == nil {
				return
			}

			for _, match := range matches {
				userids = append(userids, match[1])
			}
		}()

		if message.User == "" {
			continue
		}

		userids = append(userids, message.User)
	}

	iter := ctx.db.Users.Find(
		bson.M{
			"_id": bson.M{
				"$in": userids,
			},
		}).Iter()

	defer iter.Close()

	users := []models.User{}
	if err := iter.All(&users); err != nil {
		return err
	}

	for _, user := range users {
		usr := UserResponse{}
		if err := utils.Merge(&usr, user); err != nil {
			log.Error(err.Error())
		}

		response.Related.Users[user.ID] = usr
	}

	return ctx.Write(response)
}

func (api *api) health(ctx *Context) error {
	ctx.Write("Approaching Neutral Zone, all systems normal and functioning.")
	return nil
}

func (api *api) updateIds(ctx *Context) error {
	iter := ctx.db.Messages.Find(nil).Iter()

	defer iter.Close()

	message := models.Message{}
	for iter.Next(&message) {
		newID := fmt.Sprintf("%s-%s-%s", message.Team, message.Channel, message.Timestamp)
		if newID == message.ID {
			continue
		}

		if err := ctx.db.Messages.RemoveId(message.ID); err != nil {
			log.Errorf("Error during deleting: %s", err.Error())
		}

		message.ID = newID

		if _, err := ctx.db.Messages.UpsertId(message.ID, &message); err != nil {
			log.Errorf("Error during upserting: %s", err.Error())
		}

		log.Info(newID, message.ID)
	}

	return nil
}

func (api *api) reIndex(ctx *Context) error {
	iter := ctx.db.Messages.Find(nil).Batch(1000).Iter()

	defer iter.Close()

	bulk := api.es.Bulk()

	count := 0

	message := models.Message{}
	for iter.Next(&message) {
		if message.IsDeleted {
			continue
		}

		bulk = bulk.Add(elastic.NewBulkIndexRequest().
			Index("slackarchive").
			Type("message").
			Id(message.ID).
			Doc(message),
		)

		if bulk.NumberOfActions() < 1000 {
			continue
		}

		response, err := bulk.Do(context.Background())
		if err != nil {
			log.Error(err.Error())
		} else {
			indexed := response.Indexed()
			count += len(indexed)

			log.Info("Bulk indexing: %d total %d.", len(indexed), count)
		}
	}

	return nil
}

func hash(s string) string {
	h := sha1.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

// serveWs handles websocket requests from the peer.
func (api *api) serveWs(w http.ResponseWriter, r *http.Request) {
	if auth := r.Header.Get("Authorization"); auth == "" {
		w.WriteHeader(403)
		return
	} else if !strings.HasPrefix(auth, "Token") {
		w.WriteHeader(403)
		return
	} else if strings.Compare(auth[6:], hash(api.config.Bot.Token)) != 0 {
		w.WriteHeader(403)
		return
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error("Error upgrading connection:", err)
		return
	}

	c := &connection{send: make(chan []byte, 256), ws: ws, api: api}
	defer c.Close()

	api.register <- c
	log.Infof("Connection upgraded: %s", ws.RemoteAddr())

	go c.readPump()
	c.writePump()
}

func (api *api) Serve() {
	r := mux.NewRouter()

	r.HandleFunc("/health.html", api.ContextHandlerFunc(api.health)).Methods("GET")

	// r.HandleFunc("/updateids", api.ContextHandlerFunc(api.updateIds)).Methods("GET")
	// r.HandleFunc("/reindex", api.ContextHandlerFunc(api.reIndex)).Methods("GET")

	sr := r.PathPrefix("/v1").Subrouter()

	sr.HandleFunc("/messages", api.ContextHandlerFunc(api.messagesHandler)).Methods("GET")
	sr.HandleFunc("/channels", api.ContextHandlerFunc(api.channelsHandler)).Methods("GET")
	sr.HandleFunc("/users", api.ContextHandlerFunc(api.usersHandler)).Methods("GET")
	sr.HandleFunc("/team", api.ContextHandlerFunc(api.teamHandler)).Methods("GET")
	/*
		api.HandleFunc("/messages", messagesHandler).Methods("GET")
		api.HandleFunc("/me", meHandler).Methods("GET")
	*/
	sr.HandleFunc("/oauth/login", api.ContextHandlerFunc(api.oAuthLoginHandler)).Methods("GET")
	sr.HandleFunc("/oauth/callback", api.ContextHandlerFunc(api.oAuthCallbackHandler)).Methods("GET")

	// run websocket server
	go api.run()
	go api.indexer()

	r.HandleFunc("/ws", api.serveWs)

	sh := http.FileServer(
		AssetFS(),
	)

	r.PathPrefix("/").Handler(sh)
	r.NotFoundHandler = sh

	var handler http.Handler = r

	// install middlewares
	handler = handlers.LoggingHandler(handler)
	handler = handlers.RecoverHandler(handler)
	handler = handlers.RedirectHandler(handler)
	handler = handlers.CorsHandler(handler)

	// disable rate limiter for now
	// handler = ratelimit.Request(ratelimit.IP).Rate(30, 60*time.Second).LimitBy(memory.New())(sr)

	httpAddr := api.config.Listen

	log.Infof("SlackArchive server started. %v", httpAddr)
	log.Info("---------------------------")

	if httpsAddr := api.config.ListenTLS; httpsAddr != "" {
		go func() {
			m := autocert.Manager{
				Prompt: autocert.AcceptTOS,
				Cache:  autocert.DirCache(path.Join(api.config.Data, "cache")),
				HostPolicy: func(_ context.Context, host string) error {
					found := true

					/*
						for _, h := range []string{"slackarchive.io"} {
							found = found || strings.HasSuffix(host, h)
						}
					*/

					if !found {
						return errors.New("acme/autocert: host not configured")
					}

					return nil
				},
			}

			handler = m.HTTPHandler(handler)

			// SSL
			s := &http.Server{
				Addr:    httpsAddr,
				Handler: handler,
				TLSConfig: &tls.Config{
					GetCertificate: m.GetCertificate,
				},
			}

			if err := s.ListenAndServeTLS("", ""); err != nil {
				panic(err)
			}
		}()
	}

	stop := make(chan os.Signal, 1)

	signal.Notify(stop, os.Interrupt)

	h := &http.Server{Addr: httpAddr, Handler: handler}

	go func() {
		if err := h.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe %s: %v", httpAddr, err)
		}
	}()

	<-stop
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

	h.Shutdown(ctx)

	//mg.Wait()
}
