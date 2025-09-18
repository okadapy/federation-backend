package main

import (
	files "federation-backend/app/api/file"
	gallery_item "federation-backend/app/api/gallery-item"
	"federation-backend/app/api/shared/crud"
	"federation-backend/app/config"
	"federation-backend/app/db/models"
	"federation-backend/app/interfaces"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	config.Init()
	var app = gin.Default()
	var db, dbErr = gorm.Open(sqlite.Open("test.db"), &gorm.Config{})

	if dbErr != nil {
		panic(dbErr)
	}

	var fileService, fsrvErr = files.NewService(db, config.App.FileStoragePath)

	if fsrvErr != nil {
		panic(fsrvErr)
	}

	db.AutoMigrate(
		models.User{},
		models.CallBack{},
		models.Match{},
		models.File{},
		models.GalleryItem{},
	)

	routerController := map[interfaces.Controller]*gin.RouterGroup{
		crud.NewCrudController[models.User](db):     app.Group("/user"),
		crud.NewCrudController[models.CallBack](db): app.Group("/callback"),
		crud.NewCrudController[models.Match](db):    app.Group("/match"),
		gallery_item.NewController(db, fileService): app.Group("/gallery"),
	}

	for controller, router := range routerController {
		interfaces.RegisterRoutes(controller, router)
		fmt.Println("router for " + router.BasePath() + " is initialized")
	}

	if exc := app.Run(config.Server.GetHostURL()); exc != nil {
		panic(exc)
	}
}
