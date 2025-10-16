// Go equivalent of gin_test.rye
// To run: go run gin_test.go
package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	// Create a Gin router (equivalent to srv: router)
	srv := gin.Default()

	// This one will match /user/john/ and also /user/john/send
	srv.GET("/user/:name/*action", func(c *gin.Context) {
		name := c.Param("name")
		action := c.Param("action")
		c.String(http.StatusOK, name+" is"+action)
	})

	// Exact routes are resolved before param routes
	srv.GET("/users/groups", func(c *gin.Context) {
		c.String(http.StatusOK, "The available groups are [admin, users, guests]")
	})

	// Route to test query parameters
	srv.GET("/search", func(c *gin.Context) {
		query := c.Query("q")
		limit := c.Query("limit")
		result := gin.H{
			"query":   query,
			"limit":   limit,
			"results": []string{"result1", "result2", "result3"},
		}
		c.JSON(http.StatusOK, result)
	})

	// Start the server
	srv.Run(":8080")
}
