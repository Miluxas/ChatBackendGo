package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/casbin/casbin"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/miluxas/ChatBackendGo/models"
)

func main() {
	// load the casbin model and policy from files, database is also supported.
	e, _ := casbin.NewEnforcer("authz_model.conf", "authz_policy.csv")
	store := sessions.NewCookieStore([]byte("sessionSuperSecret"))
	router := gin.New()
	router.Use(sessions.Sessions("sessionName", store))

	api := router.Group("/Chat")
	// no authentication endpoints
	{
		api.POST("/login", loginHandler)
	}
	// basic authentication endpoints
	{
		basicAuth := router.Group("/Chat")
		basicAuth.Use(newAuthorizer(e))
		//basicAuth.Use(authz.NewAuthorizer(e))
		basicAuth.Use(checkUserAuthentication())
		{
			basicAuth.GET("/logout", logoutHandler)
			basicAuth.POST("/CreateNewChat", startNewPeerChat)
			basicAuth.POST("/CreateGroupChat", startNewGroupChat)
			basicAuth.POST("/SendMessageToChat", sendMessageToChat)
			basicAuth.POST("/JoinToChat", joinToChat)
			basicAuth.POST("/AddMemberToChat", addMemberToChat)
			basicAuth.POST("/GetChat", getChat)
		}
	}

	router.Run()
}

func checkUserAuthentication(auths ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		user := session.Get("user")
		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user needs to be signed in to access this service"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func newAuthorizer(e *casbin.Enforcer) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !checkPermission(c, e) {
			requirePermission(c)
		}
	}
}

func checkPermission(c *gin.Context, a *casbin.Enforcer) bool {
	session := sessions.Default(c)
	user := session.Get("user")
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user needs to be signed in to access this service"})
		c.Abort()
		return false
	}
	method := c.Request.Method
	path := c.Request.URL.Path

	allowed, err := a.Enforce(user, path, method)
	if err != nil {
		panic(err)
	}

	return allowed
}

func requirePermission(c *gin.Context) {
	c.AbortWithStatus(403)
}

func getUserID(c *gin.Context) string {
	session := sessions.Default(c)
	return fmt.Sprintf("%v", session.Get("user"))
}

/********************************************************************************/
/*	user login 																	*/
/*																				*/
/********************************************************************************/
type user struct {
	Username string `form:"username" json:"username" xml:"username" binding:"required"`
	Password string `form:"password" json:"password" xml:"password" binding:"required"`
}

func loginHandler(c *gin.Context) {
	user := user{}

	if err := c.ShouldBind(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	session := sessions.Default(c)

	userID := models.AuthenticateUser(user.Username, user.Password)

	if strings.Trim(user.Username, " ") == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username can't be empty"})
	}
	if userID == "__" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid auth type"})
	}
	session.Set("user", userID)

	err := session.Save()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate session token"})
		return
	}

	//fmt.Println("login :", c.GetString("ginadmin/user_id"))

	c.JSON(http.StatusOK, gin.H{"message": "authentication successful"})
}

/********************************************************************************/
/*	user logout 																*/
/*																				*/
/********************************************************************************/
func logoutHandler(c *gin.Context) {
	session := sessions.Default(c)
	// this would only be hit if the user was authenticated
	session.Delete("user")

	err := session.Save()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate session token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "successfully logged out"})

}

/********************************************************************************/
/*	start new peer chat															*/
/*																				*/
/********************************************************************************/
type startNewChat struct {
	Title      string `form:"title" json:"title" xml:"title" binding:"required"`
	PeerUserID string `form:"peerUserId" json:"peerUserId" xml:"peerUserId" binding:"required"`
}

func startNewPeerChat(c *gin.Context) {
	newChat := startNewChat{}
	if err := c.ShouldBind(&newChat); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	newChatID := models.StartNewPeerChat(newChat.Title, getUserID(c), newChat.PeerUserID)
	c.JSON(http.StatusCreated, gin.H{"status": http.StatusCreated, "message": "Chat created successfully!", "newChatID": newChatID})
}

