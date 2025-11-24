package main

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	r.ParseMultipartForm(10 << 20) // 10 MB max mem

	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse thumbnail", err)
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")

	mediaType, _, _ := mime.ParseMediaType(contentType)
	if mediaType != "image/jpeg" && mediaType != "image/png" {
		respondWithError(w, http.StatusBadRequest, "Invalid thumbnail type", errors.New("only JPEG or PNG allowed"))
		return
	}

	// data, err := io.ReadAll(file)
	// if err != nil {
	// 	respondWithError(w, http.StatusInternalServerError, "Could not read file", err)
	// 	return
	// }

	video, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Could not find video", err)
		return
	}
	if video.UserID != userID {
		respondWithJSON(w, http.StatusUnauthorized, "Not owner")
		return
	}

	// Base64 approach
	// encodedImgData := base64.StdEncoding.EncodeToString(data)
	// dataURL := fmt.Sprintf("data:%s;base64,%s", contentType, encodedImgData)
	// video.ThumbnailURL = &dataURL

	// Save to disk approach
	contentTypeParts := strings.Split(contentType, "/")
	if len(contentTypeParts) == 0 {
		respondWithJSON(w, http.StatusBadRequest, "Invalid content type")
		return
	}
	path := filepath.Join(cfg.assetsRoot, fmt.Sprintf("%s.%s", videoIDString, contentTypeParts[len(contentTypeParts)-1]))
	assetFile, err := os.Create(path)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not create file", err)
		return
	}
	_, err = io.Copy(assetFile, file)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not write file", err)
		return
	}

	thumbURL := fmt.Sprintf("http://localhost:%s/%s", cfg.port, path)
	video.ThumbnailURL = &thumbURL

	err = cfg.db.UpdateVideo(video)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not save to DB", err)
		return
	}

	respondWithJSON(w, http.StatusOK, video)
}
