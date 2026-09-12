package model

// Prompt 提示词记录。
type Prompt struct {
	ID        string   `json:"id" gorm:"primaryKey;size:64"`
	// title/coverUrl 来自 GitHub 提示词同步，长度不可控（实测 title 最长 3200+），用 text。
	Title     string   `json:"title" gorm:"type:text"`
	CoverURL  string   `json:"coverUrl" gorm:"type:text"`
	Prompt    string   `json:"prompt" gorm:"type:text"`
	Tags      []string `json:"tags" gorm:"serializer:json"`
	Category  string   `json:"category" gorm:"index;size:64"`
	GithubURL string   `json:"githubUrl" gorm:"-"`
	Preview   string   `json:"preview" gorm:"type:text"`
	CreatedAt string   `json:"createdAt" gorm:"size:64"`
	UpdatedAt string   `json:"updatedAt" gorm:"size:64"`
}

// PromptList 提示词分页结果。
type PromptList struct {
	Items      []Prompt `json:"items"`
	Tags       []string `json:"tags"`
	Categories []string `json:"categories"`
	Total      int      `json:"total"`
}

// PromptCategory 提示词分类。
type PromptCategory struct {
	Category    string `json:"category" gorm:"primaryKey"`
	Name        string `json:"name"`
	Description string `json:"description"`
	GithubURL   string `json:"githubUrl"`
	Remote      bool   `json:"remote"`
	UpdatedAt   string `json:"updatedAt"`
}
