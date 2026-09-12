package model

type VideoGenerationLog struct {
	ID          string `json:"id" gorm:"primaryKey;size:64"`
	UserID      string `json:"userId" gorm:"index;index:idx_video_generation_logs_user_deleted_created,priority:1;size:64"`
	TaskID      string `json:"taskId" gorm:"index;size:64"`
	VideoID     string `json:"videoId" gorm:"index;size:64"`
	Status      string `json:"status" gorm:"index;size:32"`
	PayloadJSON string `json:"payloadJson" gorm:"type:mediumtext"`
	CreatedAt   string `json:"createdAt" gorm:"index;index:idx_video_generation_logs_user_deleted_created,priority:3;size:64"`
	UpdatedAt   string `json:"updatedAt" gorm:"index;size:64"`
	DeletedAt   string `json:"deletedAt" gorm:"index;index:idx_video_generation_logs_user_deleted_created,priority:2;size:64"`
}

type ImageGenerationLog struct {
	ID          string `json:"id" gorm:"primaryKey;size:64"`
	UserID      string `json:"userId" gorm:"index;index:idx_image_generation_logs_user_deleted_created,priority:1;size:64"`
	TaskID      string `json:"taskId" gorm:"index;size:64"`
	ImageID     string `json:"imageId" gorm:"index;size:64"`
	Status      string `json:"status" gorm:"index;size:32"`
	PayloadJSON string `json:"payloadJson" gorm:"type:mediumtext"`
	CreatedAt   string `json:"createdAt" gorm:"index;index:idx_image_generation_logs_user_deleted_created,priority:3;size:64"`
	UpdatedAt   string `json:"updatedAt" gorm:"index;size:64"`
	DeletedAt   string `json:"deletedAt" gorm:"index;index:idx_image_generation_logs_user_deleted_created,priority:2;size:64"`
}
