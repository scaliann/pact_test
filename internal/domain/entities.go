package domain

type Session struct {
	SessionId string
	QrCode    string
}

type MessageUpdate struct {
	MessageID int64
	From      string
	Text      string
	Timestamp int64
}
