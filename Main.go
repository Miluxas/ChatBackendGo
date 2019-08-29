package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/casbin/casbin"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"

	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/miluxas/ChatBackendGo/models"
)

func main() {
	// load the casbin model and policy from files, database is also supported.
	e, _ := casbin.NewEnforcer("authz_model.conf", "authz_policy.csv")
	router := gin.New()
	store := cookie.NewStore([]byte("secret"))
	router.Use(sessions.Sessions("mysession", store))
	router.LoadHTMLGlob("*.html")
	router.Use(static.Serve("/js/", static.LocalFile("./js", true)))
	router.GET("/", indexHandler)
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
			basicAuth.POST("/LeaveFromChat", leaveFromChat)
			basicAuth.POST("/GetChat", getChat)
			basicAuth.POST("/GetChatList", getChatList)
			basicAuth.GET("/Stream", stream)
		}
	}

	router.Run(":3031")
	//http.ListenAndServe(":3000", nil)
}

func indexHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", nil)
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
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user needs to be signed in to access this service(get user from session)"})
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
	userI := models.AuthenticateUser(user.Username, user.Password)

	if strings.Trim(user.Username, " ") == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username can't be empty"})
	}
	if userI.ID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid auth type"})
	}
	session.Set("user", userI.ID)
	err := session.Save()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate session token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "authentication successful", "usr": userI})
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
	Title      string `form:"title" json:"Title" xml:"title" binding:"required"`
	PeerUserID string `form:"peerUserId" json:"peerUserId" xml:"peerUserId" binding:"required"`
	ID         string
}

func startNewPeerChat(c *gin.Context) {
	newChat := startNewChat{}
	if err := c.ShouldBind(&newChat); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newChatID := models.StartNewPeerChat(newChat.Title, getUserID(c), newChat.PeerUserID)
	newChat.ID = newChatID
	newAlert := models.Alert{
		AlertType: "NewChatCreated",
		Data:      newChat,
	}
	models.SendAlertToMember(newChatID, newAlert)
	c.JSON(http.StatusCreated, gin.H{"status": http.StatusCreated, "message": "Chat created successfully!", "newChatID": newChatID})
}

/********************************************************************************/
/*	start new group chat														*/
/*																				*/
/********************************************************************************/
type newGroupChat struct {
	Title    string `form:"title" json:"Title" xml:"title" binding:"required"`
	ChatType string `form:"chatType" json:"chatType" xml:"chatType" binding:"required"`
	ID       string
}

func startNewGroupChat(c *gin.Context) {
	newGroupChat := newGroupChat{}
	if err := c.ShouldBind(&newGroupChat); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	newChatID := models.StartNewGroupChat(newGroupChat.Title, getUserID(c), newGroupChat.ChatType)
	newGroupChat.ID = newChatID
	newAlert := models.Alert{
		AlertType: "NewChatCreated",
		Data:      newGroupChat,
	}
	models.SendAlertToMember(newChatID, newAlert)

	c.JSON(http.StatusCreated, gin.H{"status": http.StatusCreated, "message": "Chat created successfully!", "newChatID": newChatID})
}

/********************************************************************************/
/*	send message to a chat														*/
/*																				*/
/********************************************************************************/
type newMessage struct {
	ChatID  string `form:"chatId" json:"chatId" xml:"chatId" binding:"required"`
	Message string `form:"message" json:"Content" xml:"message" binding:"required"`
	OwnerID string
}

func sendMessageToChat(c *gin.Context) {
	newMessage := newMessage{}
	if err := c.ShouldBind(&newMessage); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	newID, err := models.SendMessageToChat(newMessage.ChatID, getUserID(c), newMessage.Message)
	newMessage.OwnerID = getUserID(c)
	if err == nil {
		newAlert := models.Alert{
			AlertType: "NewMessageAdded",
			Data:      newMessage,
		}
		models.SendAlertToMember(newMessage.ChatID, newAlert)
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
	ChatID   string `form:"chatId" json:"chatId" xml:"chatId" binding:"required"`
	MemberID string
}

func joinToChat(c *gin.Context) {
	chat := chat{}
	if err := c.ShouldBind(&chat); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	newID, err := models.JoinToChat(chat.ChatID, getUserID(c))
	if err == nil {
		newAlert := models.Alert{
			AlertType: "JoinedToChat",
			Data:      chat,
		}
		models.SendAlertToMember(chat.ChatID, newAlert)
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
type member struct {
	ChatID string `form:"chatId" json:"ID" xml:"chatId" binding:"required"`
	UserID string `form:"userId" json:"UserID" xml:"userId" binding:"required"`
	Title  string
}

func addMemberToChat(c *gin.Context) {
	member := member{}
	if err := c.ShouldBind(&member); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	title, newID, err := models.AddOtherUserToChat(member.ChatID, getUserID(c), member.UserID)
	if err == nil {
		member.Title = title
		newAlert := models.Alert{
			AlertType: "AddedToChat",
			Data:      member,
		}
		models.SendAlertToOneMember(member.UserID, newAlert)

		newAlert.AlertType = "NewMemberAdded"
		models.SendAlertToMember(member.ChatID, newAlert)

		c.JSON(http.StatusCreated, gin.H{"status": http.StatusCreated, "message": "new member added successfully!", "newId": newID})
		return
	}

	{
		c.JSON(http.StatusCreated, gin.H{"status": http.StatusNotFound, "message": err.Error()})
	}
}

func leaveFromChat(c *gin.Context) {
	chat := chat{}
	if err := c.ShouldBind(&chat); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	leftUserID, leftMemberID, err := models.LeaveChat(chat.ChatID, getUserID(c))
	if err == nil {
		chat.MemberID = leftMemberID
		newAlert := models.Alert{
			AlertType: "MemberLeftChat",
			Data:      chat,
		}
		models.SendAlertToMember(chat.ChatID, newAlert)
		newAlert.AlertType = "LeftChat"
		models.SendAlertToOneMember(leftUserID, newAlert)

		c.JSON(http.StatusCreated, gin.H{"status": http.StatusCreated, "message": "member left chat successfully!"})
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

/********************************************************************************/
/*	get the user chat list as a json string										*/
/*																				*/
/********************************************************************************/
func getChatList(c *gin.Context) {
	jChatList, err := models.GetChatList(getUserID(c))
	if err == nil {
		c.JSON(http.StatusCreated, gin.H{"status": http.StatusCreated, "message": "Chat loaded successfully!", "jChatList": jChatList})
		return
	}
	{
		c.JSON(http.StatusCreated, gin.H{"status": http.StatusNotFound, "message": err.Error()})
	}
}

/********************************************************************************/
/*	get the realtime stream 													*/
/*																				*/
/********************************************************************************/
func stream(c *gin.Context) {
	userID := getUserID(c)
	listener := models.OpenListener(userID)
	defer models.CloseListener(userID, listener)

	clientGone := c.Writer.CloseNotify()

	c.Stream(func(w io.Writer) bool {
		select {
		case <-clientGone:
			return false
		case mes := <-listener:
			//fmt.Println(mes)
			alert := mes.(models.Alert)
			c.SSEvent(alert.AlertType, alert.Data)
			return true
		}
	})
}
