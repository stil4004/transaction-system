package handler

import (
	"bs/internal/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	services *service.Service
}

func NewHandler(services *service.Service) *Handler {
	return &Handler{services: services}
}

func (h *Handler) InitRoutes() *gin.Engine {
	router := gin.New()
	
	api := router.Group("/api")
	{
		api.POST("/invoice", h.AddToWallet)
		api.POST("/withdraw", h.TakeFromWallet)
		api.POST("/transfer", h.TransferTo)
		api.GET("/balance/:wid", h.GetAllBalancesByID)
		api.GET("/balance/:wid/:cur", h.GetBalanceByID)
	}
	
	return router
}
