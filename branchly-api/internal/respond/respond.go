package respond

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type errorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type errorResponse struct {
	Error errorPayload `json:"error"`
}

type dataResponse struct {
	Data any `json:"data"`
}

func JSONError(c *gin.Context, status int, code, message string) {
	c.JSON(status, errorResponse{
		Error: errorPayload{Code: code, Message: message},
	})
}

func JSONData(c *gin.Context, status int, data any) {
	c.JSON(status, dataResponse{Data: data})
}

func JSONOK(c *gin.Context, data any) {
	JSONData(c, http.StatusOK, data)
}

func JSONCreated(c *gin.Context, data any) {
	JSONData(c, http.StatusCreated, data)
}
