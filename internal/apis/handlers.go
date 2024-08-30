package api

import (
	"blob_store_service/internal/blobstore"
	"blob_store_service/pkg/middlewares"
	"mime/multipart"
	"strconv"
	"strings"
	"sync"

	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	bs *blobstore.BlobStore
}

func NewHandler(bs *blobstore.BlobStore) *Handler {
	return &Handler{bs: bs}
}

func (h *Handler) handleSaveBlob(fileHeader *multipart.FileHeader, wg *sync.WaitGroup, blobsChan chan *blobstore.Blob, errChan chan error, blob *blobstore.Blob) {
	defer wg.Done()

	file, err := fileHeader.Open()
	if err != nil {
		errChan <- err
		return
	}
	defer file.Close()
	blob.FileName = fileHeader.Filename
	blob.ContentType = fileHeader.Header.Get("Content-Type")
	savedBlob, err := h.bs.SaveBlob(file, blob)
	if err != nil {
		errChan <- err
		return
	}
	blobsChan <- savedBlob
}
func (h *Handler) UploadBlobs(c *gin.Context) {
	currentUser := middlewares.GetUserAuthFromContext(c)
	form, err := c.MultipartForm()

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Unable to parse form",
		})
		return
	}

	files := form.File["files"]
	amountFiles := len(files)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "No files uploaded",
		})
		return
	}

	blobsChan := make(chan *blobstore.Blob, amountFiles)
	errChan := make(chan error, amountFiles)
	wg := sync.WaitGroup{}
	var generalBlob blobstore.Blob
	if err := c.Bind(&generalBlob); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	generalBlob.OwnerID = currentUser.UserId

	for _, fileHeader := range files {
		wg.Add(1)
		go h.handleSaveBlob(fileHeader, &wg, blobsChan, errChan, generalBlob.Clone())
	}

	wg.Wait()
	close(blobsChan)
	close(errChan)
	var blobs []*blobstore.Blob
	var errors []string

	for blob := range blobsChan {
		blobs = append(blobs, blob)
	}
	for err := range errChan {
		errors = append(errors, err.Error())
	}

	if len(errors) > 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": strings.Join(errors, ", "),
		})
		return
	}
	c.JSON(http.StatusOK, blobs)
}
func (h *Handler) GetBlobInfo(c *gin.Context) {
	id := c.Param("id")
	blob, err := h.bs.GetBlob(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, blob)
}
func (h *Handler) GetListBlobInfoWithPagination(c *gin.Context) {
	currentUser := middlewares.GetUserAuthFromContext(c)
	pageParam, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		pageParam = 1
	}
	filter := map[string]interface{}{
		"owner_id": currentUser.UserId,
	}

	h.bs.GetListBlobWithPagination(10, pageParam, filter)
}
func (h *Handler) DownloadBlob(c *gin.Context) {
	id := c.Param("id")
	blob, err := h.bs.GetBlob(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": err.Error(),
		})
		return
	}
	c.Header("Content-Type", blob.ContentType)
	c.Header("Content-Disposition", "attachment; filename="+blob.FileName)
	c.Header("Content-Length", fmt.Sprintf("%d", blob.Size))

	err = h.bs.StreamBlob(id, c.Writer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}
}

func (h *Handler) DeleteBlob(c *gin.Context) {
	id := c.Param("id")
	err := h.bs.DeleteBlob(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "deleted",
	})
}
