package handler

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"scaffolding/internal/service"
	"scaffolding/pkg/response"
)

type ArticleHandler struct {
	svc *service.ArticleService
}

func NewArticleHandler(svc *service.ArticleService) *ArticleHandler {
	return &ArticleHandler{svc: svc}
}

type createArticleReq struct {
	Title   string   `json:"title" binding:"required,max=200"`
	Content string   `json:"content" binding:"required"`
	Tags    []string `json:"tags"`
}

func (h *ArticleHandler) Create(c *gin.Context) {
	var req createArticleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		if very, ok := errors.AsType[validator.ValidationErrors](err); ok {
			response.ParamError(c, very)
			return
		}
		response.Error(c, err)
		return
	}

	article, err := h.svc.Create(c.Request.Context(), req.Title, req.Content, req.Tags)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, article)
}

func (h *ArticleHandler) Get(c *gin.Context) {
	response.Success(c, nil)
}
