package main

import (
	"federation-backend/app/api/document"
	files "federation-backend/app/api/file"
	galleryItem "federation-backend/app/api/gallery-item"
	"federation-backend/app/api/match"
	"federation-backend/app/api/news"
	"federation-backend/app/api/shared/crud"
	"federation-backend/app/api/team"
	"federation-backend/app/config"
	"federation-backend/app/db/models"
	"federation-backend/app/interfaces"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	swagger "github.com/swaggo/gin-swagger"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// @title           Federation Backend API
// @version         1.0
// @description     API для Federation Backend
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.example.com/support
// @contact.email  support@example.com

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

func main() {
	config.Init()
	var app = gin.Default()
	var logger = log.Default()
	var db, dbErr = gorm.Open(mysql.Open("root:root@tcp(db:3306)/federation?parseTime=true"), &gorm.Config{})

	app.Use(CORSMiddleware())

	if dbErr != nil {
		panic(dbErr)
	}

	var fileService, fsrvErr = files.NewService(db, config.App.FileStoragePath)

	if fsrvErr != nil {
		panic(fsrvErr)
	}

	if err := db.AutoMigrate(
		models.User{},
		models.CallBack{},
		models.Match{},
		models.File{},
		models.GalleryItem{},
		models.News{},
		models.Chapter{},
		models.Team{},
		models.Document{},
	); err != nil {
		logger.Fatal(err)
	}

	var api = app.Group("/api")

	routerController := map[interfaces.Controller]*gin.RouterGroup{
		crud.NewCrudController[models.User](db, logger):     api.Group("/user"),
		crud.NewCrudController[models.CallBack](db, logger): api.Group("/callback"),
		galleryItem.NewController(db, fileService):          api.Group("/gallery"),
		news.NewController(db, fileService):                 api.Group("/news"),
		crud.NewCrudController[models.Chapter](db, logger):  api.Group("/chapter"),
		team.NewController(db, fileService):                 api.Group("/team"),
		match.NewController(db, logger):                     api.Group("/match"),
		document.NewController(db, fileService):             api.Group("/document"),
	}

	fileController, err := files.NewController(db, config.App.FileStoragePath)
	if err != nil {
		logger.Fatal(err)
	}

	fileGroup := api.Group("/files")
	{
		fileGroup.DELETE("/:filename", fileController.DeleteFile)
	}

	for controller, router := range routerController {

		fmt.Println("\n\ninitializing router for ", router.BasePath())
		interfaces.RegisterRoutes(controller, router)
		fmt.Println("router for " + router.BasePath() + " is initialized\n\n")
	}

	api.GET("/swagger/*any", swagger.WrapHandler(swaggerFiles.Handler))
	api.Static("/files", "./files")

	if exc := app.Run(":8080"); exc != nil {
		panic(exc)
	}
}
