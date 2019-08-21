package main

import (
	"net/http"
	"strings"

	"github.com/casbin/casbin"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/miluxas/ChatBackendGo/models"
)

//User a user info that read from post
type User struct {
	Username string `form:"username" json:"username" xml:"username" binding:"required"`
	Password string `form:"password" json:"password" xml:"password" binding:"required"`
}

func main() {
	// load the casbin model and policy from files, database is also supported.
	e, _ := casbin.NewEnforcer("authz_model.conf", "authz_policy.csv")
	store := sessions.NewCookieStore([]byte("sessionSuperSecret"))
	router := gin.New()
	router.Use(sessions.Sessions("sessionName", store))

	api := router.Group("/Chat")
	// no authentication endpoints
	{
		api.GET("/login", loginHandler)
	}
	// basic authentication endpoints
	{
		basicAuth := router.Group("/Chat")
		basicAuth.Use(newAuthorizer(e))
		basicAuth.Use(checkUserAuthentication())
		{
			basicAuth.GET("/logout", logoutHandler)
			basicAuth.GET("/CreateNewChat/:title", startNewPeerChat)
		}
	}

	router.Run()
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

// RequirePermission returns the 403 Forbidden to the client
func requirePermission(c *gin.Context) {
	c.AbortWithStatus(403)
}

func startNewPeerChat(c *gin.Context) {
	models.StartNewPeerChat("ffdfdf", c.Param("title"), "normal@e.c")
	c.JSON(http.StatusCreated, gin.H{"status": http.StatusCreated, "message": "Chat created successfully!"})
	//fmt.Println(models.ChatList[len(models.ChatList)-1])
}

func loginHandler(c *gin.Context) {
	user := User{
		Username: "admin",
		Password: "admin",
	}
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

	c.JSON(http.StatusOK, gin.H{"message": "authentication successful"})
}

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

//CheckUserAuthentication a func
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
