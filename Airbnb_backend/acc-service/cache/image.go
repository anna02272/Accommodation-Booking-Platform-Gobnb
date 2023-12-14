package cache

import (
	"encoding/json"
	"io"
)

type Image struct {
	ID    string `json:"image_id"`
	AccID string `json:"accommodation_id"`
}

type Images []*Image

func (o *Image) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(o)
}

func (o *Image) FromJSON(r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(o)
}

func (i *Images) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(i)
}

func (i *Images) FromJSON(r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(i)
}
