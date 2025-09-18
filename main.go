package main

import (
	"federation-backend/app/api/shared/crud"
	"federation-backend/app/config"
	"federation-backend/app/db/models"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	config.Init()
	var app = gin.Default()
	var db, err = gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	db.AutoMigrate(
		models.User{},
		models.CallBack{},
		models.Match{},
	)

	var userRouter = app.Group("/user")
	var callbackRouter = app.Group("/callback")
	var matchRouter = app.Group("/match")

	var userController = crud.NewCrudController[models.User](db)
	var matchController = crud.NewCrudController[models.Match](db)
	var callbackController = crud.NewCrudController[models.CallBack](db)

	userController.RegisterRoutes(userRouter)
	matchController.RegisterRoutes(matchRouter)
	callbackController.RegisterRoutes(callbackRouter)

	if exc := app.Run(config.Server.GetHostURL()); exc != nil {
		panic(exc)
	}
}
