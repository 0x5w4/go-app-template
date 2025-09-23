package shared

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"strings"
)

func CheckBase64Image(base64Str string) (string, error) {
	if strings.HasPrefix(base64Str, "data:image/") {
		parts := strings.SplitN(base64Str, ",", 2)
		if len(parts) > 1 {
			base64Str = parts[1]
		}
	}

	data, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 string: %w", err)
	}

	_, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("failed to decode image: %w", err)
	}

	if format == "jpeg" || format == "png" {
		return format, nil
	}

	return "", fmt.Errorf("unsupported image format: %s", format)
}
