package router

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"cloud.google.com/go/logging"
	"github.com/gin-gonic/gin"
)

func SetLogRoutes(g *gin.RouterGroup) {
	g.POST("/ioLogging", ioLogging)
}

type IoLog struct {
	Category      string `json:"category,omitempty"`
	CategorySub   string `json:"categorySub,omitempty"`
	Txt           string `json:"txt,omitempty"`
	Ip            string `json:"ip,omitempty"`
	XForwardedFor string `json:"xForwardedFor,omitempty"`
}

func LogSeverityParse(severity string) (logging.Severity, error) {
	switch strings.TrimSpace(strings.ToLower(severity)) {
	case "info":
		return logging.Info, nil
	case "debug":
		return logging.Debug, nil
	case "warning":
		return logging.Warning, nil
	case "warn":
		return logging.Warning, nil
	case "error":
		return logging.Error, nil
	case "emergency":
		return logging.Emergency, nil
	default:
		return logging.Default, fmt.Errorf("not matched severity: %s string on LogSeverityParse ", severity)
	}
}
func ioLogging(c *gin.Context) {
	ctx := context.Background()
	// c.String(http.StatusBadRequest, "Fail at GetQuery(price)")
	// Creates a client.
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		log.Fatalf("Not Specified env GOOGLE_CLOUD_PROJECT: %v", projectID)
	}
	// log.Printf("projectID: %s, len: %v", projectID, len(projectID))
	client, err := logging.NewClient(ctx, projectID)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Failed to create client: %v", err))
	}
	defer client.Close()
	logName := c.PostForm("logName")
	if logName == "" {
		c.String(http.StatusBadRequest, "logName Field is required")
	}
	// log.Printf("Log Name: %s, len: %v", logName, len(logName))
	logger := client.Logger(strings.TrimSpace(strings.ToLower(logName)))
	defer logger.Flush() // Ensure the entry is written.
	severity, err := LogSeverityParse(c.PostForm("severity"))
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		log.Fatalln(err)
	}
	logger.Log(logging.Entry{
		// Log anything that can be marshaled to JSON.
		Payload: IoLog{
			Category:      c.PostForm("category"),
			CategorySub:   c.PostForm("categorySub"),
			Txt:           c.PostForm("txt"),
			Ip:            c.ClientIP(),
			XForwardedFor: c.GetHeader("X-FORWARDED-FOR"),
		},
		Severity: severity,
	})
	c.String(http.StatusAccepted, "Success")
}
