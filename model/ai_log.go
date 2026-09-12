package model

type AICallLog struct {
	ID              string `json:"id" gorm:"primaryKey;size:64"`
	UserID          string `json:"userId" gorm:"index;size:64"`
	UserDisplayName string `json:"userDisplayName" gorm:"->;-:migration"`
	Endpoint        string `json:"endpoint" gorm:"index;size:191"`
	Method          string `json:"method" gorm:"size:16"`
	Model           string `json:"model" gorm:"index;size:128"`
	ChannelID       string `json:"channelId" gorm:"index;size:64"`
	ChannelName     string `json:"channelName" gorm:"size:191"`
	Status          int    `json:"status" gorm:"index"`
	DurationMs      int64  `json:"durationMs"`
	Credits         int    `json:"credits"`
	RequestBody     string `json:"requestBody" gorm:"type:mediumtext"`
	ResponseBody    string `json:"responseBody" gorm:"type:mediumtext"`
	Error           string `json:"error" gorm:"type:mediumtext"`
	CreatedAt       string `json:"createdAt" gorm:"index;size:64"`
}

type AICallLogList struct {
	Items []AICallLog `json:"items"`
	Total int         `json:"total"`
}
