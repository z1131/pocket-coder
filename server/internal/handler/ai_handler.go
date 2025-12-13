package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"pocket-coder-server/internal/service"
	"pocket-coder-server/pkg/response"
)

type AIHandler struct {
	aiService *service.AIService
}

func NewAIHandler(aiService *service.AIService) *AIHandler {
	return &AIHandler{aiService: aiService}
}

// GenerateCommand 处理命令生成请求
func (h *AIHandler) GenerateCommand(c *gin.Context) {
	var req service.GenerateCommandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Prompt == "" {
		response.Fail(c, http.StatusBadRequest, "Prompt cannot be empty")
		return
	}

	result, err := h.aiService.GenerateCommand(c.Request.Context(), &req)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, result)
}
