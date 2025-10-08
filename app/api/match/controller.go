package match

import (
	"federation-backend/app/api/shared/crud"
	"federation-backend/app/db/models"
	"federation-backend/app/db/models/enums"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Controller struct {
	*crud.Controller[models.Match]
	match *crud.Service[models.Match]
	teams *crud.Service[models.Team]
}

type CreateMatchDTO struct {
	League  string    `json:"league" binding:"required"`
	Date    time.Time `json:"date" binding:"required"`
	Sex     enums.Sex `json:"sex" binding:"required"`
	TeamIDs []uint    `json:"team_ids" binding:"required"`
}

func (c Controller) Create(ctx *gin.Context) {
	var dto CreateMatchDTO
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	if dto.TeamIDs == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "must provide team_ids"})
	}

	var item models.Match
	item.Date = dto.Date
	item.League = dto.League
	item.Sex = dto.Sex

	if err := c.match.Create(ctx, &item); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	for _, id := range dto.TeamIDs {
		team, err := c.teams.Get(ctx.Request.Context(), id)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}

		item.Teams = append(item.Teams, team)
	}

	c.match.Db.Save(&item)

	ctx.JSON(http.StatusCreated, gin.H{"message": "match created"})
}

type UpdateMatchDTO struct {
	League  *string    `json:"league" `
	Date    *time.Time `json:"date" `
	Sex     *enums.Sex `json:"sex"`
	TeamIDs []uint     `json:"teams" `
}

func (c Controller) Update(ctx *gin.Context) {
	var dto UpdateMatchDTO

	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
	}

	if err := ctx.ShouldBindJSON(&dto); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	item, err := c.match.Get(ctx.Request.Context(), uint(id))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	if dto.TeamIDs != nil {
		var teams []*models.Team
		for _, id := range dto.TeamIDs {
			team, err := c.teams.Get(ctx.Request.Context(), id)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			teams = append(teams, team)
		}
		item.Teams = teams
	}
	if dto.League != nil {
		item.League = *dto.League
	}

	if dto.Date != nil {
		item.Date = *dto.Date
	}

	if dto.Sex != nil {
		item.Sex = *dto.Sex
	}

	c.match.Db.Save(&item)
	ctx.JSON(http.StatusOK, gin.H{"message": "match updated"})
}

func NewController(db *gorm.DB, logger *log.Logger) *Controller {
	return &Controller{
		Controller: crud.NewCrudController[models.Match](db, logger),
		match:      crud.NewCrudService[models.Match](db, logger),
		teams:      crud.NewCrudService[models.Team](db, logger),
	}

}
