package video

import (
	"fmt"
	"os/exec"
)

func ProcessVideoForFastStart(filePath string) (string, error) {
	outFilePath := fmt.Sprintf("%s.processing", filePath)

	cmd := exec.Command("ffmpeg", "-i", filePath, "-c", "copy", "-movflags", "faststart", "-f", "mp4", outFilePath)
	err := cmd.Run()
	if err != nil {
		return "", err
	}

	return outFilePath, nil
}
