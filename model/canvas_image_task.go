package model

type CanvasImageTask struct {
	ID              string `json:"id" gorm:"primaryKey;size:64"`
	UserID          string `json:"userId" gorm:"size:64"`
	UserDisplayName string `json:"userDisplayName" gorm:"size:191"`
	Source          string `json:"source" gorm:"size:32"`
	SourceID        string `json:"sourceId" gorm:"size:64"`
	NodeID          string `json:"nodeId" gorm:"size:64"`
	Model           string `json:"model" gorm:"size:128"`
	ChannelID       string `json:"channelId" gorm:"size:64"`
	UserChannelID   string `json:"userChannelId" gorm:"size:64"`
	ChannelName     string `json:"channelName" gorm:"size:191"`
	Status          string `json:"status" gorm:"size:32"`
	Progress        int    `json:"progress"`
	Prompt          string `json:"prompt" gorm:"type:mediumtext"`
	GenerationType  string `json:"generationType" gorm:"size:32"`
	Endpoint        string `json:"endpoint" gorm:"size:191"`
	ContentType     string `json:"contentType" gorm:"size:128"`
	RequestBody     string `json:"requestBody" gorm:"type:mediumtext"`
	ResponseBody    string `json:"responseBody" gorm:"type:mediumtext"`
	Error           string `json:"error" gorm:"type:mediumtext"`
	ErrorDetail     string `json:"errorDetail" gorm:"type:mediumtext"`
	ImageURL        string `json:"imageUrl" gorm:"type:mediumtext"`
	StorageKey      string `json:"storageKey" gorm:"size:512"`
	Width           int    `json:"width"`
	Height          int    `json:"height"`
	MimeType        string `json:"mimeType" gorm:"size:128"`
	Bytes           int64  `json:"bytes"`
	CreatedAt       string `json:"createdAt" gorm:"size:64"`
	UpdatedAt       string `json:"updatedAt" gorm:"size:64"`
	StartedAt       string `json:"startedAt" gorm:"size:64"`
	CompletedAt     string `json:"completedAt" gorm:"size:64"`
}
