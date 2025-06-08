package database

import (
	"fmt"
	"strconv"
)

func guildIdToInt(guildID string) (int, error) {
	intGuildID, err := strconv.Atoi(guildID)

	if err != nil {
		return 0, fmt.Errorf("invalid guild ID: %w", err)
	}

	return intGuildID, nil
}

func deleteIfExists(guildID string) error {
	guildIdToInt, err := guildIdToInt(guildID)

	if err != nil {
		return err
	}

	if err := DeleteVFSFiles(guildIdToInt); err != nil {
		return fmt.Errorf("failed to delete existing VFS files: %w", err)
	}

	collection := MongoSession.Database("godin").Collection("guilds")

	_, err = collection.DeleteOne(MongoContext, map[string]any{
		"_id": guildIdToInt,
	})

	if err != nil {
		return err
	}

	return nil
}

func CreateGuild(guildID string) error {
	guildIdToInt, err := guildIdToInt(guildID)

	if err != nil {
		return err
	}

	collection := MongoSession.Database("godin").Collection("guilds")

	deleteIfExists(guildID)

	_, err = collection.InsertOne(MongoContext, map[string]any{
		"_id": guildIdToInt,

		"vfs_channels": []any{},
	})

	if err != nil {
		return err
	}

	return nil
}

func AddVFSDirectories(guildID string, channels []map[string]any) error {
	guildIdToInt, err := guildIdToInt(guildID)

	if err != nil {
		return err
	}

	collection := MongoSession.Database("godin").Collection("guilds")

	_, err = collection.UpdateOne(MongoContext, map[string]any{
		"_id": guildIdToInt,
	}, map[string]any{
		"$addToSet": map[string]any{
			"vfs_channels": map[string]any{
				"$each": channels,
			},
		},
	})

	if err != nil {
		return err
	}

	return nil
}

func GetVFSDirectories(guildID int) ([]map[string]any, error) {
	collection := MongoSession.Database("godin").Collection("guilds")

	var result struct {
		VFSChannels []map[string]any `bson:"vfs_channels"`
	}

	err := collection.FindOne(MongoContext, map[string]any{
		"_id": guildID,
	}).Decode(&result)

	if err != nil {
		return nil, err
	}

	return result.VFSChannels, nil
}

func GetVFSFiles(guildID int) ([]map[string]any, error) {
	collection := MongoSession.Database("godin").Collection("files")

	cursor, err := collection.Find(MongoContext, map[string]any{
		"file_guild_id": guildID,
	})

	if err != nil {
		return nil, err
	}

	var files []map[string]any
	if err := cursor.All(MongoContext, &files); err != nil {
		return nil, err
	}

	return files, nil
}

type VFSFilePart struct {
	PartID            int64  `bson:"_id"`
	PartAttachmentURL string `bson:"attachment_url"`
	PartSize          int32  `bson:"part_size"`
	PartChannelID     int64  `bson:"channel_id"`
	PartIndex         int32  `bson:"part_index"`
}

type VFSFile struct {
	FileID        string        `bson:"_id"`
	FileGuildID   int           `bson:"file_guild_id"`
	FileName      string        `bson:"file_name"`
	FileSize      int32         `bson:"file_size"`
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

func AppendVFSFile(fileID string, file VFSFile) error {
	collection := MongoSession.Database("godin").Collection("files")

	_, err := collection.InsertOne(MongoContext, file)

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
