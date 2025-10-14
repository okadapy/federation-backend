package files

import (
	"federation-backend/app/db/models"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Controller struct {
	service *Service
}

// UploadFile handles single file upload
func (c *Controller) UploadFile(ctx *gin.Context) {
	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	file, err := c.service.SaveFile(fileHeader)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"status": "success",
		"file":   file,
	})
}

// UploadMultipleFiles handles multiple file uploads
func (c *Controller) UploadMultipleFiles(ctx *gin.Context) {
	form, err := ctx.MultipartForm()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "multipart form required"})
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "no files provided"})
		return
	}

	var uploadedFiles []interface{}
	var errors []string

	for _, fileHeader := range files {
		file, err := c.service.SaveFile(fileHeader)
		if err != nil {
			errors = append(errors, err.Error())
			continue
		}
		uploadedFiles = append(uploadedFiles, file)
	}

	response := gin.H{
		"uploaded": uploadedFiles,
		"status":   "partial success",
	}

	if len(errors) > 0 {
		response["errors"] = errors
		if len(uploadedFiles) == 0 {
			response["status"] = "failed"
			ctx.JSON(http.StatusInternalServerError, response)
			return
		}
	} else {
		response["status"] = "success"
	}

	ctx.JSON(http.StatusCreated, response)
}

// DeleteFile deletes a file by filename
func (c *Controller) DeleteFile(ctx *gin.Context) {
	filename := ctx.Param("filename")
	if filename == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "filename is required"})
		return
	}

	if err := c.service.DeleteFile(filename); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success"})
}

// GetFileInfo gets file information from database by ID
func (c *Controller) GetFileInfo(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var fileModel models.File // You'll need to define this or use your existing File model
	if err := c.service.db.First(&fileModel, uint(id)).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	ctx.JSON(http.StatusOK, fileModel)
}

// GetAllFiles returns all files from database
func (c *Controller) GetAllFiles(ctx *gin.Context) {
	var files []models.File // You'll need to define this or use your existing File model
	if err := c.service.db.Find(&files).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, files)
}

// CheckFileExists checks if a file exists
func (c *Controller) CheckFileExists(ctx *gin.Context) {
	filename := ctx.Param("filename")
	if filename == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "filename is required"})
		return
	}

	filePath := c.service.GetFilePath(filename)

	// Security check
	if filepath.Base(filename) != filename {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid filename"})
		return
	}

	if _, err := filepath.EvalSymlinks(filePath); err != nil {
		ctx.JSON(http.StatusOK, gin.H{"exists": false})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"exists": true})
}

// GetStorageInfo returns storage information
func (c *Controller) GetStorageInfo(ctx *gin.Context) {
	// You can implement storage statistics here
	// For example: total files, storage usage, etc.
	ctx.JSON(http.StatusOK, gin.H{
		"storage_path": c.service.storagePath,
		"status":       "active",
	})
}

func NewController(db *gorm.DB, storagePath string) (*Controller, error) {
	service, err := NewService(db, storagePath)
	if err != nil {
		return nil, err
	}

	return &Controller{
		service: service,
	}, nil
}
