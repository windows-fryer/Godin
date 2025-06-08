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
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/charmbracelet/log"
	"github.com/godin/internal/database"
	"github.com/godin/internal/discord"
	"github.com/google/uuid"
)

const chunkSize = 10485316 //Discord said so

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
		log.Error("Failed to get next part", "error", err)

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
	parts := 0

	for {
		buf := make([]byte, chunkSize)
		curOffset := 0

		for {
			n, err := part.Read(buf[curOffset:])

			if err != nil && !errors.Is(err, io.EOF) {
				break
			}

			if n == 0 {
				break
			}

			curOffset += n
			fileSize += n
		}

		if curOffset == 0 {
			break
		}

		log.Info("Received chunk", "size", curOffset)

		webhookDocument, err := getWebhook(guildIdToInt)

		if err != nil {
			log.Error("Failed to get webhook URL", "error", err)
			http.Error(w, "Failed to get webhook URL", http.StatusInternalServerError)

			return
		}

		for nibble := range buf {
			buf[nibble] ^= 0x55
		}

		wg.Add(1)

		go func(webhookID int64, webhookToken string, payload []byte, part int) {
			defer wg.Done()

			webhookIDtoStr := strconv.FormatInt(webhookID, 10)

			msg, err := uploadToWebhook(webhookIDtoStr, webhookToken, payload)
			if err != nil {
				log.Error("Failed to upload chunk", "error", err, "webhook_id", webhookID)

				return
			}

			mu.Lock()

			messages = append(messages, WebhookMessage{msg, part, len(payload)})

			mu.Unlock()

			log.Info("Chunk uploaded successfully", "size", len(payload), "webhook_id", webhookID)
		}(webhookDocument["webhook_id"].(int64), webhookDocument["webhook_token"].(string), buf, parts)

		parts++
	}

	wg.Wait()

	sortedParts := make([]WebhookMessage, len(messages))

	for _, msg := range messages {
		sortedParts[msg.Part] = msg
	}

	parsedMessages := database.VFSFile{
		FileID:        uuid.NewString(),
		FileName:      part.FileName(),
		FileSize:      int32(fileSize),
		FileTimestamp: time.Now().Unix(),
		FileGuildID:   guildIdToInt,

		FileParts: make([]database.VFSFilePart, len(sortedParts)),
	}

	currentSize := 0

	for i, msg := range sortedParts {
		messageIdToInt, _ := strconv.Atoi(msg.Message.ID)
		channelIdToInt, _ := strconv.Atoi(msg.Message.ChannelID)

		parsedMessages.FileParts[i] = database.VFSFilePart{
			PartID: int64(messageIdToInt),

			PartChannelID:     int64(channelIdToInt),
			PartAttachmentURL: strings.Split(msg.Message.Attachments[0].URL, "https://cdn.discordapp.com/attachments/")[1],
			PartSize:          int32(msg.Size),
			PartIndex:         int32(currentSize),
		}

		currentSize += msg.Size
	}

	if err := database.AppendVFSFile(r.Header.Get("X-Guild-ID"), parsedMessages); err != nil {
		log.Error("Failed to append VFS file", "error", err)
		http.Error(w, "Failed to append VFS file", http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(parsedMessages.FileID))

	log.Info("Upload completed successfully", "guild_id", guildIdToInt, "file_id", parsedMessages.FileID, "file_size", fileSize)
}
