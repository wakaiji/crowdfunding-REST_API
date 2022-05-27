package handler

import (
	"bwastartup/helper"
	"bwastartup/user"
	"net/http"

	"github.com/gin-gonic/gin"
)

type userHandler struct {
	service user.Service
}

func NewUserHandler(service user.Service) *userHandler {
	return &userHandler{
		service: service,
	}
}

func (h *userHandler) RegisterUser(c *gin.Context) {
	var input user.RegisterUserInput

	err := c.ShouldBindJSON(&input)
	if err != nil {
		errors := helper.FormatValidationError(err)
		errorMessage := gin.H{"errors": errors}

		webResponse := helper.APIResponse("Register account failed", http.StatusUnprocessableEntity, "error", errorMessage)
		c.JSON(http.StatusUnprocessableEntity, webResponse)
		return
	}

	newUser, err := h.service.RegisterUser(input)
	if err != nil {
		webResponse := helper.APIResponse("Register account failed", http.StatusBadRequest, "error", nil)
		c.JSON(http.StatusBadRequest, webResponse)
		return
	}

	formatter := user.FormatUser(newUser, "tokentokentoken")

	webResponse := helper.APIResponse("Success Register", http.StatusOK, "success", formatter)
	c.JSON(http.StatusOK, webResponse)
}

func (h *userHandler) Login(c *gin.Context) {
	var input user.LoginUserInput

	err := c.ShouldBindJSON(&input)
	if err != nil {
		errors := helper.FormatValidationError(err)
		errorMessage := gin.H{"errors": errors}

		webResponse := helper.APIResponse("Login failed", http.StatusUnprocessableEntity, "error", errorMessage)
		c.JSON(http.StatusUnprocessableEntity, webResponse)
		return
	}

	userLogin, err := h.service.LoginUser(input)
	if err != nil {
		errorMessage := gin.H{"error": err.Error()}

		webResponse := helper.APIResponse("Login failed", http.StatusUnprocessableEntity, "error", errorMessage)
		c.JSON(http.StatusUnprocessableEntity, webResponse)
		return
	}

	formatter := user.FormatUser(userLogin, "tokentokentoken")

	webResponse := helper.APIResponse("Success Register", http.StatusOK, "success", formatter)
	c.JSON(http.StatusOK, webResponse)
}
