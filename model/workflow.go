package model

// CreativeWorkflow 创意工作流模板。
type CreativeWorkflow struct {
	ID          string `json:"id" gorm:"primaryKey;size:64"`
	OwnerUserID string `json:"ownerUserId" gorm:"index;size:64"`
	Scope       string `json:"scope" gorm:"index;size:16"` // "private" | "public"
	Name        string `json:"name" gorm:"index;size:191"`
	Category    string `json:"category" gorm:"index;size:64"`
	Description string `json:"description" gorm:"size:1024"`
	Data        string `json:"data" gorm:"type:text"` // JSON: variables + config
	CreatedAt   string `json:"createdAt" gorm:"size:64"`
	UpdatedAt   string `json:"updatedAt" gorm:"size:64"`
	LastRunAt   string `json:"lastRunAt" gorm:"size:64"`
}
