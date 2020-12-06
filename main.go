package main

import (
	"encoding/json"
	"github.com/CloudyKit/jet"
	"github.com/alexsasharegan/dotenv"
	"golang.org/x/oauth2"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/zmb3/spotify"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

const redirectURI = "http://localhost:8080/callback"

var (
	auth  = spotify.NewAuthenticator(redirectURI,spotify.ScopeImageUpload, spotify.ScopePlaylistModifyPublic,spotify.ScopePlaylistModifyPrivate)
	state = "abc123"
	views = jet.NewSet(jet.NewOSFileSystemLoader("./templates"))
)

func main(){
	err := dotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	auth.SetAuthInfo(os.Getenv("SPOTIFY_ID"),os.Getenv("SPOTIFY_SECRET"))

	r := gin.Default()
	store := cookie.NewStore([]byte("pioneer123"))

	r.Use(sessions.Sessions("mysession", store))

	r.GET("/callback",completeAuth)
	r.GET("/",VerifyLogin(),index)
	r.GET("/test",test)

	//start server
	err = r.Run(":8080")
	if err != nil{
		log.Fatalf(err.Error())
	}
}

func test(c *gin.Context){
	t, _ := views.GetTemplate("index.jet.html")
	c.Writer.WriteHeader(200)
	if err := t.Execute(c.Writer, nil, nil); err != nil {
		log.Println(err)
	}
}

func index(c *gin.Context){
	t, _ := views.GetTemplate("index.jet.html")
	vars := make(jet.VarMap)
	token := c.MustGet("Token").(oauth2.Token)
	client := auth.NewClient(&token)
	user,e := client.CurrentUser()
	if e != nil{
		session := sessions.Default(c)
		session.Set("Token",[]byte{})
		_ = session.Save()
		c.Redirect(http.StatusTemporaryRedirect,auth.AuthURL(state))
		log.Println(e)
		return
	}
	playlists, e := client.GetPlaylistsForUser(user.ID)
	if e!= nil {
		log.Println(e)
	}
	vars.Set("playlists",playlists.Playlists)
	c.Writer.WriteHeader(200)
	if err := t.Execute(c.Writer, vars, nil); err != nil {
		log.Println(user.ID)
	}
	//c.String(http.StatusOK,user.ID)
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

func VerifyLogin() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		Token := session.Get("Token")
		if Token == nil{
			//need to login
			c.Redirect(http.StatusTemporaryRedirect, auth.AuthURL(state))
			c.AbortWithStatus(http.StatusTemporaryRedirect)
			return
		}
		var token oauth2.Token
		err := json.Unmarshal(Token.([]byte), &token)
		if err != nil {
			session.Set("Token",[]int{})
			_ = session.Save()
			c.Redirect(http.StatusTemporaryRedirect,auth.AuthURL(state))
			c.AbortWithStatus(http.StatusTemporaryRedirect)
			return
		}
		if time.Now().After(token.Expiry) {
			//token has expired
			session.Set("Token",[]int{})
			_ = session.Save()
			c.Redirect(http.StatusTemporaryRedirect,auth.AuthURL(state))
			c.AbortWithStatus(http.StatusTemporaryRedirect)
			return
		}
		c.Set("Token",token)
		c.Next()
	}
}

