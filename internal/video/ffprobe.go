// Package video contains any video specific functionality
package video

import (
	"bytes"
	"encoding/json"
	"errors"
	"math"
	"os/exec"
)

type FFProbeStream struct {
	CodecType string `json:"codec_type"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
}

type FFProbeOutput struct {
	Streams []FFProbeStream `json:"streams"`
}

func almostEqual(a, b, tolerance float64) bool {
	return math.Abs(a-b) <= tolerance
}

const (
	ar16To9 = float64(16) / float64(9)
	ar9To16 = float64(9) / float64(16)
)

func GetVideoAspectRatio(filePath string) (string, error) {
	ffprobeCmd := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filePath)

	var output bytes.Buffer
	ffprobeCmd.Stdout = &output

	err := ffprobeCmd.Run()
	if err != nil {
		return "", err
	}

	var data FFProbeOutput
	decoder := json.NewDecoder(&output)
	err = decoder.Decode(&data)
	if err != nil {
		return "", err
	}

	var videoStream FFProbeStream
	for _, s := range data.Streams {
		if s.CodecType == "video" {
			videoStream = s
			break
		}
	}

	if videoStream.CodecType != "video" {
		return "", errors.New("no video stream found on file")
	}

	aspectRatio := float64(videoStream.Width) / float64(videoStream.Height)

	if almostEqual(ar16To9, aspectRatio, 1e-3) {
		return "16:9", nil
	}

	if almostEqual(ar9To16, aspectRatio, 1e-3) {
		return "9:16", nil
	}

	return "other", nil
}
