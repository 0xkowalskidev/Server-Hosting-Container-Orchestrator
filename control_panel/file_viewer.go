package controlpanel

import (
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
	log.Println(id)
	sftpClient, err := fv.connectSFTP(id)
	if err != nil {
		log.Println("Failed to connect to SFTP:", err)
		c.Close()
		return
	}
	defer sftpClient.Close()

	fv.handleListDir(c, sftpClient, ".")

	// Continuously listen for incoming messages and handle them
	for {
		_, msg, err := c.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
			break
		}

		messageParts := strings.SplitN(string(msg), " ", 2)
		command := messageParts[0]
		path := "."
		if len(messageParts) > 1 {
			path = messageParts[1]
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
	htmlResponse.WriteString("<ul id='files'>")
	for _, file := range files {
		htmlResponse.WriteString(fmt.Sprintf("<li>%s</li>", file.Name()))
	}
	htmlResponse.WriteString("</ul>")

	if err = c.WriteMessage(websocket.TextMessage, []byte(htmlResponse.String())); err != nil {
		log.Println("write:", err)
	}
}
