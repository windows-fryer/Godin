package database

import (
	"fmt"
)

type Guild struct {
	GuildID     int            `bson:"_id"`
	VFSChannels []GuildChannel `bson:"vfs_channels"`
}

type GuildChannel struct {
	ChannelID int64 `bson:"_id"`

	WebhookID    int64  `bson:"webhook_id"`
	WebhookToken string `bson:"webhook_token"`
}

func deleteIfExists(guild *Guild) error {
	if err := DeleteVFSFiles(guild.GuildID); err != nil {
		return fmt.Errorf("failed to delete existing VFS files: %w", err)
	}

	collection := MongoSession.Database("godin").Collection("guilds")

	_, err := collection.DeleteOne(MongoContext, map[string]any{
		"_id": guild.GuildID,
	})

	if err != nil {
		return err
	}

	return nil
}

func CreateGuild(guild *Guild) error {
	collection := MongoSession.Database("godin").Collection("guilds")

	deleteIfExists(guild)

	_, err := collection.InsertOne(MongoContext, guild)

	if err != nil {
		return err
	}

	return nil
}

func UpdateGuild(guild *Guild) error {
	collection := MongoSession.Database("godin").Collection("guilds")

	_, err :=
		collection.UpdateOne(MongoContext, map[string]any{
			"_id": guild.GuildID,
		}, map[string]any{
			"$set": guild,
		})

	if err != nil {
		return fmt.Errorf("failed to update guild: %w", err)
	}

	return nil
}

func GetGuild(guildID int) (*Guild, error) {
	collection := MongoSession.Database("godin").Collection("guilds")

	var guild Guild

	err := collection.FindOne(MongoContext, map[string]any{
		"_id": guildID,
	}).Decode(&guild)

	if err != nil {
		return nil, fmt.Errorf("failed to get guild: %w", err)
	}

	return &guild, nil
}

type VFSFilePart struct {
	PartID            int64  `bson:"_id"`
	PartAttachmentURL string `bson:"attachment_url"`
	PartSize          uint64 `bson:"part_size"`
	PartChannelID     int64  `bson:"channel_id"`
	PartIndex         uint64 `bson:"part_index"`
	PartTimestamp     int64  `bson:"part_timestamp"`
}

type VFSFile struct {
	FileID        string        `bson:"_id"`
	FileGuildID   int           `bson:"file_guild_id"`
	FileName      string        `bson:"file_name"`
	FileSize      uint64        `bson:"file_size"`
	FileTimestamp int64         `bson:"file_timestamp"`
	FileParts     []VFSFilePart `bson:"parts"`
}

func GetVFSFile(fileID string) (*VFSFile, error) {
	collection := MongoSession.Database("godin").Collection("files")

	var file VFSFile

	err := collection.FindOne(MongoContext, map[string]any{
		"_id": fileID,
	}).Decode(&file)

	if err != nil {
		return nil, err
	}

	return &file, nil
}

func DeleteVFSFiles(guildID int) error {
	collection := MongoSession.Database("godin").Collection("files")

	_, err := collection.DeleteMany(MongoContext, map[string]any{
		"file_guild_id": guildID,
	})

	if err != nil {
		return err
	}

	return nil
}

func AppendVFSFile(file *VFSFile) error {
	collection := MongoSession.Database("godin").Collection("files")

	_, err := collection.InsertOne(MongoContext, file)

	if err != nil {
		return err
	}

	return nil
}

func UpdateVFSFile(file *VFSFile) error {
	collection := MongoSession.Database("godin").Collection("files")

	_, err := collection.UpdateOne(MongoContext, map[string]any{
		"_id": file.FileID,
	}, map[string]any{
		"$set": file,
	})

	if err != nil {
		return err
	}

	return nil
}

func DeleteVFSFile(fileID string) error {
	collection := MongoSession.Database("godin").Collection("files")

	_, err := collection.DeleteOne(MongoContext, map[string]any{
		"_id": fileID,
	})

	if err != nil {
		return err
	}

	return nil
}
