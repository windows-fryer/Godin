package routes

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"mime"

	"github.com/charmbracelet/log"
	"github.com/godin/internal/database"
	"github.com/godin/internal/discord"
)

type VFSReadSeeker struct {
	parts    []database.VFSFilePart
	client   *http.Client
	off      int64
	size     int64
	buf      []byte
	bufStart int64
}

func newVFSReadSeeker(client *http.Client, parts []database.VFSFilePart, size int64) *VFSReadSeeker {
	return &VFSReadSeeker{parts: parts, client: client, size: size}
}

func (v *VFSReadSeeker) Seek(offset int64, whence int) (int64, error) {
	var newOff int64

	switch whence {
	case io.SeekStart:
		newOff = offset
	case io.SeekCurrent:
		newOff = v.off + offset
	case io.SeekEnd:
		newOff = v.size + offset
	default:
		return 0, fmt.Errorf("invalid whence")
	}

	if newOff < 0 || newOff > v.size {
		return 0, fmt.Errorf("offset out of range")
	}

	v.off = newOff

	return v.off, nil
}

func (v *VFSReadSeeker) Read(p []byte) (int, error) {
	if v.off >= v.size {
		return 0, io.EOF
	}

	if v.buf == nil || v.off < v.bufStart || v.off >= v.bufStart+int64(len(v.buf)) {
		var dp database.VFSFilePart
		var start int64

		for _, part := range v.parts {
			start = int64(part.PartIndex)
			end := start + int64(part.PartSize)

			if v.off < end {
				dp = part
				break
			}
		}

		url := "https://cdn.discordapp.com/attachments/" + dp.PartAttachmentURL

		req, _ := http.NewRequest("GET", url, nil)

		req.Header.Set("Range", fmt.Sprintf("bytes=0-%d", dp.PartSize-1))

		resp, err := v.client.Do(req)

		if err != nil {
			return 0, err
		}

		defer resp.Body.Close()

		raw, err := io.ReadAll(resp.Body)

		if err != nil {
			return 0, err
		}

		for i := range raw {
			raw[i] ^= 0x55
		}

		v.buf = raw
		v.bufStart = start
	}

	offsetInBuf := v.off - v.bufStart

	n := copy(p, v.buf[offsetInBuf:])

	v.off += int64(n)

	if n == 0 {
		return 0, io.EOF
	}

	return n, nil
}

func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	if !strings.HasPrefix(r.URL.Path, "/v1/download/") {
		http.Error(w, "Invalid URL path", http.StatusBadRequest)
		log.Error("Invalid URL path", "path", r.URL.Path)

		return
	}

	parsedURL := strings.Split(strings.TrimPrefix(r.URL.Path, "/v1/download/"), "/")

	fileId := parsedURL[0]

	if strings.Contains(fileId, ".") {
		fileId = strings.Split(fileId, ".")[0]
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Info("Received download request", "file_id", fileId, "method", r.Method, "url", r.URL.String())

	discord.UpdateMessageAttachments(fileId)

	fileDocument, err := database.GetVFSFile(fileId)

	if err != nil {
		http.Error(w, "Failed to get file", http.StatusInternalServerError)
		log.Error("Failed to get file from database", "error", err, "file_id", fileId)

		return
	}

	ext := filepath.Ext(fileDocument.FileName)
	mimeType := mime.TypeByExtension(ext)

	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	escaped := url.PathEscape(fileDocument.FileName)

	var disposition string

	if strings.HasPrefix(mimeType, "image/") || strings.HasPrefix(mimeType, "video/") || strings.HasPrefix(mimeType, "audio/") {
		disposition = "inline"
	} else {
		disposition = "attachment"
	}

	w.Header().Set("Content-Disposition",
		fmt.Sprintf(
			"%s; filename=%q; filename*=UTF-8''%s",
			disposition, fileDocument.FileName, escaped,
		),
	)

	httpClient := &http.Client{}

	reader := newVFSReadSeeker(httpClient, fileDocument.FileParts, int64(fileDocument.FileSize))
	modTime := time.Unix(fileDocument.FileTimestamp, 0)

	http.ServeContent(w, r, fileDocument.FileName, modTime, reader)

	log.Info("File served", "file_id", fileId, "file_name", fileDocument.FileName, "size", fileDocument.FileSize, "duration", time.Since(start).String())
}
