package pushqueue

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const (
	StatusInvalidUUID         = "9000" // invalid UUID
	StatusInvalidCode         = "9001" // there's no code
	StatusInvalidSecretKey    = "9002" // invalid secret_key
	StatusPushNotReady        = "9003" // latest push has not been completed
	StatusInternalServerError = "9004" // internal server error
)

// PushResponse contains data responded for push request.
type PushResponse struct {
	// "success" if push request has succeded, "fail" otherwise.
	Result string `json:"result"`

	// Code indicates why push request failed.
	Code string `json:"code,omitempty"`

	ErrorDescription string `json:"error_description,omitempty"`
}

func (r *PushResponse) Error() string {
	// TODO(coffeeport): When error occurs, r.ErrorDescription can be
	// empty string?
	return fmt.Sprintf("%s: %s", r.Code, r.ErrorDescription)
}

type Owner struct {
	UUID      string
	SecretKey string
}

func NewPushRequest(o *Owner, code, body string) *http.Request {
	v := make(url.Values, 4)
	v.Set("uuid", o.UUID)
	v.Set("secret_key", o.SecretKey)
	v.Set("code", code)
	v.Set("body", body)

	r, _ := http.NewRequest("POST", "http://push.doday.net/api/push",
		strings.NewReader(v.Encode()))
	return r
}

func Push(o *Owner, code, body string) error {
	r := NewPushRequest(o, code, body)
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return err
	}
	v := new(PushResponse)
	err = json.NewDecoder(resp.Body).Decode(v)
	if err != nil {
		resp.Body.Close()
		return err
	}
	err = resp.Body.Close()
	if err != nil {
		return err
	}
	if v.Result != "success" {
		return v
	}
	return nil
}

func StickyPush(o *Owner, code, body string) error {
	for {
		err := Push(o, code, body)
		if err == nil {
			return nil
		}
		r, ok := err.(*PushResponse)
		if !ok {
			return err
		}
		if r.Code == StatusPushNotReady {
			continue
		}
		return err
	}
}
