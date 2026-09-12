package model

// StorageObject 存储对象（S3/R2 文件索引）。
type StorageObject struct {
	ID         string `json:"id" gorm:"primaryKey;size:64"`
	ProviderID string `json:"providerId" gorm:"index;size:64"`
	Bucket     string `json:"bucket" gorm:"size:128"`
	ObjectKey  string `json:"objectKey" gorm:"uniqueIndex;size:512"`
	PublicURL  string `json:"publicUrl" gorm:"size:1024"`
	MimeType   string `json:"mimeType" gorm:"size:128"`
	Bytes      int64  `json:"bytes"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	SHA256     string `json:"sha256" gorm:"size:64"`
	CreatedBy  string `json:"createdBy" gorm:"index;size:64"`
	CreatedAt  string `json:"createdAt" gorm:"size:64"`
	DeletedAt  string `json:"deletedAt" gorm:"size:64"`
}
