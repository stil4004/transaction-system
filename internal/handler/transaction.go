package handler

import (
	"bs"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type GetAllList struct {
	Data []bs.Answer `json:"data"`
}

func (h *Handler) AddToWallet(c *gin.Context) {
	var input bs.Request

	if err := c.BindJSON(&input); err != nil {
		NewErrorResponse(c, http.StatusBadRequest, "Problem with request")
		fmt.Println(err)
		return
	}

	err := h.services.Transactions.AddSum(input)	

	if err != nil {
		NewErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, statusResponse{Status: "ok"})
}

func (h *Handler) TransferTo(c *gin.Context) {
	var input bs.Transfer

	if err := c.BindJSON(&input); err != nil {
		NewErrorResponse(c, http.StatusBadRequest, "Problem with request")
		fmt.Println(err)
		return
	}

	err := h.services.Transactions.TransferTo(input)

	if err != nil {
		NewErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, statusResponse{Status: "ok"})
}

func (h *Handler) TakeFromWallet(c *gin.Context) {
	var input bs.Request

	if err := c.BindJSON(&input); err != nil {
		NewErrorResponse(c, http.StatusBadRequest, "Problem with request")
		return
	}

	err := h.services.Transactions.TakeOff(input)
	if err != nil {
		NewErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, statusResponse{Status: "ok"})
}

func (h *Handler) GetBalance(c *gin.Context) {
	list, err := h.services.Transactions.GetBalance()

	if err != nil {
		NewErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, GetAllList{
		Data: list,
	})
}
