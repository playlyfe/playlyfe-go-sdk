![Playlyfe Go SDK](https://dev.playlyfe.com/images/assets/pl-go-sdk.png "Playlyfe Go SDK")

# Playlyfe Go SDK [![Go Report Card](https://goreportcard.com/badge/github.com/playlyfe/playlyfe-go-sdk)](https://goreportcard.com/report/github.com/playlyfe/playlyfe-go-sdk) [![GoDoc](https://godoc.org/github.com/playlyfe/playlyfe-go-sdk?status.svg)](https://godoc.org/github.com/playlyfe/playlyfe-go-sdk) [![coverage](http://gocover.io/_badge/github.com/playlyfe/playlyfe-go-sdk)](https://gocover.io/github.com/playlyfe/playlyfe-go-sdk)

This is the official OAuth 2.0 Go or Golang client SDK for the Playlyfe API.
It supports the `client_credentials` and `authorization code` OAuth 2.0 flows.
For a complete API Reference checkout [Playlyfe Developers](https://dev.playlyfe.com/docs/api.html) for more information.

# Examples
The Playlyfe class allows you to make rest api calls like GET, POST, .. etc.  
To get started create a new playlyfe object using client credentials flow and then start making requests
```go
import "github.com/playlyfe/playlyfe-go-sdk"

type Player struct {
    ID string `json:"id"`
    Alias string `json:"alias"`
}

var johny Player

func main() { 
    pl := playlyfe.NewClientV2("Your client id", "Your client secret", nil, nil)
    err := pl.Get("/runtime/player", playlyfe.H{"player_id": "johny"}, johny)  // To get player profile
}
```

# Install
```sh
go get github.com/playlyfe/playlyfe-go-sdk
```
# Using
### Create a client
  If you haven't created a client for your game yet just head over to [Playlyfe](http://playlyfe.com) and login into your account, and go to the game settings and click on client.

###1. Client Credentials Flow
In the client page select Yes for both the first and second questions
![client](https://cloud.githubusercontent.com/assets/1687946/7930229/2c2f14fe-0924-11e5-8c3b-5ba0c10f066f.png)
```go
import "github.com/playlyfe/playlyfe-go-sdk"

pl := playlyfe.NewClientV2("Your client id", "Your client secret", nil, nil)
```
###2. Authorization Code Flow
In the client page select yes for the first question and no for the second
![auth](https://cloud.githubusercontent.com/assets/1687946/7930231/2c31c1fe-0924-11e5-8cb5-73ca0a002bcb.png)
```go
import "github.com/playlyfe/playlyfe-go-sdk"

pl := playlyfe.NewCodeV2("Your client id", "Your client secret", "redirect_uri", nil, nil)
```
In development the sdk caches the access token in memory so you don"t need to  the persist access token object. But in production it is highly recommended to persist the token to a database. It is very simple and easy to do it with redis. You can see the test cases for more examples.

In production you need to pass the Load and Store functions whose signature is like this,
```go
load func() (token string, expires_at int64) {
    println("Loading from redis")
    return "", 50
},
store func(token string, expires_at int64) {
    println("Storing to redis")
}
```
## 3. Custom Login Flow using JWT(JSON Web Token)
In the client page select no for the first question and yes for the second
![jwt](https://cloud.githubusercontent.com/assets/1687946/7930230/2c2f2caa-0924-11e5-8dcf-aed914a9dd58.png)
```go
import "github.com/playlyfe/playlyfe-go-sdk"

token, err := playlyfe.createJWT("your client_id", "your client_secret", 
    "player_id", // The player id associated with your user
    []string{"player.runtime.read", "player.runtime.write"}, // The scopes the player has access to
    3600; // 1 hour expiry Time
)
```
This is used to create jwt token which can be created when your user is authenticated. This token can then be sent to the frontend and or stored in your session. With this token the user can directly send requests to the Playlyfe API as the player.

# Client Scopes
![Client](https://cloud.githubusercontent.com/assets/1687946/9349193/e00fe91c-465f-11e5-8094-6e03c64a662c.png)

Your client has certain access control restrictions. There are 3 kind of resources in the Playlyfe REST API they are,

1.`/admin` -> routes for you to perform admin actions like making a player join a team

2.`/design` -> routes for you to make design changes programmatically

3.`/runtime` -> routes which the users will generally use like getting a player profile, playing an action

The resources accessible to this client can be configured to have a read permission that means only `GET` requests will work.

The resources accessible to this client can be configured to have a write permission that means only `POST`, `PATCH`, `PUT`, `DELETE` requests will work.

The version restriction is only for the design resource and can be used to restrict the client from accessing any version of the game design other than the one specified. By default it allows all.

If access to a route is not allowed and then you make a request to that route then you will get an error like this,
```json
{
  "error": "access_denied",
  "error_description": "You are not allowed to access this api route"
}
```

# Methods
**API**
```go
error API("GET", // The request method can be GET/POST/PUT/PATCH/DELETE
    "", // The api route to get data from
    playlyfe.H{}, // The query params that you want to send to the route
    struct{} ,// The data you want to post to the api
    result interface{}, // The unmarshalled data
    false // Whether you want the response to be in raw string form or json
)
```
**Get**
```go
error Get("", // The api route to get data from
    playlyfe.H{}, // The query params that you want to send to the,
    result interface{}, // The unmarshalled data
)
```
**Post**
```go
error Post("", // The api route to post data to
    playlyfe.H{}, // The query params that you want to send to the route
    struct{},// The data you want to post to the api
    result interface{}, // The unmarshalled data
)
```
**Patch**
```go
error Patch("" // The api route to patch data
    playlyfe.H{} // The query params that you want to send to the route
    struct{} ,// The data you want to post to the api
    result interface{}, // The unmarshalled data
)
```
**Put**
```go
error Put("" // The api route to put data
    playlyfe.H{}, // The query params that you want to send to the route
    struct{} ,// The data you want to post to the api
    result interface{}, // The unmarshalled data
)
```
**Delete**
```go
error Delete("" // The api route to delete the component
    playlyfe.H{} // The query params that you want to send to the route,
    result interface{}, // The unmarshalled data
)
```
**Get Login URL**
```go
GetLoginURL() string
//This will return the url to which the user needs to be redirected for the user to login.
```

**Exchange Code**
```go
ExchangeCode(code string)
//This is used in the auth code flow so that the sdk can get the access token.
//Before any request to the playlyfe api is made this has to be called atleast once.
//This should be called in the the route/controller which you specified in your redirect_uri
```

**Errors**  
An ```*Error``` is returned whenever an error from the PlaylyfeAPI occurs in each call.The Error contains a Name and Description field which can be used to determine the type of error that occurred.

You have to type cast the error to a *PlaylyfeError first like this,
```go
err := pl.Get("/player", playlyfe.H{"player_id": "student1"}, &player)
if pe, ok := err.(*playlyfe.Error); ok {
    return pe
} else {
    panic(err) // do what needs to be done on this type of error
}
```


You can read the docs at [GoDoc](https://godoc.org/github.com/playlyfe/playlyfe-go-sdk)


License
=======
Playlyfe Go SDK  
http://dev.playlyfe.com/  
Copyright(c) 2014-2015, Playlyfe IT Solutions Pvt. Ltd, support@playlyfe.com

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
