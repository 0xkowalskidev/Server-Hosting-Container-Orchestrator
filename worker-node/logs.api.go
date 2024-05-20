package workernode

import (
	"0xKowalski1/container-orchestrator/config"
	"net/http"

	"github.com/hpcloud/tail"
	"github.com/labstack/echo/v4"
)

type MetricsAndLogsApi struct {
	cfg *config.Config
}

func NewMetricsAndLogsApi(cfg *config.Config) *MetricsAndLogsApi {
	return &MetricsAndLogsApi{
		cfg: cfg,
	}
}

func (api *MetricsAndLogsApi) StreamLogsHandler(c echo.Context) error {
	namespace := api.cfg.Namespace
	containerID := c.Param("containerID")
	logFilePath := api.cfg.LogPath + namespace + "-" + containerID + ".log"

	t, err := tail.TailFile(logFilePath, tail.Config{Follow: true})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to tail log file: "+err.Error())
	}

	// Ensure the tailing goroutine is properly stopped when the client disconnects.
	defer t.Stop()

	c.Response().Header().Set(echo.HeaderContentType, "text/plain")
	c.Response().WriteHeader(http.StatusOK)

	for {
		select {
		case line := <-t.Lines:
			if line.Err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "Error reading log line: "+line.Err.Error())
			}
			if _, err := c.Response().Write([]byte(line.Text + "\n")); err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "Failed to write to client: "+err.Error())
			}
			c.Response().Flush() // Ensure the data is sent to the client immediately.
		case <-c.Request().Context().Done():
			return nil // Properly end the response if the client disconnects.
		}
	}
}

// Start starts the Echo server and sets up routes.
func (api *MetricsAndLogsApi) Start() {
	e := echo.New()

	e.GET("/containers/:containerID/logs", api.StreamLogsHandler)

	e.Logger.Fatal(e.Start(":" + "8081"))
}