/********************************************************************************/
/*	start new group chat														*/
/*																				*/
/********************************************************************************/
type newGroupChat struct {
	Title    string `form:"title" json:"title" xml:"title" binding:"required"`
	ChatType string `form:"chatType" json:"chatType" xml:"chatType" binding:"required"`
}

func startNewGroupChat(c *gin.Context) {
	newGroupChat := newGroupChat{}
	if err := c.ShouldBind(&newGroupChat); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	newChatID := models.StartNewGroupChat(newGroupChat.Title, getUserID(c), newGroupChat.ChatType)
	c.JSON(http.StatusCreated, gin.H{"status": http.StatusCreated, "message": "Chat created successfully!", "newChatID": newChatID})
}

/********************************************************************************/
/*	send message to a chat														*/
/*																				*/
/********************************************************************************/
type newMessage struct {
	ChatID  string `form:"chatId" json:"chatId" xml:"chatId" binding:"required"`
	Message string `form:"message" json:"message" xml:"message" binding:"required"`
}

func sendMessageToChat(c *gin.Context) {
	newMessage := newMessage{}
	if err := c.ShouldBind(&newMessage); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	newID, err := models.SendMessageToChat(newMessage.ChatID, getUserID(c), newMessage.Message)
	if err == nil {
		c.JSON(http.StatusCreated, gin.H{"status": http.StatusCreated, "message": "Message created successfully!", "newId": newID})
		return
	}
	{
		c.JSON(http.StatusCreated, gin.H{"status": http.StatusNotFound, "message": err.Error()})
	}
}

/********************************************************************************/
/*	join to a chat																*/
/*																				*/
/********************************************************************************/
type chat struct {
	ChatID string `form:"chatId" json:"chatId" xml:"chatId" binding:"required"`
}

func joinToChat(c *gin.Context) {
	chat := chat{}
	if err := c.ShouldBind(&chat); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	newID, err := models.JoinToChat(chat.ChatID, getUserID(c))
	if err == nil {
		c.JSON(http.StatusCreated, gin.H{"status": http.StatusCreated, "message": "joined to chat successfully!", "newId": newID})
		return
	}
	{
		c.JSON(http.StatusCreated, gin.H{"status": http.StatusNotFound, "message": err.Error()})
	}
}

/********************************************************************************/
/*	add a new member to a chat													*/
/*																				*/
/********************************************************************************/
type newMember struct {
	ChatID string `form:"chatId" json:"chatId" xml:"chatId" binding:"required"`
	UserID string `form:"userId" json:"userId" xml:"userId" binding:"required"`
}

func addMemberToChat(c *gin.Context) {
	newMember := newMember{}
	if err := c.ShouldBind(&newMember); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	newID, err := models.AddOtherUserToChat(newMember.ChatID, getUserID(c), newMember.UserID)
	if err == nil {
		c.JSON(http.StatusCreated, gin.H{"status": http.StatusCreated, "message": "new member added successfully!", "newId": newID})
		return
	}
	{
		c.JSON(http.StatusCreated, gin.H{"status": http.StatusNotFound, "message": err.Error()})
	}
}

/********************************************************************************/
/*	get the chat as a json string												*/
/*																				*/
/********************************************************************************/
func getChat(c *gin.Context) {
	chat := chat{}
	if err := c.ShouldBind(&chat); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	jChat, err := models.GetChat(chat.ChatID, getUserID(c))
	if err == nil {
		c.JSON(http.StatusCreated, gin.H{"status": http.StatusCreated, "message": "Chat loaded successfully!", "jChat": jChat})
		return
	}
	{
		c.JSON(http.StatusCreated, gin.H{"status": http.StatusNotFound, "message": err.Error()})
	}
}
