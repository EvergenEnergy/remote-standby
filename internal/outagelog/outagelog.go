package outagelog

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"
)

type Handler struct {
	logger     *slog.Logger
	fileHandle *os.File
	filePath   string
}

func NewHandler(filePath string, logger *slog.Logger) (*Handler, error) {
	fileHandle, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("creating outage log file handle for path %s: %w", filePath, err)
	}
	return &Handler{filePath: filePath, fileHandle: fileHandle, logger: logger}, nil
}

func (h *Handler) Append(message string, details map[string]string) {
	errDetails := []string{}
	for k, v := range details {
		errDetails = append(errDetails, fmt.Sprintf("%s=%s", k, v))
	}

	errStr := fmt.Sprintf("%s: %s [%s]\n", time.Now().Format(time.RFC3339), message, strings.Join(errDetails, "+"))

	if _, err := h.fileHandle.Write([]byte(errStr)); err != nil {
		h.logger.Error("appending to outage log", "error", err)
	}
}

func (h *Handler) Close() {
	if err := h.fileHandle.Close(); err != nil {
		h.logger.Error("closing outage log", "error", err)
	}
}
