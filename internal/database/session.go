package database

type SessionRequest struct {
	FileName string `json:"file_name"`
	FileSize int64  `json:"file_size"`
}

type Session struct {
	SessionToken     string `json:"session_token"`
	SessionChunkSize int64  `json:"session_chunk_size"`
}

func CreateSession(service ServiceItemSearch, session SessionRequest) (*Session, error) {
	// dynamoClient.PutItem(context.TODO(), &dynamodb.PutItemInput{
	// 	TableName: aws.String("GodinVFS-Sessions"),
	// 	,
	// })

	return nil, nil
}
