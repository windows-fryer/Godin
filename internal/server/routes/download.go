package routes

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/godin/internal/database"
)

type VFSReadSeeker struct {
	parts  []database.VFSFilePart
	client *http.Client
	off    int64
	size   int64
}

func NewVFSReadSeeker(client *http.Client, parts []database.VFSFilePart, size int64) *VFSReadSeeker {
	return &VFSReadSeeker{parts: parts, client: client, size: size}
}

// Seek sets the read offset
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

// Read will pull exactly len(p) bytes (or less at EOF) from the appropriate part(s)
func (v *VFSReadSeeker) Read(p []byte) (int, error) {
	if v.off >= v.size {
		return 0, io.EOF
	}

	// Find which FilePart contains v.off
	for _, part := range v.parts {
		start := int64(part.PartIndex)
		end := start + int64(part.PartSize)
		if v.off < end {
			// how many bytes we can read from this part
			maxInPart := int(end - v.off)
			toRead := len(p)
			if toRead > maxInPart {
				toRead = maxInPart
			}

			// request exactly that slice from Discord
			url := "https://cdn.discordapp.com/attachments/" + part.PartAttachmentURL
			req, _ := http.NewRequest("GET", url, nil)
			// Range header within that part
			req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", v.off-start, v.off-start+int64(toRead)-1))
			resp, err := v.client.Do(req)
			if err != nil {
				return 0, err
			}
			defer resp.Body.Close()

			// fetch & XOR-decode
			buf := make([]byte, toRead)
			if _, err := io.ReadFull(resp.Body, buf); err != nil {
				return 0, err
			}
			for i := range buf {
				buf[i] ^= 0x55
			}
			copy(p, buf)

			v.off += int64(toRead)
			return toRead, nil
		}
	}

	// somehow past all parts
	return 0, io.EOF
}

func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, "/v1/download/") {
		http.Error(w, "Invalid URL path", http.StatusBadRequest)
		log.Error("Invalid URL path", "path", r.URL.Path)

		return
	}

	parsedURL := strings.Split(strings.TrimPrefix(r.URL.Path, "/v1/download/"), "/")

	fileId := parsedURL[0]

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Info("Received download request", "file_id", fileId, "method", r.Method, "url", r.URL.String())

	fileDocument, err := database.GetVFSFile(fileId)

	if err != nil {
		http.Error(w, "Failed to get file", http.StatusInternalServerError)
		log.Error("Failed to get file from database", "error", err, "file_id", fileId)

		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Content-Length", strconv.FormatInt(int64(fileDocument.FileSize), 10))
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Content-Disposition", "attachment; filename="+fileDocument.FileName)

	rangeHeader := r.Header.Get("Range")

	byteRange := strings.Split(strings.TrimPrefix(rangeHeader, "bytes="), "-")

	minimumRange, err := strconv.Atoi(byteRange[0])

	if err != nil {
		minimumRange = 0
	}

	var maximumRange int

	if len(byteRange) > 1 && byteRange[1] != "" {
		maximumRange, err = strconv.Atoi(byteRange[1])

		if err != nil {
			http.Error(w, "Invalid maximum range", http.StatusBadRequest)
			log.Error("Invalid maximum range", "error", err, "file_id", fileId, "range_header", rangeHeader)

			return
		}
	} else {
		maximumRange = -1 // -1 indicates no maximum range
	}

	log.Info("Parsed range header", "minimum_range", minimumRange, "maximum_range", byteRange)

	log.Info("Processing file download", "file_id", fileId, "range_header", rangeHeader)

	streamedParts := make([]database.VFSFilePart, 0, len(fileDocument.FileParts))

	for _, part := range fileDocument.FileParts {
		partIndex := int(part.PartIndex)

		if partIndex+int(part.PartSize) <= minimumRange {
			continue
		}

		if maximumRange != -1 && partIndex > maximumRange {
			break
		}

		streamedParts = append(streamedParts, part)
	}

	httpClient := &http.Client{}

	for _, part := range streamedParts {
		// partResponse, err := http.Get("https://cdn.discordapp.com/attachments/" + partURL)

		// Request the part from the CDN with a valid range header to avoid downloading the entire file

		partIndex := int(part.PartIndex)
		// partSize := int(part.PartSize)

		partURL := "https://cdn.discordapp.com/attachments/" + part.PartAttachmentURL

		request, err := http.NewRequest("GET", partURL, nil)

		if err != nil {
			http.Error(w, "Failed to create request for part", http.StatusInternalServerError)
			log.Error("Failed to create request for part", "error", err, "part_url", partURL)

			return
		}

		partMinimumRange := max(0, minimumRange-partIndex)

		var partMaximumRange int

		if maximumRange != -1 {
			partMaximumRange = maximumRange - partIndex
		} else {
			partMaximumRange = int(part.PartSize)
		}

		log.Info("Setting range header for part", "part_url", partURL, "minimum_range", partMinimumRange, "maximum_range", partMaximumRange)

		request.Header.Set("Range", "bytes="+strconv.Itoa(partMinimumRange)+"-"+strconv.Itoa(partMaximumRange))

		log.Info("Downloading part", "range", request.Header.Get("Range"), "part_url", partURL)

		partResponse, err := httpClient.Do(request)

		if err != nil {
			http.Error(w, "Failed to download part", http.StatusInternalServerError)
			log.Error("Failed to download part", "error", err)

			return
		}

		defer partResponse.Body.Close()

		buffer := make([]byte, 1024*1024)

		totalBytesRead := 0

		for {
			n, err := partResponse.Body.Read(buffer)
			totalBytesRead += n

			for nibble := range buffer[:n] {
				buffer[nibble] ^= 0x55
			}

			if n > 0 {
				if _, err := w.Write(buffer[:n]); err != nil {
					http.Error(w, "Failed to write part to response", http.StatusInternalServerError)
					log.Error("Failed to write part to response", "error", err)

					return
				}
			}

			if err != nil {
				if err.Error() != "EOF" {
					http.Error(w, "Failed to read part", http.StatusInternalServerError)
					log.Error("Failed to read part", "error", err)

					return
				}
				break
			}
		}

		log.Info("Successfully streamed part", "size", partResponse.ContentLength)
	}
}
