package model

type AssetType string

const (
	AssetTypeText  AssetType = "text"
	AssetTypeImage AssetType = "image"
)

// Asset 素材记录。
type Asset struct {
	ID          string    `json:"id" gorm:"primaryKey;size:64"`
	Title       string    `json:"title" gorm:"size:512"`
	Type        AssetType `json:"type" gorm:"size:16"`
	CoverURL    string    `json:"coverUrl" gorm:"size:512"`
	Tags        []string  `json:"tags" gorm:"serializer:json"`
	Category    string    `json:"category" gorm:"size:64"`
	Description string    `json:"description" gorm:"size:1024"`
	Content     string    `json:"content,omitempty" gorm:"type:text"`
	URL         string    `json:"url,omitempty" gorm:"size:1024"`
	CreatedAt   string    `json:"createdAt" gorm:"size:64"`
	UpdatedAt   string    `json:"updatedAt" gorm:"size:64"`
}

// AssetList 素材分页结果。
type AssetList struct {
	Items []Asset  `json:"items"`
	Tags  []string `json:"tags"`
	Total int      `json:"total"`
}
