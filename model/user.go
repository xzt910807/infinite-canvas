package model

type UserRole string

const (
	UserRoleGuest UserRole = "guest"
	UserRoleUser  UserRole = "user"
	UserRoleAdmin UserRole = "admin"
)

type UserStatus string

const (
	UserStatusActive UserStatus = "active"
	UserStatusBan    UserStatus = "ban"
)

// User 系统用户。
type User struct {
	ID          string     `json:"id" gorm:"primaryKey;size:64"`
	Username    string     `json:"username" gorm:"uniqueIndex;size:64"`
	Password    string     `json:"password,omitempty" gorm:"size:255"`
	Email       string     `json:"email" gorm:"size:191"`
	DisplayName string     `json:"displayName" gorm:"size:191"`
	AvatarURL   string     `json:"avatarUrl" gorm:"size:512"`
	Role        UserRole   `json:"role" gorm:"size:16"`
	Credits     int        `json:"credits"`
	AffCode     string     `json:"affCode" gorm:"uniqueIndex;size:32"`
	AffCount    int        `json:"affCount"`
	InviterID   string     `json:"inviterId" gorm:"size:64"`
	GithubID    string     `json:"githubId" gorm:"size:64"`
	LinuxDoID   string     `json:"linuxDoId" gorm:"index;size:64"`
	WechatID    string     `json:"wechatId" gorm:"size:64"`
	NewApiID    string     `json:"newApiId" gorm:"index;size:64"`
	NewApiToken string     `json:"newApiToken,omitempty" gorm:"index;size:255"`
	Status      UserStatus `json:"status" gorm:"size:16"`
	LastLoginAt string     `json:"lastLoginAt" gorm:"size:64"`
	Extra       string     `json:"extra" gorm:"type:text"`
	CreatedAt   string     `json:"createdAt" gorm:"size:64"`
	UpdatedAt   string     `json:"updatedAt" gorm:"size:64"`
}

// UserList 用户分页结果。
type UserList struct {
	Items []User `json:"items"`
	Total int    `json:"total"`
}

// AuthUser 用户公开信息。
type AuthUser struct {
	ID          string   `json:"id"`
	Username    string   `json:"username"`
	DisplayName string   `json:"displayName"`
	AvatarURL   string   `json:"avatarUrl"`
	Role        UserRole `json:"role"`
	Credits     int      `json:"credits"`
	CreatedAt   string   `json:"createdAt"`
	UpdatedAt   string   `json:"updatedAt"`
}

// AuthSession 登录会话信息。
type AuthSession struct {
	Token string   `json:"token"`
	User  AuthUser `json:"user"`
}

func PublicUser(user User) AuthUser {
	return AuthUser{
		ID:          user.ID,
		Username:    user.Username,
		DisplayName: user.DisplayName,
		AvatarURL:   user.AvatarURL,
		Role:        user.Role,
		Credits:     user.Credits,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}
}

type CreditLogType string

const (
	CreditLogTypeAdminAdjust CreditLogType = "admin_adjust"
	CreditLogTypeAIConsume   CreditLogType = "ai_consume"
	CreditLogTypeAIRefund    CreditLogType = "ai_refund"
)

// CreditLog 用户算力点变更流水。
type CreditLog struct {
	ID        string        `json:"id" gorm:"primaryKey;size:64"`
	UserID    string        `json:"userId" gorm:"index;size:64"`
	Type      CreditLogType `json:"type" gorm:"size:32"`
	Amount    int           `json:"amount"`
	Balance   int           `json:"balance"`
	RelatedID string        `json:"relatedId" gorm:"size:64"`
	Remark    string        `json:"remark" gorm:"size:512"`
	Extra     string        `json:"extra" gorm:"type:text"`
	CreatedAt string        `json:"createdAt" gorm:"size:64"`
}

type CreditLogList struct {
	Items []CreditLog `json:"items"`
	Total int         `json:"total"`
}
