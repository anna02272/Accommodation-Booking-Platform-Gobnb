package domain

import (
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"io"
)

type NotificationCreate struct {
	HostId    string `bson:"host_id" json:"host_id"`
	HostEmail string `bson:"host_email" json:"host_email" validate:"required,host_email"`
	Text      string `bson:"notification_text" json:"notification_text"`
}

type Notification struct {
	ID          primitive.ObjectID `bson:"notification_id,omitempty" json:"notification_id"`
	HostId      string             `bson:"host_id" json:"host_id"`
	HostEmail   string             `bson:"host_email" json:"host_email" validate:"required,host_email"`
	Text        string             `bson:"notification_text" json:"notification_text"`
	DateAndTime primitive.DateTime `bson:"date_and_time" json:"date_and_time"`
}

type Notifications []*Notification

func (o *Notification) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(o)
}

func (o *Notifications) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(o)
}

func (o *Notification) FromJSON(r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(o)
}
