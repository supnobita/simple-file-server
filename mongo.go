package main

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

//https://hackernoon.com/build-restful-api-in-go-and-mongodb-5e7f2ec4be94
//url="mongodb://mongoadmin:secret@127.0.0.1/?connect=direct&authMechanism=SCRAM-SHA-1"
type HashDAO struct {
	Server   string
	Database string
}

var db *mgo.Database

const (
	COLLECTION = "hash"
)

type Hash struct {
	key         string
	orig_path   string
	refer_paths []string
}

func (m *HashDAO) Connect() *mgo.Session {
	session, err := mgo.Dial(m.Server)
	if err != nil {
		panic("connection error " + m.Server + " " + err.Error())
	}
	db = session.DB(m.Database)
	return session
}

//md5:{orig_path,[refer1, refer2]}
func (m *HashDAO) InsertHash(hash Hash) error {
	err := db.C(COLLECTION).Insert(&hash)
	return err
}

func (m *HashDAO) DeletetHash(hash Hash) error {
	err := db.C(COLLECTION).Remove(&hash)
	return err
}

func (m *HashDAO) FindHash(hash string) (Hash, error) {
	var h Hash
	err := db.C(COLLECTION).Find(bson.M{"key": hash}).One(&h)
	return h, err
}

func (m *HashDAO) UpdateHash(hash Hash) error {
	err := db.C(COLLECTION).Update(bson.M{"key": hash.key}, hash)
	return err
}

func (m *HashDAO) DeletetHashByKey(key string) error {
	err := db.C(COLLECTION).Remove(bson.M{"key": key})
	return err
}
