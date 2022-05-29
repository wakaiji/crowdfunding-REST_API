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

func (h *userHandler) CheckEmailAvailibility(c *gin.Context) {
	var input user.EmailInput

	err := c.ShouldBindJSON(&input)
	if err != nil {
		errors := helper.FormatValidationError(err)
		errorMessage := gin.H{"errors": errors}

		webResponse := helper.APIResponse("Email checking failed", http.StatusUnprocessableEntity, "error", errorMessage)
		c.JSON(http.StatusUnprocessableEntity, webResponse)
		return
	}

	isEmailAvailable, err := h.service.IsEmailAvailable(input)
	if err != nil {
		errorMessage := gin.H{"errors": "Server error"}

		webResponse := helper.APIResponse("Email checking failed", http.StatusUnprocessableEntity, "error", errorMessage)
		c.JSON(http.StatusUnprocessableEntity, webResponse)
		return
	}

	metaMessage := "Email has been registered"
	if isEmailAvailable {
		metaMessage = "Email is available"
	}

	data := gin.H{
		"is_available": isEmailAvailable,
	}

	webResponse := helper.APIResponse(metaMessage, http.StatusOK, "success", data)
	c.JSON(http.StatusOK, webResponse)
}
