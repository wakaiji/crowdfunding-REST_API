package main

import (
	"bwastartup/auth"
	"bwastartup/campaign"
	"bwastartup/handler"
	"bwastartup/helper"
	"bwastartup/payment"
	"bwastartup/transaction"
	"bwastartup/user"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	webHandler "bwastartup/web/handler"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-contrib/multitemplate"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func init() {

	// IF IN LOCAL -> UN-REMARK COMMAND BELOW!

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

}

func main() {
	MYSQL_USER := os.Getenv("MYSQL_USER")
	MYSQL_PASSWORD := os.Getenv("MYSQL_PASSWORD")
	MYSQL_URL := os.Getenv("MYSQL_URL")
	MYSQL_DB := os.Getenv("MYSQL_DB")
	dsn := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?charset=utf8mb4&parseTime=True&loc=Local", MYSQL_USER, MYSQL_PASSWORD, MYSQL_URL, MYSQL_DB)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err.Error())
	}

	campaignRepository := campaign.NewRepository(db)
	userRepository := user.NewRepository(db)
	transactionRepository := transaction.NewRepository(db)

	userService := user.NewService(userRepository)
	campaignService := campaign.NewService(campaignRepository)
	authService := auth.NewService()
	paymentService := payment.NewService()
	transactionService := transaction.NewService(transactionRepository, campaignRepository, paymentService)

	userHandler := handler.NewUserHandler(userService, authService)
	campaignHandler := handler.NewCampaignHandler(campaignService)
	transactionHandler := handler.NewTransactionHandler(transactionService)

	userWebHandler := webHandler.NewUserHandler(userService)

	router := gin.Default()
	router.Use(CORSMiddleware())

	cookieStore := cookie.NewStore([]byte(auth.SECRET_KEY))
	router.Use(sessions.Sessions("bwastartup", cookieStore))

	router.HTMLRender = loadTemplates("./web/templates")

	router.Static("/css", "./web/assets/css")
	router.Static("/js", "./web/assets/js")
	router.Static("/webfonts", "./web/assets/webfonts")
	router.Static("/images", "./images")

	api := router.Group("/api/v1")

	api.POST("/users", userHandler.RegisterUser)
	api.POST("/sessions", userHandler.Login)
	api.POST("/email_checkers", userHandler.CheckEmailAvailibility)
	api.POST("/avatars", AuthMiddleware(authService, userService), userHandler.UploadAvatar)
	api.GET("/users/fetch", AuthMiddleware(authService, userService), userHandler.FetchUser)

	api.GET("/campaigns", campaignHandler.GetCampaigns)
	api.GET("/campaigns/:id", campaignHandler.GetCampaign)
	api.POST("/campaigns", AuthMiddleware(authService, userService), campaignHandler.CreateCampaign)
	api.PUT("/campaigns/:id", AuthMiddleware(authService, userService), campaignHandler.UpdateCampaign)

	api.POST("/campaign-images", AuthMiddleware(authService, userService), campaignHandler.UploadImage)

	api.GET("/campaigns/:id/transactions", AuthMiddleware(authService, userService), transactionHandler.GetCampaignTransaction)
	api.GET("/transactions", AuthMiddleware(authService, userService), transactionHandler.GetUserTransaction)
	api.POST("/transactions", AuthMiddleware(authService, userService), transactionHandler.CreateTransaction)
	api.POST("/transactions/notification", transactionHandler.GetNotification)

	router.GET("/users", authAdminMiddleware(), userWebHandler.Index)
	router.GET("/users/new", authAdminMiddleware(), userWebHandler.New)
	router.POST("/users", authAdminMiddleware(), userWebHandler.Create)
	router.GET("/users/edit/:id", authAdminMiddleware(), userWebHandler.Edit)
	router.POST("/users/update/:id", authAdminMiddleware(), userWebHandler.Update)
	router.GET("/users/avatar/:id", authAdminMiddleware(), userWebHandler.NewAvatar)
	router.POST("/users/avatar/:id", authAdminMiddleware(), userWebHandler.CreateAvatar)

	campaignWebHandler := webHandler.NewCampaignHandler(campaignService, userService)
	router.GET("/campaigns", authAdminMiddleware(), campaignWebHandler.Index)
	router.GET("/campaigns/new", authAdminMiddleware(), campaignWebHandler.New)
	router.POST("/campaigns", authAdminMiddleware(), campaignWebHandler.Create)
	router.GET("/campaigns/image/:id", authAdminMiddleware(), campaignWebHandler.NewImage)
	router.POST("/campaigns/image/:id", authAdminMiddleware(), campaignWebHandler.CreateImage)
	router.GET("/campaigns/edit/:id", authAdminMiddleware(), campaignWebHandler.Edit)
	router.POST("/campaigns/update/:id", authAdminMiddleware(), campaignWebHandler.Update)
	router.GET("/campaigns/show/:id", authAdminMiddleware(), campaignWebHandler.Show)

	transactionWebHandler := webHandler.NewTransactionHandler(transactionService)
	router.GET("/transactions", authAdminMiddleware(), transactionWebHandler.Index)

	sessionWebHandler := webHandler.NewSession(userService)
	router.GET("/login", sessionWebHandler.New)
	router.POST("/session", sessionWebHandler.Create)
	router.GET("/logout", sessionWebHandler.Destroy)

	router.Run()
}

func AuthMiddleware(authService auth.Service, userService user.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		if !strings.Contains(authHeader, "Bearer") {
			response := helper.APIResponse("Unauthorized", http.StatusUnauthorized, "error", nil)
			c.AbortWithStatusJSON(http.StatusUnauthorized, response)
			return
		}

		var tokenString string
		arrayToken := strings.Split(authHeader, " ")

		if len(arrayToken) == 2 {
			tokenString = arrayToken[1]
		}

		token, err := authService.ValidateToken(tokenString)
		if err != nil {
			response := helper.APIResponse("Unauthorized", http.StatusUnauthorized, "error", nil)
			c.AbortWithStatusJSON(http.StatusUnauthorized, response)
			return
		}

		claim, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			response := helper.APIResponse("Unauthorized", http.StatusUnauthorized, "error", nil)
			c.AbortWithStatusJSON(http.StatusUnauthorized, response)
			return
		}

		userID := int(claim["user_id"].(float64))

		user, err := userService.GetUserById(userID)
		if err != nil {
			response := helper.APIResponse("Unauthorized", http.StatusUnauthorized, "error", nil)
			c.AbortWithStatusJSON(http.StatusUnauthorized, response)
			return
		}

		if user.ID == 0 {
			response := helper.APIResponse("Unauthorized", http.StatusUnauthorized, "error", nil)
			c.AbortWithStatusJSON(http.StatusUnauthorized, response)
			return
		}

		c.Set("currentUser", user)
	}
}

func loadTemplates(templatesDir string) multitemplate.Renderer {
	r := multitemplate.NewRenderer()

	layouts, err := filepath.Glob(templatesDir + "/layouts/*")
	if err != nil {
		panic(err.Error())
	}

	includes, err := filepath.Glob(templatesDir + "/**/*")
	if err != nil {
		panic(err.Error())
	}

	for _, include := range includes {
		layoutCopy := make([]string, len(layouts))
		copy(layoutCopy, layouts)
		files := append(layoutCopy, include)
		r.AddFromFiles(filepath.Base(include), files...)
	}

	return r
}

func authAdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)

		userIDSession := session.Get("userID")

		if userIDSession == nil {
			c.Redirect(http.StatusFound, "/login")
			return
		}
	}
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST,HEAD,PATCH, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
