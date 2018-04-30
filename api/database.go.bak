package api

import mgo "gopkg.in/mgo.v2"

type database struct {
	Channels *mgo.Collection
	Teams    *mgo.Collection
	Users    *mgo.Collection
	Messages *mgo.Collection
}

func Database(session *mgo.Session) *database {
	mgodb := session.DB("slackarchive")

	db := database{}
	db.Teams = mgodb.C("teams")
	db.Users = mgodb.C("users")
	db.Channels = mgodb.C("channels")
	db.Messages = mgodb.C("messages")

	return &db
}
