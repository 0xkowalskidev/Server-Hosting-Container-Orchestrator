package workernode

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/msteinert/pam"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type SFTPServer struct {
	config Config
}

func NewSFTPServer(config Config) *SFTPServer {
	return &SFTPServer{
		config: config,
	}
}

func (s *SFTPServer) Start() error {
	// Load the server's private SSH key
	privateBytes, err := os.ReadFile(s.config.SftpKeyPath)
	if err != nil {
		return fmt.Errorf("failed to load host key: %v", err)
	}

	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		return fmt.Errorf("failed to parse host key: %v", err)
	}

	// Configure the SSH server with PAM-based password authentication
	sshConfig := &ssh.ServerConfig{
		PasswordCallback: s.passwordAuth,
	}
	sshConfig.AddHostKey(private)

	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", s.config.SftpPort))
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %v", s.config.SftpPort, err)
	}
	defer listener.Close()
	log.Printf("SFTP server listening on port %s...", s.config.SftpPort)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept incoming connection: %v", err)
			continue
		}

		go s.handleConn(conn, sshConfig)
	}
}

func (s *SFTPServer) passwordAuth(c ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
	// Using PAM to authenticate user
	t, err := pam.StartFunc("sshd", c.User(), func(s pam.Style, msg string) (string, error) {
		return string(password), nil
	})
	if err != nil {
		return nil, fmt.Errorf("PAM start failed: %v", err)
	}
	err = t.Authenticate(0)
	if err != nil {
		return nil, ssh.ErrNoAuth
	}
	return nil, nil
}

func (s *SFTPServer) handleConn(conn net.Conn, config *ssh.ServerConfig) {
	sshConn, chans, reqs, err := ssh.NewServerConn(conn, config)
	if err != nil {
		log.Printf("SSH handshake failed: %v", err)
		return
	}
	defer sshConn.Close()
	go ssh.DiscardRequests(reqs)

	// Get the username from the SSH connection
	username := sshConn.User()

	for newChannel := range chans {
		if newChannel.ChannelType() != "session" {
			newChannel.Reject(ssh.UnknownChannelType, "unsupported channel type")
			continue
		}

		channel, requests, err := newChannel.Accept()
		if err != nil {
			log.Printf("Channel acceptance failed: %v", err)
			continue
		}

		// Pass username to handleSFTP for user-specific directory setup
		go s.handleSFTP(channel, requests, username)
	}
}

func (s *SFTPServer) handleSFTP(channel ssh.Channel, requests <-chan *ssh.Request, username string) {
	for req := range requests {
		if req.Type == "subsystem" && strings.HasPrefix(string(req.Payload[4:]), "sftp") {
			req.Reply(true, nil)

			// Set the user-specific root directory
			userRoot := filepath.Join(s.config.MountsPath, username)

			// Initialize the SFTP server with user root and debug options
			server, err := sftp.NewServer(
				channel,
				sftp.WithDebug(os.Stderr),
				sftp.WithServerWorkingDirectory(userRoot),
			)
			if err != nil {
				log.Printf("Failed to start SFTP server: %v", err)
				return
			}

			// Serve SFTP requests
			if err := server.Serve(); err != nil && err != io.EOF {
				log.Printf("SFTP server error: %v", err)
			}
			return
		}
		req.Reply(false, nil)
	}
}
