package playlyfe

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func panicOnError(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
		return
	}
}

func getPlError(t *testing.T, err error) *Error {
	if pe, ok := err.(*Error); ok {
		return pe
	}
	panicOnError(t, err)
	return nil
}

type Player struct {
	ID    string `json:"id"`
	Alias string `json:"alias"`
}

type Players struct {
	Data  []Player `json:"data"`
	Total uint64   `json:"total"`
}

type Process struct {
	ID     string `json:"id,omitempty"`
	State  string `json:"state,omitempty"`
	Name   string `json:"name"`
	Access string `json:"access"`
}

var player Player
var players Players
var resp interface{}
var process, processPatched Process

func TestInvalidClient(t *testing.T) {
	pl := NewClientV2("Zjc0MWU0N2MtODkzNS00ZWNmLWEwNmYtY2M1MGMxNGQ", "YNDQtNDMwMC00YTdkLWFiM2MtNTg0Y2ZkOThjYTZkMGIyNWVlNDAtNGJiMC0xMWU0LWI2NGEtYjlmMmFkYTdjOTI3", nil, nil)
	err := pl.Get("/player", H{"player_id": "student1"}, &player)
	assert.Equal(t, getPlError(t, err).Name, "client_auth_fail")
}

func TestWrongRoute(t *testing.T) {
	pl := NewClientV2("Zjc0MWU0N2MtODkzNS00ZWNmLWEwNmYtY2M1MGMxNGQ1YmQ4", "YzllYTE5NDQtNDMwMC00YTdkLWFiM2MtNTg0Y2ZkOThjYTZkMGIyNWVlNDAtNGJiMC0xMWU0LWI2NGEtYjlmMmFkYTdjOTI3", nil, nil)
	err := pl.Get("/qqq", H{"player_id": "student1"}, &player)
	assert.Equal(t, getPlError(t, err).Name, "route_not_found")
}

func TestLoadStore(t *testing.T) {
	pl := NewClientV2(
		"Zjc0MWU0N2MtODkzNS00ZWNmLWEwNmYtY2M1MGMxNGQ1YmQ4",
		"YzllYTE5NDQtNDMwMC00YTdkLWFiM2MtNTg0Y2ZkOThjYTZkMGIyNWVlNDAtNGJiMC0xMWU0LWI2NGEtYjlmMmFkYTdjOTI3",
		func() (token string, expires_at int64) {
			println("Loading")
			return "", 50
		},
		func(token string, expires_at int64) {
			println("Storing")
		},
	)
	pl.checkToken(H{})
}

func TestClientV2(t *testing.T) {
	pl := NewClientV2("Zjc0MWU0N2MtODkzNS00ZWNmLWEwNmYtY2M1MGMxNGQ1YmQ4", "YzllYTE5NDQtNDMwMC00YTdkLWFiM2MtNTg0Y2ZkOThjYTZkMGIyNWVlNDAtNGJiMC0xMWU0LWI2NGEtYjlmMmFkYTdjOTI3", nil, nil)
	err := pl.Get("/runtime/player", H{}, &player)
	assert.Equal(t, getPlError(t, err).Name, "invalid_player")
	err = pl.Get("/runtime/player", H{"player_id": "student1"}, &player)
	panicOnError(t, err)
	assert.Equal(t, player.ID, "student1", "")
	err = pl.Get("/runtime/players", H{"player_id": "student1"}, &players)
	panicOnError(t, err)
	assert.NotNil(t, players.Total)

	// rawData := &Raw{Data: ""}
	// err = pl.GetRaw("/runtime/player", H{}, rawData)
	// println(rawData.Data)
	// assert.Contains(t, rawData, "invalid_player")

	err = pl.Get("/runtime/definitions/processes", H{"player_id": "student1"}, &resp)
	panicOnError(t, err)
	err = pl.Get("/runtime/definitions/teams", H{"player_id": "student1"}, &resp)
	panicOnError(t, err)
	err = pl.Get("/runtime/processes", H{"player_id": "student1"}, &resp)
	panicOnError(t, err)
	err = pl.Get("/runtime/teams", H{"player_id": "student1"}, &resp)
	panicOnError(t, err)
	var pd struct {
		Definition string `json:"definition"`
	}
	pd.Definition = "module1"
	err = pl.Post("/runtime/processes", H{"player_id": "student1"}, pd, &process)
	assert.Equal(t, process.State, "ACTIVE", "")
	err = pl.Post("/runtime/processes", H{"player_id": "student1"}, pd, &process)
	panicOnError(t, err)
	id := process.ID
	process.ID = ""    // Clearing these fields
	process.State = "" // Clearing these fields
	process.Name = "patched_process"
	pl.Patch("/runtime/processes/"+id, H{"player_id": "student1"}, process, &processPatched)
	panicOnError(t, err)
	assert.Equal(t, processPatched.Name, "patched_process", "")
	var msg struct {
		Message string `json:"message"`
	}
	pl.Delete("/runtime/processes/"+id, H{"player_id": "student1"}, &msg)
	assert.Contains(t, msg.Message, "Process")
}

func TestRedirectURI(t *testing.T) {
	pl := NewCodeV2("Zjc0MWU0N2MtODkzNS00ZWNmLWEwNmYtY2M1MGMxNGQ1YmQ4", "YzllYTE5NDQtNDMwMC00YTdkLWFiM2MtNTg0Y2ZkOThjYTZkMGIyNWVlNDAtNGJiMC0xMWU0LWI2NGEtYjlmMmFkYTdjOTI3", "http://localhost/code", nil, nil)
	println(pl.GetLoginURL())
}

func TestJWT(t *testing.T) {
	token, err := CreateJWT(
		"Zjc0MWU0N2MtODkzNS00ZWNmLWEwNmYtY2M1MGMxNGQ1YmQ4",
		"YzllYTE5NDQtNDMwMC00YTdkLWFiM2MtNTg0Y2ZkOThjYTZkMGIyNWVlNDAtNGJiMC0xMWU0LWI2NGEtYjlmMmFkYTdjOTI3",
		"student1",
		[]string{"player.runtime.read", "player.runtime.write"},
		50,
	)
	panicOnError(t, err)
	parts := strings.Split(token, ":")
	assert.Equal(t, parts[0], "Zjc0MWU0N2MtODkzNS00ZWNmLWEwNmYtY2M1MGMxNGQ1YmQ4")
}

func TestGraphQL(t *testing.T) {
	client := &GraphQLClient{
		EndPoint:   "http://localhost:3212/graphql",
		GameID:     "demo",
		GameSecret: "mysupersecretpassword",
		RuntimeID:  "production",
		Branch:     "master",
		Version:    "latest",
		PlayerID:   "admin",
		HTTPClient: &http.Client{},
	}
	token, err := client.getToken()
	panicOnError(t, err)
	parts := strings.Split(token, ":")
	assert.Equal(t, parts[0], "demo")
}
