package playlyfe

import (
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"github.com/parnurzeal/gorequest"
	"net/url"
	"strings"
	"time"
)

const (
	authorizationURL string = "https://playlyfe.com/auth"
	tokenURL         string = "https://playlyfe.com/auth/token"
	apiBaseURL       string = "https://api.playlyfe.com/"
)

type (
	// H is a simpler way to declare JSON objects
	H map[string]interface{}

	// A is a simpler way to declare JSON arrays
	A []interface{}

	// Load callback is called when a new request is made
	// you need to return the stored access token and it expiresAt time
	Load func() (token string, expiresAt int64)

	// Store callback is called when a token is requested
	// you need to store the token and its expiresAt time
	Store func(token string, expiresAt int64)

	// Error is any error returned from the Playlyfe API
	Error struct {
		Name        string `json:"error"`
		Description string `json:"error_description"`
	}

	// Playlyfe stores all information related to the client
	Playlyfe struct {
		clientID     string
		clientSecret string
		clientType   string
		version      string
		redirectURI  string
		apiEndpoint  string
		code         string
		accessToken  string
		expiresAt    int64
		client       *gorequest.SuperAgent
		load         Load
		store        Store
	}
)

func (e *Error) Error() string {
	return e.Name + `: ` + e.Description
}

// New create a new playlyfe client
func New(clientID string, clientSecret string, clientType string, version string, redirectURI string, load Load, store Store) *Playlyfe {
	pl := &Playlyfe{
		clientID:     clientID,
		clientSecret: clientSecret,
		clientType:   clientType,
		version:      version,
		redirectURI:  redirectURI,
		client:       gorequest.New(),
		apiEndpoint:  apiBaseURL + version,
	}
	if load == nil {
		load = func() (token string, expiresAt int64) {
			return pl.accessToken, pl.expiresAt
		}
	}
	if store == nil {
		store = func(token string, expiresAt int64) {
			pl.accessToken = token
			pl.expiresAt = expiresAt
		}
	}
	pl.load = load
	pl.store = store
	return pl
}

// NewClientV2 creates a new playlyfe client with client crendentials flow
func NewClientV2(clientID, clientSecret string, load Load, store Store) *Playlyfe {
	return New(clientID, clientSecret, "client", "v2", "", load, store)
}

// NewCodeV2 creates a new playlyfe client with authorization code flow
func NewCodeV2(clientID, clientSecret, redirectURI string, load Load, store Store) *Playlyfe {
	return New(clientID, clientSecret, "code", "v2", redirectURI, load, store)
}

