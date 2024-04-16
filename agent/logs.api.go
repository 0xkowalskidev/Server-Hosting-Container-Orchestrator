package agent

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hpcloud/tail"
)

// Http log server
// StreamLogsHandler streams container logs to the client.
func (a *Agent) StreamLogsHandler(c *gin.Context) {
	namespace := c.Param("namespace") // TAKE ME FROM CONFIG, ENSURE CORRECT
	containerID := c.Param("containerID")
	logFilePath := a.cfg.LogPath + namespace + "-" + containerID + ".log" // HANDLE THIS PROPERLY

	t, err := tail.TailFile(logFilePath, tail.Config{Follow: true})
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to tail log file: %s", err)
		return
	}

	// Ensure the tailing goroutine is properly stopped when the client disconnects.
	defer t.Stop()

	c.Stream(func(w io.Writer) bool {
		select {
		case line := <-t.Lines:
			if line.Err != nil {
				c.Error(line.Err) // Use Gin's error handling mechanism.
				return false
			}
			_, err := c.Writer.WriteString(line.Text + "\n")
			if err != nil {
				// Error writing to the client, stop streaming.
				return false
			}
			return true
		case <-c.Request.Context().Done():
			// Client disconnected, stop streaming.
			return false
		}
	})
}

func (a *Agent) startLogApi() {
	r := gin.Default()

	// Set up the route with URL parameters captured by Gin.
	r.GET("/namespaces/:namespace/containers/:containerID/logs", a.StreamLogsHandler)

	r.Run(":8081")
}
