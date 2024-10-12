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

		go s.handleSFTP(channel, requests)
	}
}

func (s *SFTPServer) handleSFTP(channel ssh.Channel, requests <-chan *ssh.Request) {
	for req := range requests {
		if req.Type == "subsystem" && strings.HasPrefix(string(req.Payload[4:]), "sftp") {
			req.Reply(true, nil)

			handler := newSFTPHandler(s.config.MountsPath)
			handlers := sftp.Handlers{
				FileGet:  handler,
				FilePut:  handler,
				FileCmd:  handler,
				FileList: handler,
			}

			server := sftp.NewRequestServer(channel, handlers)

			if err := server.Serve(); err != nil && err != sftp.ErrSshFxConnectionLost {
				log.Printf("SFTP server error: %v", err)
			}
			return
		}
		req.Reply(false, nil)
	}
}

type sftpHandler struct {
	root string
}

func newSFTPHandler(rootDir string) *sftpHandler {
	return &sftpHandler{root: rootDir}
}

func (h *sftpHandler) Fileread(req *sftp.Request) (io.ReaderAt, error) {
	filePath := filepath.Join(h.root, req.Filepath)
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func (h *sftpHandler) Filewrite(req *sftp.Request) (io.WriterAt, error) {
	filePath := filepath.Join(h.root, req.Filepath)
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return nil, err
	}
	return file, nil
}

type listerAt struct {
	files []os.FileInfo
}

// ListAt populates the buffer with directory entries, up to the length of the buffer
func (l *listerAt) ListAt(buffer []os.FileInfo, offset int64) (int, error) {
	if offset >= int64(len(l.files)) {
		return 0, io.EOF
	}
	n := copy(buffer, l.files[offset:])
	if n < len(buffer) {
		return n, io.EOF
	}
	return n, nil
}

// Filelist implementation for FileLister
func (h *sftpHandler) Filelist(req *sftp.Request) (sftp.ListerAt, error) {
	dirPath := filepath.Join(h.root, req.Filepath)
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	fileInfos := make([]os.FileInfo, len(entries))
	for i, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			return nil, err
		}
		fileInfos[i] = info
	}

	return &listerAt{files: fileInfos}, nil
}

func (h *sftpHandler) Filecmd(req *sftp.Request) error {
	filePath := filepath.Join(h.root, req.Filepath)
	switch req.Method {
	case "Setstat":
		return os.Chmod(filePath, req.Attributes().FileMode())
	case "Rename":
		return os.Rename(filePath, filepath.Join(h.root, req.Target))
	case "Remove":
		return os.Remove(filePath)
	case "Mkdir":
		return os.Mkdir(filePath, 0755)
	case "Rmdir":
		return os.Remove(filePath)
	default:
		return sftp.ErrSshFxOpUnsupported
	}
}
