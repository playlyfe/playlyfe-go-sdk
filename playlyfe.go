package playlyfe

import (
	"encoding/json"
	"github.com/parnurzeal/gorequest"
	"net/url"
	"strings"
	"time"
)

const (
	authorizationUrl string = "https://playlyfe.com/auth"
	tokenUrl         string = "https://playlyfe.com/auth/token"
	apiBaseUrl       string = "https://api.playlyfe.com/"
)

type (
	H map[string]interface{}

	Raw struct {
		Data string
	}

	Load  func() (token string, expires_at int64)
	Store func(token string, expires_at int64)

	PlaylyfeError struct {
		Name        string `json:"error"`
		Description string `json:"error_description"`
	}

	Playlyfe struct {
		client_id     string
		client_secret string
		client_type   string
		version       string
		redirect_uri  string
		apiEndpoint   string
		code          string
		access_token  string
		expires_at    int64
		client        *gorequest.SuperAgent
		load          Load
		store         Store
	}
)

func (self *PlaylyfeError) Error() string {
	return self.Name + `: ` + self.Description
}

func New(client_id string, client_secret string, client_type string, version string, redirect_uri string, load Load, store Store) *Playlyfe {
	pl := &Playlyfe{
		client_id:     client_id,
		client_secret: client_secret,
		client_type:   client_type,
		version:       version,
		redirect_uri:  redirect_uri,
		client:        gorequest.New(),
		apiEndpoint:   apiBaseUrl + version,
	}
	if load == nil {
		load = func() (token string, expires_at int64) {
			return pl.access_token, pl.expires_at
		}
	}
	if store == nil {
		store = func(token string, expires_at int64) {
			pl.access_token = token
			pl.expires_at = expires_at
		}
	}
	pl.load = load
	pl.store = store
	return pl
}

func NewClientV2(client_id, client_secret string, load Load, store Store) *Playlyfe {
	return New(client_id, client_secret, "client", "v2", "", load, store)
}

func NewClientV1(client_id, client_secret string, load Load, store Store) *Playlyfe {
	return New(client_id, client_secret, "client", "v1", "", load, store)
}

func NewCodeV2(client_id, client_secret, redirect_uri string, load Load, store Store) *Playlyfe {
	return New(client_id, client_secret, "code", "v2", redirect_uri, load, store)
}

func NewCodeV1(client_id, client_secret, redirect_uri string, load Load, store Store) *Playlyfe {
	return New(client_id, client_secret, "code", "v1", redirect_uri, load, store)
}

func CreateJWT(client_id, client_secret, player_id string, scopes []string, expiry uint64) string {
	return "JWT"
}

func (self *Playlyfe) checkPlError(body string) error {
	if strings.Contains(body, `"error"`) {
		var plError PlaylyfeError
		err := json.Unmarshal([]byte(body), &plError)
		if err != nil {
			plError.Name = "invalid_body"
			plError.Description = "The Response Body could not be unmarshalled"
			return &plError
		}
		return &plError
	}
	return nil
}

func (self *Playlyfe) newJSONError() error {
	return &PlaylyfeError{"invalid_body", "The Response Body could not be unmarshalled"}
}

func (self *Playlyfe) getToken() (string, int64, error) {
	var post_body string
	if self.client_type == "client" {
		post_body = `{"client_id":"` + self.client_id + `", "client_secret": "` + self.client_secret + `", "grant_type": "client_credentials"}`
	} else {
		post_body = `{"client_id":"` + self.client_id + `", "client_secret": "` + self.client_secret + `", "redirect_uri": "` + self.redirect_uri + `", "code": "` + self.code + `", "grant_type": "authorization_code"}`
	}
	_, body, errs := self.client.Post(tokenUrl).Type("json").Send(post_body).End()
	if errs != nil {
		return "", 0, errs[0]
	}
	plError := self.checkPlError(body)
	if plError != nil {
		return "", 0, plError
	}
	token := H{}
	json.Unmarshal([]byte(body), &token)
	expires_at := int64(token["expires_in"].(float64)) + time.Now().Unix()
	return token["access_token"].(string), expires_at, nil
}

func (self *Playlyfe) checkToken(query H) error {
	var err error
	token, expires_at := self.load()
	if token == "" || expires_at <= time.Now().Unix() {
		token, expires_at, err = self.getToken()
		if err != nil {
			return err
		}
		self.store(token, expires_at)
	}
	query["access_token"] = token
	return nil
}

func (self *Playlyfe) Api(method string, route string, query H, postbody interface{}, result interface{}, raw bool) error {
	var body string
	var err error
	var errs []error
	var plError PlaylyfeError
	err = self.checkToken(query)
	if err != nil {
		return err
	}
	params := url.Values{}
	for k, v := range query {
		params.Add(k, v.(string))
	}
	api_route := self.apiEndpoint + route + "?" + params.Encode()
	switch method {
	case "GET":
		_, body, errs = self.client.Get(api_route).Query(query).End()
	case "POST":
		_, body, errs = self.client.Post(api_route).Query(query).Send(postbody).End()
	case "PATCH":
		_, body, errs = self.client.Patch(api_route).Query(query).Send(postbody).End()
	case "PUT":
		_, body, errs = self.client.Put(api_route).Query(query).Send(postbody).End()
	case "DELETE":
		_, body, errs = self.client.Delete(api_route).Query(query).End()
	default:
		_, body, errs = self.client.Head(api_route).Query(query).End()
	}
	if errs != nil {
		return errs[0]
	}
	if strings.Contains(body, `"error"`) {
		err = json.Unmarshal([]byte(body), &plError)
		if err != nil {
			return self.newJSONError()
		}
		return &plError
	} else {
		if raw {
			// Do nothing for now
			// if str, ok := result.(*Raw); ok {
			// 	str.Data = body
			// }
		} else {
			err = json.Unmarshal([]byte(body), &result)
			if err != nil {
				return self.newJSONError()
			}
		}
	}
	return nil
}

func (self *Playlyfe) Get(route string, query H, result interface{}) error {
	return self.Api("GET", route, query, nil, result, false)
}

func (self *Playlyfe) GetRaw(route string, query H, result interface{}) error {
	return self.Api("GET", route, query, nil, result, true)
}

func (self *Playlyfe) Post(route string, query H, body interface{}, result interface{}) error {
	return self.Api("POST", route, query, body, result, false)
}

func (self *Playlyfe) Patch(route string, query H, body interface{}, result interface{}) error {
	return self.Api("PATCH", route, query, body, result, false)
}

func (self *Playlyfe) Put(route string, query H, body interface{}, result interface{}) error {
	return self.Api("PUT", route, query, body, result, false)
}

func (self *Playlyfe) Delete(route string, query H, result interface{}) error {
	return self.Api("DELETE", route, query, nil, result, false)
}

func (self *Playlyfe) ExchangeCode(code string) {
	self.code = code
}

func (self *Playlyfe) GetLoginUrl() string {
	params := url.Values{}
	params.Add("response_type", "code")
	params.Add("redirect_uri", self.redirect_uri)
	params.Add("client_id", self.client_id)
	return authorizationUrl + "?" + params.Encode()
}

func (self *Playlyfe) GetLogoutUrl() string {
	return ""
}
