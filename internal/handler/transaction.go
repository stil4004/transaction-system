package handler

import (
	"bs"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type GetAllList struct {
	Data []bs.Answer `json:"data"`
}



type GetWallet struct{
	WalletID uint64 `json:"wallet_id"`
	Currencies []bs.WalletCurrency `json:"currencies"`
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

func (h *Handler) GetBalanceByID(c *gin.Context) {
	walletid, err := strconv.Atoi(c.Param("wid"))
	if err != nil{
		NewErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	currency := c.Param("cur")

	wallet_value, err := h.services.Transactions.GetBalanceByID(uint64(walletid), currency)

	if err != nil {
		NewErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	temp_currency := bs.WalletCurrency{
		Currency: currency,
		Value: wallet_value,
	}

	c.JSON(http.StatusOK, GetWallet{
		WalletID: uint64(walletid),
		Currencies: []bs.WalletCurrency{temp_currency},
	})
}

func (h *Handler) GetAllBalancesByID(c *gin.Context) {
	walletid, err := strconv.Atoi(c.Param("wid"))
	if err != nil{
		NewErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	wallet_list, err := h.services.Transactions.GetAllBalancesByID(uint64(walletid))

	if err != nil {
		NewErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, GetWallet{
		WalletID: uint64(walletid),
		Currencies: wallet_list,
	})
}
