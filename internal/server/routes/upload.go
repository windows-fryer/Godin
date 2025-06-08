package routes

import (
	"bytes"
	"errors"
	"io"
	"math/rand/v2"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/charmbracelet/log"
	"github.com/godin/internal/database"
	"github.com/godin/internal/discord"
	"github.com/google/uuid"
)

const chunkSize = 10485316 //Discord said so

var chunkBufferPool = sync.Pool{
	New: func() any {
		b := make([]byte, chunkSize)
		return &b
	},
}

func getWebhook(guildId int) (map[string]any, error) {
	directories, err := database.GetVFSDirectories(guildId)

	if err != nil {
		log.Error("Failed to get VFS directories", "error", err)

		return nil, err
	}

	directoryCount := len(directories)
	randomIndex := rand.IntN(directoryCount)

	return directories[randomIndex], nil
}

func uploadToWebhook(webhookId string, webhookToken string, data []byte) (*discordgo.Message, error) {
	return discord.DiscordClient.WebhookExecute(webhookId, webhookToken, true, &discordgo.WebhookParams{
		Files: []*discordgo.File{{
			Name:   uuid.New().String(),
			Reader: bytes.NewReader(data),
		}},
	})
}

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	log.Info("Received upload request", "method", r.Method, "url", r.URL.String())

	if !strings.HasPrefix(r.URL.Path, "/v1/upload/") {
		http.Error(w, "Invalid URL path", http.StatusBadRequest)
		log.Error("Invalid URL path", "path", r.URL.Path)

		return
	}

	guildId := strings.TrimPrefix(r.URL.Path, "/v1/upload/")
	guildIdToInt, err := strconv.Atoi(guildId)

	if err != nil {
		http.Error(w, "Invalid guild ID", http.StatusBadRequest)
		log.Error("Invalid guild ID", "error", err, "guild_id", guildId)

		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)

		return
	}

	reader, err := r.MultipartReader()

	if err != nil {
		log.Error("Failed to read multipart form", "error", err)
		http.Error(w, "Failed to read multipart form", http.StatusBadRequest)
		return
	}

	part, err := reader.NextPart()

	if err != nil {
		log.Error("Failed to get next part from multipart reader", "error", err)
		http.Error(w, "Failed to process multipart data", http.StatusInternalServerError) // Send error to client
		return
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	type WebhookMessage struct {
		Message *discordgo.Message
		Part    int
		Size    int
	}

	messages := make([]WebhookMessage, 0)
	fileSize := 0

	var parts uint64

	for {
		bufferPtr := chunkBufferPool.Get().(*[]byte)
		buffer := *bufferPtr
		curOffset := 0

		for {
			n, readErr := part.Read(buffer[curOffset:])
			if n > 0 {
				curOffset += n
				fileSize += n
			}

			if readErr != nil {
				if errors.Is(readErr, io.EOF) {
					break
				}

				log.Error("Error reading from multipart part", "error", readErr)
				http.Error(w, "Failed to read uploaded file stream", http.StatusInternalServerError)

				chunkBufferPool.Put(bufferPtr)

				return
			}

			if curOffset == len(buffer) {
				break
			}

			if n == 0 {
				break
			}
		}
		if curOffset == 0 {
			chunkBufferPool.Put(bufferPtr)
			break
		}

		payload := make([]byte, curOffset)
		copy(payload, buffer[:curOffset])

		chunkBufferPool.Put(bufferPtr)

		partNum := int(atomic.AddUint64(&parts, 1) - 1)

		log.Info("Processing chunk", "part_index", partNum, "size", len(payload))

		webhookDocument, whErr := getWebhook(guildIdToInt)
		if whErr != nil {
			log.Error("Failed to get webhook URL", "error", whErr)
			http.Error(w, "Failed to get webhook URL", http.StatusInternalServerError)
			return // Abort upload
		}

		for i := range payload {
			payload[i] ^= 0x55
		}

		wg.Add(1)

		go func(webhookID int64, webhookToken string, dataToSend []byte, partNum int) {
			defer wg.Done()

			webhookIDtoStr := strconv.FormatInt(webhookID, 10)
			msg, uploadErr := uploadToWebhook(webhookIDtoStr, webhookToken, dataToSend)

			if uploadErr != nil {
				log.Error("Failed to upload chunk", "error", uploadErr, "webhook_id", webhookID, "part_num", partNum)

				return
			}

			mu.Lock()
			messages = append(messages, WebhookMessage{Message: msg, Part: partNum, Size: len(dataToSend)})
			mu.Unlock()

			log.Info("Chunk uploaded successfully", "part_num", partNum, "size", len(dataToSend), "webhook_id", webhookID)
		}(webhookDocument["webhook_id"].(int64), webhookDocument["webhook_token"].(string), payload, partNum)
	}

	wg.Wait()

	totalParts := int(atomic.LoadUint64(&parts))

	if len(messages) != totalParts {
		log.Error("Mismatch in expected parts and successfully uploaded parts", "expected", parts, "successful", len(messages))
		http.Error(w, "One or more file parts failed to upload", http.StatusInternalServerError)
		return
	}

	sortedParts := make([]WebhookMessage, totalParts)
	for _, msg := range messages {
		if msg.Part < 0 || msg.Part >= totalParts {
			log.Error("Invalid part number", "part", msg.Part, "total_parts", totalParts)
			http.Error(w, "Internal error processing uploaded parts", http.StatusInternalServerError)
			return
		}
		sortedParts[msg.Part] = msg
	}

	for i, p := range sortedParts {
		if p.Message == nil {
			log.Error("A file part is missing after sorting, indicating an upload failure for that part.", "missing_part_index", i)
			http.Error(w, "A file part failed to upload or process correctly", http.StatusInternalServerError)
			return
		}
	}

	parsedMessages := database.VFSFile{
		FileID:        uuid.NewString(),
		FileName:      part.FileName(),
		FileSize:      uint64(fileSize),
		FileTimestamp: time.Now().Unix(),
		FileGuildID:   guildIdToInt,
		FileParts:     make([]database.VFSFilePart, len(sortedParts)),
	}

	currentSizeOffset := 0
	for i, msg := range sortedParts {
		messageIdToInt, _ := strconv.Atoi(msg.Message.ID)
		channelIdToInt, _ := strconv.Atoi(msg.Message.ChannelID)

		parsedMessages.FileParts[i] = database.VFSFilePart{
			PartID:            int64(messageIdToInt),
			PartChannelID:     int64(channelIdToInt),
			PartAttachmentURL: strings.Split(msg.Message.Attachments[0].URL, "https://cdn.discordapp.com/attachments/")[1],
			PartSize:          uint64(msg.Size),
			PartIndex:         uint64(currentSizeOffset),
		}
		currentSizeOffset += msg.Size
	}

	if err := database.AppendVFSFile(parsedMessages); err != nil {
		log.Error("Failed to append VFS file", "error", err)
		http.Error(w, "Failed to append VFS file", http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(parsedMessages.FileID))

	log.Info("Upload completed successfully", "guild_id", guildIdToInt, "file_id", parsedMessages.FileID, "file_size", fileSize)
}
