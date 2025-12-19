package handler

import (
	"gateway/internal/model"
	"gateway/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type LedgerHandler struct {
	service service.LedgerGatewayService
}

func NewLedgerHandler(s service.LedgerGatewayService) *LedgerHandler {
	if s == nil {
		panic("BudgetHandler requires service")
	}
	return &LedgerHandler{service: s}
}

func (l *LedgerHandler) Register(r *gin.RouterGroup) {
	budget := r.Group("/budget")
	{
		budget.POST("/", l.BudgetAdd)
		budget.GET("/", l.BudgetGet)
		budget.GET("/list", l.BudgetList)
	}
	transaction := r.Group("/transaction")
	{
		transaction.POST("/", l.TransactionAdd)
		transaction.GET("/", l.TransactionGet)
		transaction.GET("/list", l.TransactionList)
	}
}

func (l *LedgerHandler) BudgetAdd(c *gin.Context) {
	var req model.BudgetAdd
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Category == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "category must not be empty"})
		return
	}
	err := l.service.BudgetAdd(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

func (l *LedgerHandler) BudgetGet(c *gin.Context) {
	var req model.BudgetGet
	category := c.Query("category")
	if category == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "category must not be empty"})
		return
	}
	req.Category = category
	resp, err := l.service.BudgetGet(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (l *LedgerHandler) BudgetList(c *gin.Context) {
	resp, err := l.service.BudgetList(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (l *LedgerHandler) TransactionAdd(c *gin.Context) {
	var req model.TrasnactionAdd
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Amount == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "amount must not be empty"})
		return
	}
	if req.Category == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "category must not be empty"})
		return
	}
	if req.Date == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "date must not be empty"})
		return
	}
	if req.Description == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "category must not be empty"})
		return
	}
	resp, err := l.service.TransactionAdd(c.Request.Context(), req)
	if err != nil && err.Error() == "rpc error: code = Internal desc = add transaction: budget exceeded" {
		c.JSON(http.StatusConflict, gin.H{"error": "budget exceeded"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (l *LedgerHandler) TransactionGet(c *gin.Context) {
	var req model.TransactionGet
	idStr := c.Query("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id must not be empty"})
		return
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id must be int64"})
		return
	}
	req.Id = id
	resp, err := l.service.TransactionGet(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (l *LedgerHandler) TransactionList(c *gin.Context) {
	resp, err := l.service.TransactionList(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}
