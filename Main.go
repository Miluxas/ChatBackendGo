package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/miluxas/ChatBackendGo/models"
)

func main() {
	// load the casbin model and policy from files, database is also supported.
	//e, _ := casbin.NewEnforcer("authz_model.conf", "authz_policy.csv")
	models.SetCurrentUser("admin@e.c")
	// define your router, and use the Casbin authz middleware.
	// the access that is denied by authz will return HTTP 403 error.
	router := gin.Default()
	//router.Use(authz.NewAuthorizer(e))

	v1 := router.Group("/Chat")
	{
		v1.GET("/CreateNewChat/:title", startNewPeerChat)
	}
	router.Run()

	/*models.StartNewPeerChat("ffdfdf", "title one", "normal@e.c")
	models.SendMessageToChat("ffdfdf", "Hi dear. How is it going? ")
	models.SetCurrentUser("kalim@e.c")
	models.JoinToChat("ffdfdf")
	models.SetCurrentUser("admin@a.c")
	models.AddOtherUserToChat("ffdfdf", "solivan@e.c")*/
	fmt.Println(models.ChatList[0])
}

func startNewPeerChat(c *gin.Context) {
	//fmt.Println(c)
	models.StartNewPeerChat("ffdfdf", c.Param("title"), "normal@e.c")
	c.JSON(http.StatusCreated, gin.H{"status": http.StatusCreated, "message": "Chat created successfully!"})
	fmt.Println(models.ChatList[len(models.ChatList)-1])

}
