package match

import (
	"federation-backend/app/api/shared/crud"
	"federation-backend/app/db/models"
	"federation-backend/app/db/models/enums"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Controller struct {
	db    *gorm.DB
	match *crud.Service[models.Match]
	teams *crud.Service[models.Team]
}

func (c Controller) Delete(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return // ← Added return
	}

	if err := c.match.Delete(ctx.Request.Context(), uint(id)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return // ← Added return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Match deleted"})
}

func (c Controller) Get(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return // ← Added return
	}

	var match *models.Match
	// Improved preloading
	result := c.db.Preload("Teams").Preload("Teams.TeamLogo").Where("id = ?", id).First(&match)
	if result.Error != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"message": "match not found"})
		return // ← Added return
	}

	ctx.JSON(http.StatusOK, match)
}

func (c Controller) GetAll(ctx *gin.Context) {
	var matches []*models.Match
	result := c.db.Preload("Teams").Preload("Teams.TeamLogo").Find(&matches)
	if result.Error != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return // ← Added return
	}

	if len(matches) == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{"message": "no matches found"})
		return // ← Added return
	}

	ctx.JSON(http.StatusOK, matches)
}

type CreateMatchDTO struct {
	League  string    `json:"league" binding:"required"`
	Date    string    `json:"date" binding:"required"`
	Sex     enums.Sex `json:"sex" binding:"required"`
	TeamIDs []uint    `json:"team_ids" binding:"required"`
}

func (c Controller) Create(ctx *gin.Context) {
	var dto CreateMatchDTO
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return // ← Added return
	}

	if len(dto.TeamIDs) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "must provide team_ids"})
		return // ← Added return
	}

	var item models.Match
	var date time.Time
	var err error
	if date, err = c.parseDate(dto.Date); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	item.Date = date
	item.League = dto.League
	item.Sex = dto.Sex
	log.Println(dto.TeamIDs)

	// Create match first
	if err := c.match.Create(ctx, &item); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return // ← Added return
	}

	// More efficient team association using WHERE IN
	var teams []*models.Team
	if err := c.teams.Db.Where("id IN ?", dto.TeamIDs).Find(&teams).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return // ← Added return
	}

	// Associate teams with match
	if err := c.match.Db.Model(&item).Association("Teams").Append(teams); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return // ← Added return
	}

	log.Println(teams)

	ctx.JSON(http.StatusCreated, gin.H{"message": "match created", "id": item.Id})
}

type UpdateMatchDTO struct {
	League  *string    `json:"league"`
	Date    *string    `json:"date"`
	Sex     *enums.Sex `json:"sex"`
	TeamIDs []uint     `json:"team_ids"`
}

func (c Controller) Update(ctx *gin.Context) {
	var dto UpdateMatchDTO

	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return // ← Added return
	}

	if err := ctx.ShouldBindJSON(&dto); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return // ← Added return
	}

	item, err := c.match.Get(ctx.Request.Context(), uint(id))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return // ← Added return
	}

	if dto.League != nil {
		item.League = *dto.League
	}

	if dto.Date != nil {
		var date time.Time
		if date, err = c.parseDate(*dto.Date); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		item.Date = date
	}

	if dto.Sex != nil {
		item.Sex = *dto.Sex
	}

	if err := c.match.Db.Save(item).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return // ← Added return
	}

	if dto.TeamIDs != nil {
		var teams []*models.Team
		if err := c.teams.Db.Where("id IN ?", dto.TeamIDs).Find(&teams).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return // ← Added return
		}
		if err := c.match.Db.Model(item).Association("Teams").Replace(teams); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return // ← Added return
		}
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "match updated"})
}

func NewController(db *gorm.DB, logger *log.Logger) *Controller {
	return &Controller{
		db:    db,
		match: crud.NewCrudService[models.Match](db, logger),
		teams: crud.NewCrudService[models.Team](db, logger),
	}
}

func (c *Controller) parseDate(date string) (time.Time, error) {
	// Try parsing as timestamp first
	if timestamp, err := strconv.ParseInt(date, 10, 64); err == nil {
		return time.Unix(timestamp, 0), nil
	}

	// Try parsing as RFC3339 or other date formats
	formats := []string{
		time.RFC3339,
		"2006-01-02",
		"2006-01-02 15:04:05",
	}

	for _, format := range formats {
		if parsed, err := time.Parse(format, date); err == nil {
			return parsed, nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid date format: %s", date)
}
