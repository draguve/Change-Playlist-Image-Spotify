package main

import (
	"encoding/json"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"log"
	"net/http"

	"github.com/zmb3/spotify"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

const redirectURI = "http://localhost:8080/callback"

var (
	auth  = spotify.NewAuthenticator(redirectURI, spotify.ScopeUserReadPrivate)
	state = "abc123"
)

func main(){
	err := godotenv.Load()
	if err != nil{
		log.Fatalf(err.Error())
	}

	r := gin.Default()
	store := cookie.NewStore([]byte("pioneer123"))
	r.Use(sessions.Sessions("mysession", store))

	r.GET("/",index)
	r.GET("/callback",completeAuth)

	//start server
	r.Run(":8080")
}

func index(c *gin.Context){
	session := sessions.Default(c)
	Token := session.Get("Token")
	if Token == nil{
		//need to login
		url := auth.AuthURL(state)
		c.Redirect(http.StatusTemporaryRedirect, url)
		return
	}
	var token oauth2.Token
	err := json.Unmarshal(Token.([]byte), &token)
	if err != nil {
		c.String(http.StatusBadGateway,"couldn't unmarshal token")
		return
	}
	client := auth.NewClient(&token)
	user, _ := client.CurrentUser()
	c.String(http.StatusOK,user.ID)
}

func completeAuth(c *gin.Context){
	session := sessions.Default(c)
	tok, err := auth.Token(state,c.Request)
	if err != nil {
		c.String(http.StatusForbidden,"Couldn't Get Token")
		log.Println(err)
		return
	}
	jsonBytes, _ := json.Marshal(tok)
	session.Set("Token",jsonBytes)
	err = session.Save()
	if err != nil{
		log.Println(err)
		c.String(http.StatusBadGateway,err.Error())
		return
	}
	c.Redirect(http.StatusTemporaryRedirect,"/")
}