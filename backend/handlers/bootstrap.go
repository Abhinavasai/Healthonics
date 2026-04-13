package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type BootstrapHandler struct{}

func NewBootstrapHandler() *BootstrapHandler {
	return &BootstrapHandler{}
}

func (h *BootstrapHandler) Get(c *gin.Context) {
	if _, ok := getClaims(c); !ok {
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"app_name":                   "Healthonyx",
		"api_version":                "v1",
		"shell_mobile_breakpoint_px": 768,
	})
}

