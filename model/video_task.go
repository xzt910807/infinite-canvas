package model

type VideoTask struct {
	ID              string `json:"id" gorm:"primaryKey;size:64"`
	UserID          string `json:"userId" gorm:"index;size:64"`
	UserDisplayName string `json:"userDisplayName" gorm:"size:191"`
	Model           string `json:"model" gorm:"index;size:128"`
	ChannelID       string `json:"channelId" gorm:"index;size:64"`
	UserChannelID   string `json:"userChannelId" gorm:"index;size:64"`
	ChannelName     string `json:"channelName" gorm:"size:191"`
	Source          string `json:"source" gorm:"index;size:32"`
	SourceID        string `json:"source_id" gorm:"index;size:64"`
	UpstreamTaskID  string `json:"upstreamTaskId" gorm:"index;size:64"`
	UpstreamVideoID string `json:"upstreamVideoId" gorm:"index;size:64"`
	Status          string `json:"status" gorm:"index:idx_video_tasks_status_created_at,priority:1;size:32"`
	Progress        int    `json:"progress"`
	Seconds         string `json:"seconds" gorm:"size:16"`
	Size            string `json:"size" gorm:"size:32"`
	VideoURL        string `json:"videoUrl" gorm:"type:text"`
	Error           string `json:"error" gorm:"type:text"`
	ErrorDetail     string `json:"errorDetail" gorm:"type:text"`
	RequestBody     string `json:"requestBody" gorm:"type:text"`
	ResponseBody    string `json:"responseBody" gorm:"type:text"`
	LastResponse    string `json:"lastResponse" gorm:"type:text"`
	Credits         int    `json:"credits"`
	CreatedAt       string `json:"createdAt" gorm:"index;index:idx_video_tasks_status_created_at,priority:2;size:64"`
	UpdatedAt       string `json:"updatedAt" gorm:"index;size:64"`
	StartedAt       string `json:"startedAt" gorm:"size:64"`
	CompletedAt     string `json:"completedAt" gorm:"size:64"`
	LastPolledAt    string `json:"lastPolledAt" gorm:"index;size:64"`
}
