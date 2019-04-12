package ws

import (
	"encoding/json"
	"net/http"

	"github.com/matobi/mam-golib/pkg/errid"
)

// ReplyJSON Return a json object to http.
func ReplyJSON(w http.ResponseWriter, data interface{}) error {
	b, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		return errid.New("failed marshal json").Cause(err)
	}
	return ReplyRawJSON(w, b)
}

// ReplyRawJSON Return a json string to http
func ReplyRawJSON(w http.ResponseWriter, rawJSON []byte) error {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Write([]byte(rawJSON))
	return nil
}

// ReplyError Return a error message and status code to http
func ReplyError(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), errid.GetCode(err))
}
