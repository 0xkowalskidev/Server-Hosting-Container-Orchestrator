package controlpanel

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/gofiber/contrib/websocket"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type FileViewer struct {
	commandHandlers map[string]func(*websocket.Conn, *sftp.Client, string)
}

func NewFileViewer() *FileViewer {
	fv := &FileViewer{
		commandHandlers: make(map[string]func(*websocket.Conn, *sftp.Client, string)),
	}

	fv.commandHandlers["ls"] = fv.handleListDir

	return fv
}

func (fv *FileViewer) HandleFileViewer(c *websocket.Conn) {
	id := c.Params("id")
	log.Println("Connected to gameserver ID:", id)

	sftpClient, err := fv.connectSFTP(id)
	if err != nil {
		log.Println("Failed to connect to SFTP:", err)
		c.Close()
		return
	}
	defer sftpClient.Close()

	// Send initial directory listing on connection
	fv.handleListDir(c, sftpClient, ".")

	for {
		_, msg, err := c.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
			break
		}

		var data map[string]interface{}
		if err := json.Unmarshal(msg, &data); err != nil {
			log.Printf("Failed to parse JSON: %v", err)
			continue
		}

		var command, path string
		for key, value := range data {
			if key == "HEADERS" {
				continue // Skip headers
			}
			command = key
			if val, ok := value.(string); ok {
				path = val
			}
		}

		if handler, ok := fv.commandHandlers[command]; ok {
			handler(c, sftpClient, path)
		} else {
			log.Printf("Unknown command: %s", command)
		}
	}
}

func (fv *FileViewer) connectSFTP(gameserverId string) (*sftp.Client, error) {
	config := &ssh.ClientConfig{
		User:            gameserverId,
		Auth:            []ssh.AuthMethod{ssh.Password("password")}, // TODO TEMP
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	conn, err := ssh.Dial("tcp", "localhost:2022", config) // TODO TEMP
	if err != nil {
		return nil, err
	}

	client, err := sftp.NewClient(conn)
	if err != nil {
		conn.Close()
		return nil, err
	}
	return client, nil
}

func (fv *FileViewer) handleListDir(c *websocket.Conn, client *sftp.Client, path string) {
	files, err := client.ReadDir(path)
	if err != nil {
		log.Println("Failed to read directory:", err)
		c.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Error: %v", err)))
		return
	}

	var htmlResponse strings.Builder
	// Breadcrumbs
	htmlResponse.WriteString("<ul id='breadcrumbs'>")
	parts := strings.Split(path, "/")
	for i := range parts {
		if i == 0 && parts[i] == "." {
			htmlResponse.WriteString(`<li name='ls' value='.'>Home</li>`)
		} else if parts[i] != "" {
			partialPath := strings.Join(parts[:i+1], "/")
			htmlResponse.WriteString(fmt.Sprintf(
				`<li name='ls' value='%s'>%s</li>`,
				partialPath, parts[i],
			))
		}
	}
	htmlResponse.WriteString("</ul>")

	// File Form
	htmlResponse.WriteString("<form id='files' hx-trigger='click' ws-send='true'>") // Create a form with ws-send

	for _, file := range files {
		if file.IsDir() {
			htmlResponse.WriteString(fmt.Sprintf(`
                <button type='submit' name='ls' value='%s/%s'>%s/</button><br>`,
				path, file.Name(), file.Name(),
			))
		} else {
			htmlResponse.WriteString(fmt.Sprintf("<span>%s</span><br>", file.Name()))
		}
	}
	htmlResponse.WriteString("</form>") // Close the form

	if err = c.WriteMessage(websocket.TextMessage, []byte(htmlResponse.String())); err != nil {
		log.Println("write:", err)
	}
}