func (p *Playlyfe) checkPlError(body string) error {
	if strings.Contains(body, `"error"`) {
		var plError Error
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

func (p *Playlyfe) newJSONError() error {
	return &Error{"invalid_body", "The Response Body could not be unmarshalled"}
}

func (p *Playlyfe) getToken() (string, int64, error) {
	var postBody string
	if p.clientType == "client" {
		postBody = `{"client_id":"` + p.clientID + `", "client_secret": "` + p.clientSecret + `", "grant_type": "client_credentials"}`
	} else {
		postBody = `{"client_id":"` + p.clientID + `", "client_secret": "` + p.clientSecret + `", "redirect_uri": "` + p.redirectURI + `", "code": "` + p.code + `", "grant_type": "authorization_code"}`
	}
	_, body, errs := p.client.Post(tokenURL).Type("json").Send(postBody).End()
	if errs != nil {
		return "", 0, errs[0]
	}
	plError := p.checkPlError(body)
	if plError != nil {
		return "", 0, plError
	}
	token := H{}
	json.Unmarshal([]byte(body), &token)
	expiresAt := int64(token["expires_in"].(float64)) + time.Now().Unix()
	return token["access_token"].(string), expiresAt, nil
}

func (p *Playlyfe) checkToken(query H) error {
	var err error
	token, expiresAt := p.load()
	if token == "" || expiresAt <= time.Now().Unix() {
		token, expiresAt, err = p.getToken()
		if err != nil {
			return err
		}
		p.store(token, expiresAt)
	}
	query["access_token"] = token
	return nil
}

// API makes a an API request to the Playlyfe API
func (p *Playlyfe) API(method string, route string, query H, postbody interface{}, result interface{}, raw bool) error {
	var body string
	var err error
	var errs []error
	var plError Error
	err = p.checkToken(query)
	if err != nil {
		return err
	}
	params := url.Values{}
	for k, v := range query {
		params.Add(k, v.(string))
	}
	apiRoute := p.apiEndpoint + route + "?" + params.Encode()
	switch method {
	case "GET":
		_, body, errs = p.client.Get(apiRoute).Query(query).End()
	case "POST":
		_, body, errs = p.client.Post(apiRoute).Query(query).Send(postbody).End()
	case "PATCH":
		_, body, errs = p.client.Patch(apiRoute).Query(query).Send(postbody).End()
	case "PUT":
		_, body, errs = p.client.Put(apiRoute).Query(query).Send(postbody).End()
	case "DELETE":
		_, body, errs = p.client.Delete(apiRoute).Query(query).End()
	default:
		_, body, errs = p.client.Head(apiRoute).Query(query).End()
	}
	if errs != nil {
		return errs[0]
	}
	if strings.Contains(body, `"error"`) {
		err = json.Unmarshal([]byte(body), &plError)
		if err != nil {
			return p.newJSONError()
		}
		return &plError
	}
	if raw {
		if str, ok := result.(*[]byte); ok {
			*str = []byte(body)
		}
	} else {
		err = json.Unmarshal([]byte(body), &result)
		if err != nil {
			return p.newJSONError()
		}
	}
	return nil
}

// Get makes a GET API request to the Playlyfe API
func (p *Playlyfe) Get(route string, query H, result interface{}) error {
	return p.API("GET", route, query, nil, result, false)
}

// GetRaw makes a GET API request to the Playlyfe API but returns the raw data
// useful for images
func (p *Playlyfe) GetRaw(route string, query H, result interface{}) error {
	return p.API("GET", route, query, nil, result, true)
}

// Post makes a Post API request to the Playlyfe API
func (p *Playlyfe) Post(route string, query H, body interface{}, result interface{}) error {
	return p.API("POST", route, query, body, result, false)
}

// Patch makes a PATCH API request to the Playlyfe API
func (p *Playlyfe) Patch(route string, query H, body interface{}, result interface{}) error {
	return p.API("PATCH", route, query, body, result, false)
}

// Put makes a PUT API request to the Playlyfe API
func (p *Playlyfe) Put(route string, query H, body interface{}, result interface{}) error {
	return p.API("PUT", route, query, body, result, false)
}

// Delete makes a DELETE API request to the Playlyfe API
func (p *Playlyfe) Delete(route string, query H, result interface{}) error {
	return p.API("DELETE", route, query, nil, result, false)
}

// ExchangeCode sets the code which you got from authorization code flow
func (p *Playlyfe) ExchangeCode(code string) {
	p.code = code
}

// GetLoginURL gets the logout url
func (p *Playlyfe) GetLoginURL() string {
	params := url.Values{}
	params.Add("response_type", "code")
	params.Add("redirect_uri", p.redirectURI)
	params.Add("client_id", p.clientID)
	return authorizationURL + "?" + params.Encode()
}

// CreateJWT creates an new JWT Token which can be used with the Playlyfe API
func CreateJWT(clientID, clientSecret, playerID string, scopes []string, expiry time.Duration) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims["player_id"] = playerID
	token.Claims["scopes"] = scopes
	token.Claims["exp"] = time.Now().Add(time.Second * expiry).Unix()
	tokenString, err := token.SignedString([]byte(clientSecret))
	return clientID + ":" + tokenString, err
}
