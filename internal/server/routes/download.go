package routes

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/godin/internal/database"
)

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

		// Send 4kb chunks to the multipart writer
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
