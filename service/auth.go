package service

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/tigerowo/infinite-canvas/config"
	"github.com/tigerowo/infinite-canvas/model"
	"github.com/tigerowo/infinite-canvas/repository"
	"golang.org/x/crypto/bcrypt"
)

type TokenClaims struct {
	UserID   string         `json:"userId"`
	Username string         `json:"username"`
	Role     model.UserRole `json:"role"`
	jwt.RegisteredClaims
}

type userExtra struct {
	LinuxDo any `json:"linuxDo,omitempty"`
}

func EnsureDefaultAdmin() error {
	if strings.TrimSpace(config.Cfg.AdminUsername) == "" || strings.TrimSpace(config.Cfg.AdminPassword) == "" {
		return nil
	}
	WarnDefaultSecurityConfig()
	hasAdmin, err := repository.HasAdmin()
	if err != nil || hasAdmin {
		return err
	}
	hash, err := hashPassword(config.Cfg.AdminPassword)
	if err != nil {
		return err
	}
	_, err = repository.SaveUser(model.User{
		ID:        newID("user"),
		Username:  strings.TrimSpace(config.Cfg.AdminUsername),
		Password:  hash,
		Role:      model.UserRoleAdmin,
		AffCode:   newAffCode(),
		Status:    model.UserStatusActive,
		CreatedAt: now(),
		UpdatedAt: now(),
	})
	return err
}

func Register(username string, password string) (model.AuthSession, error) {
	settings, err := repository.GetSettings()
	if err != nil {
		return model.AuthSession{}, err
	}
	normalizedSettings := normalizeSettings(settings)
	if normalizedSettings.Public.Auth.AllowRegister != nil && !*normalizedSettings.Public.Auth.AllowRegister {
		return model.AuthSession{}, safeMessageError{message: "当前未开放注册"}
	}
	username = strings.TrimSpace(username)
	if strings.ContainsAny(username, " \t\r\n") {
		return model.AuthSession{}, safeMessageError{message: "用户名不能包含空格"}
	}
	if username == "" || password == "" {
		return model.AuthSession{}, safeMessageError{message: "用户名和密码不能为空"}
	}
	if _, ok, err := repository.GetUserByUsername(username); err != nil || ok {
		if err != nil {
			return model.AuthSession{}, err
		}
		return model.AuthSession{}, safeMessageError{message: "用户名已存在"}
	}
	hash, err := hashPassword(password)
	if err != nil {
		return model.AuthSession{}, err
	}
	user, err := repository.SaveUser(model.User{
		ID:        newID("user"),
		Username:  username,
		Password:  hash,
		Role:      model.UserRoleUser,
		AffCode:   newAffCode(),
		Status:    model.UserStatusActive,
		CreatedAt: now(),
		UpdatedAt: now(),
	})
	if err != nil {
		return model.AuthSession{}, err
	}
	return newSession(user)
}

func Login(username string, password string) (model.AuthSession, error) {
	user, ok, err := repository.GetUserByUsername(strings.TrimSpace(username))
	if err != nil {
		return model.AuthSession{}, err
	}
	if !ok || bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
		return model.AuthSession{}, safeMessageError{message: "用户名或密码错误"}
	}
	if user.Status == model.UserStatusBan {
		return model.AuthSession{}, safeMessageError{message: "账号已被禁用"}
	}
	normalizeUserDefaults(&user)
	user.LastLoginAt = now()
	user.UpdatedAt = now()
	user, err = repository.SaveUser(user)
	if err != nil {
		return model.AuthSession{}, err
	}
	return newSession(user)
}

func LinuxDoAuthorizeURL(r *http.Request, redirect string) (string, error) {
	settings, err := repository.GetSettings()
	if err != nil {
		return "", err
	}
	settings = normalizeSettings(settings)
	linuxDo := settings.Private.Auth.LinuxDo
	if !settings.Public.Auth.LinuxDo.Enabled {
		return "", safeMessageError{message: "Linux.do 登录未开启"}
	}
	if strings.TrimSpace(linuxDo.ClientID) == "" || strings.TrimSpace(linuxDo.ClientSecret) == "" {
		return "", safeMessageError{message: "Linux.do 登录未配置"}
	}
	values := url.Values{}
	values.Set("client_id", linuxDo.ClientID)
	values.Set("redirect_uri", linuxDoRedirectURI(r))
	values.Set("response_type", "code")
	values.Set("scope", "read")
	values.Set("state", base64.RawURLEncoding.EncodeToString([]byte(redirect)))
	return config.Cfg.LinuxDoAuthorizeURL + "?" + values.Encode(), nil
}

func LoginWithLinuxDo(r *http.Request, code string, state string) (model.AuthSession, string, error) {
	redirect := decodeState(state)
	settings, err := repository.GetSettings()
	if err != nil {
		return model.AuthSession{}, redirect, err
	}
	settings = normalizeSettings(settings)
	linuxDo := settings.Private.Auth.LinuxDo
	if !settings.Public.Auth.LinuxDo.Enabled {
		return model.AuthSession{}, redirect, safeMessageError{message: "Linux.do 登录未开启"}
	}
	token, err := linuxDoAccessToken(r, code, linuxDo)
	if err != nil {
		return model.AuthSession{}, redirect, err
	}
	profile, err := linuxDoProfile(token)
	if err != nil {
		return model.AuthSession{}, redirect, err
	}
	linuxDoID := fmt.Sprint(profile.ID)
	if strings.TrimSpace(linuxDoID) == "" || linuxDoID == "0" {
		return model.AuthSession{}, redirect, safeMessageError{message: "Linux.do 用户信息无效"}
	}
	user, ok, err := repository.GetUserByLinuxDoID(linuxDoID)
	if err != nil {
		return model.AuthSession{}, redirect, err
	}
	if !ok {
		if settings.Public.Auth.AllowRegister != nil && !*settings.Public.Auth.AllowRegister {
			return model.AuthSession{}, redirect, safeMessageError{message: "当前未开放注册"}
		}
		user = model.User{
			ID:          newID("user"),
			Username:    linuxDoUsername(profile.Username, linuxDoID),
			DisplayName: strings.TrimSpace(profile.Name),
			AvatarURL:   linuxDoAvatar(profile.AvatarTemplate),
			Role:        model.UserRoleUser,
			AffCode:     newAffCode(),
			LinuxDoID:   linuxDoID,
			Status:      model.UserStatusActive,
			CreatedAt:   now(),
		}
	} else if user.Status == model.UserStatusBan {
		return model.AuthSession{}, redirect, safeMessageError{message: "账号已被禁用"}
	}
	user.DisplayName = firstNonEmpty(profile.Name, user.DisplayName)
	user.AvatarURL = firstNonEmpty(linuxDoAvatar(profile.AvatarTemplate), user.AvatarURL)
	user.LastLoginAt = now()
	user.UpdatedAt = now()
	extra, _ := json.Marshal(userExtra{LinuxDo: profile})
	user.Extra = string(extra)
	user, err = repository.SaveUser(user)
	if err != nil {
		return model.AuthSession{}, redirect, err
	}
	session, err := newSession(user)
	return session, redirect, err
}

func ParseToken(tokenText string) (TokenClaims, error) {
	claims := TokenClaims{}
	token, err := jwt.ParseWithClaims(tokenText, &claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("登录状态无效")
		}
		return []byte(config.Cfg.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return TokenClaims{}, errors.New("登录状态无效")
	}
	return claims, nil
}

func CurrentAuthUser(tokenText string) (model.AuthUser, bool) {
	claims, err := ParseToken(tokenText)
	if err != nil {
		return model.AuthUser{}, false
	}
	user, ok, err := repository.GetUserByID(claims.UserID)
	if err != nil || !ok {
		return model.AuthUser{}, false
	}
	if user.Status == model.UserStatusBan {
		return model.AuthUser{}, false
	}
	return model.PublicUser(user), true
}

func ListUsers(q model.Query) (model.UserList, error) {
	users, total, err := repository.ListUsers(q)
	if err != nil {
		return model.UserList{}, err
	}
	for i := range users {
		users[i].Password = ""
		normalizeUserDefaults(&users[i])
	}
	return model.UserList{Items: users, Total: int(total)}, nil
}

func SaveUser(user model.User, password string) (model.User, error) {
	user.Username = strings.TrimSpace(user.Username)
	if strings.ContainsAny(user.Username, " \t\r\n") {
		return user, safeMessageError{message: "用户名不能包含空格"}
	}
	if user.Username == "" {
		return user, safeMessageError{message: "用户名不能为空"}
	}
	if user.Role == "" || user.Role == model.UserRoleGuest {
		user.Role = model.UserRoleUser
	}
	if user.Status == "" {
		user.Status = model.UserStatusActive
	}
	if saved, ok, err := repository.GetUserByUsername(user.Username); err != nil {
		return user, err
	} else if ok && saved.ID != user.ID {
		return user, safeMessageError{message: "用户名已存在"}
	}
	isCreate := user.ID == ""
	if isCreate {
		user.ID = newID("user")
		user.AffCode = newAffCode()
		user.CreatedAt = now()
	} else if saved, ok, err := repository.GetUserByID(user.ID); err != nil {
		return user, err
	} else if ok {
		user.CreatedAt = saved.CreatedAt
		user.Password = saved.Password
		user.AvatarURL = saved.AvatarURL
		user.Credits = saved.Credits
		user.Extra = saved.Extra
		if user.AffCode == "" {
			user.AffCode = saved.AffCode
		}
		if user.AffCode == "" {
			user.AffCode = newAffCode()
		}
		if user.LinuxDoID == "" {
			user.LinuxDoID = saved.LinuxDoID
		}
		// new-api linkage and its billing token are managed by SSO, never edited
		// from the admin form, so always keep the stored values.
		user.NewApiID = saved.NewApiID
		user.NewApiToken = saved.NewApiToken
		user.LastLoginAt = saved.LastLoginAt
	}
	if password != "" {
		hash, err := hashPassword(password)
		if err != nil {
			return user, err
		}
		user.Password = hash
	}
	if isCreate && user.Password == "" {
		return user, safeMessageError{message: "密码不能为空"}
	}
	user.UpdatedAt = now()
	user, err := repository.SaveUser(user)
	user.Password = ""
	return user, err
}

func AdjustUserCredits(id string, credits int) (model.User, error) {
	user, ok, err := repository.GetUserByID(id)
	if err != nil || !ok {
		if err != nil {
			return user, err
		}
		return user, safeMessageError{message: "用户不存在"}
	}
	oldCredits := user.Credits
	user.Credits = credits
	user.UpdatedAt = now()
	user, err = repository.SaveUser(user)
	if err == nil && oldCredits != credits {
		_, err = repository.SaveCreditLog(model.CreditLog{
			ID:        newID("credit"),
			UserID:    user.ID,
			Type:      model.CreditLogTypeAdminAdjust,
			Amount:    credits - oldCredits,
			Balance:   credits,
			Remark:    "后台手动调整",
			CreatedAt: now(),
		})
	}
	user.Password = ""
	return user, err
}

func ConsumeUserCredits(userID string, modelName string, credits int, path string) error {
	if credits <= 0 {
		return nil
	}
	user, ok, err := repository.ConsumeUserCredits(userID, credits, now())
	if err != nil {
		return err
	}
	if !ok {
		return safeMessageError{message: "算力点不足"}
	}
	extra, _ := json.Marshal(map[string]string{"model": modelName, "path": path})
	_, err = repository.SaveCreditLog(model.CreditLog{
		ID:        newID("credit"),
		UserID:    userID,
		Type:      model.CreditLogTypeAIConsume,
		Amount:    -credits,
		Balance:   user.Credits,
		Remark:    "调用模型 " + modelName,
		Extra:     string(extra),
		CreatedAt: now(),
	})
	return err
}

func RefundUserCredits(userID string, modelName string, credits int, path string) error {
	if credits <= 0 {
		return nil
	}
	user, ok, err := repository.RefundUserCredits(userID, credits, now())
	if err != nil {
		return err
	}
	if !ok {
		return safeMessageError{message: "用户不存在"}
	}
	extra, _ := json.Marshal(map[string]string{"model": modelName, "path": path})
	_, err = repository.SaveCreditLog(model.CreditLog{
		ID:        newID("credit"),
		UserID:    userID,
		Type:      model.CreditLogTypeAIRefund,
		Amount:    credits,
		Balance:   user.Credits,
		Remark:    "模型调用失败返还 " + modelName,
		Extra:     string(extra),
		CreatedAt: now(),
	})
	return err
}

func ListCreditLogs(q model.Query) (model.CreditLogList, error) {
	logs, total, err := repository.ListCreditLogs(q)
	if err != nil {
		return model.CreditLogList{}, err
	}
	return model.CreditLogList{Items: logs, Total: int(total)}, nil
}

func SaveCreditLog(log model.CreditLog) (model.CreditLog, error) {
	if log.ID == "" {
		log.ID = newID("credit")
		log.CreatedAt = now()
	}
	return repository.SaveCreditLog(log)
}

func DeleteCreditLog(id string) error {
	return repository.DeleteCreditLog(id)
}

func DeleteUser(id string) error {
	return repository.DeleteUser(id)
}

func GuestUser() model.AuthUser {
	return model.AuthUser{ID: "", Username: "guest", Role: model.UserRoleGuest}
}

func newSession(user model.User) (model.AuthSession, error) {
	token, err := newToken(user)
	if err != nil {
		return model.AuthSession{}, err
	}
	return model.AuthSession{Token: token, User: model.PublicUser(user)}, nil
}

func newToken(user model.User) (string, error) {
	expireHours := config.Cfg.JWTExpireHours
	if expireHours <= 0 {
		expireHours = 168
	}
	claims := TokenClaims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expireHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.ID,
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(config.Cfg.JWTSecret))
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

func now() string {
	return time.Now().Format(time.RFC3339)
}

func newID(prefix string) string {
	return prefix + "-" + uuid.NewString()
}

func newAffCode() string {
	return strings.ToUpper(strings.ReplaceAll(uuid.NewString()[:8], "-", ""))
}

func normalizeUserDefaults(user *model.User) {
	if user.Status == "" {
		user.Status = model.UserStatusActive
	}
	if user.AffCode == "" {
		user.AffCode = newAffCode()
	}
}

type linuxDoTokenResponse struct {
	AccessToken string `json:"access_token"`
}

type linuxDoUserResponse struct {
	ID             int64  `json:"id"`
	Username       string `json:"username"`
	Name           string `json:"name"`
	AvatarTemplate string `json:"avatar_template"`
}

func linuxDoAccessToken(r *http.Request, code string, setting model.PrivateLinuxDoAuthSetting) (string, error) {
	values := url.Values{}
	values.Set("client_id", setting.ClientID)
	values.Set("client_secret", setting.ClientSecret)
	values.Set("grant_type", "authorization_code")
	values.Set("code", code)
	values.Set("redirect_uri", linuxDoRedirectURI(r))
	req, _ := http.NewRequest(http.MethodPost, config.Cfg.LinuxDoTokenURL, strings.NewReader(values.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	var payload linuxDoTokenResponse
	if err := doLinuxDoJSON(req, &payload); err != nil {
		return "", err
	}
	if strings.TrimSpace(payload.AccessToken) == "" {
		return "", safeMessageError{message: "Linux.do 登录失败"}
	}
	return payload.AccessToken, nil
}

func linuxDoRedirectURI(r *http.Request) string {
	return RequestOrigin(r) + "/api/auth/linux-do/callback"
}

func linuxDoProfile(token string) (linuxDoUserResponse, error) {
	req, _ := http.NewRequest(http.MethodGet, config.Cfg.LinuxDoUserInfoURL, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	var payload linuxDoUserResponse
	err := doLinuxDoJSON(req, &payload)
	return payload, err
}

func doLinuxDoJSON(req *http.Request, payload any) error {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return safeMessageError{message: "Linux.do 登录失败"}
	}
	return json.NewDecoder(bytes.NewReader(body)).Decode(payload)
}

func linuxDoUsername(username string, id string) string {
	base := strings.TrimSpace(username)
	if base == "" {
		base = "linuxdo-" + id
	}
	if _, ok, err := repository.GetUserByUsername(base); err != nil || !ok {
		return base
	}
	return base + "-" + id
}

func linuxDoAvatar(template string) string {
	if strings.TrimSpace(template) == "" {
		return ""
	}
	if strings.HasPrefix(template, "//") {
		template = "https:" + template
	}
	if strings.HasPrefix(template, "/") {
		template = "https://linux.do" + template
	}
	return strings.ReplaceAll(template, "{size}", "120")
}

func decodeState(state string) string {
	data, err := base64.RawURLEncoding.DecodeString(state)
	if err != nil {
		return "/"
	}
	return safeRedirectPath(string(data))
}

// safeRedirectPath 仅放行站内相对路径，拦截开放重定向。浏览器会忽略 URL 中的
// Tab/换行/回车，并把 //host 或 /\host 解析为协议相对的跨站地址，因此先剥离这些
// 控制字符，再拒绝 // 与 /\ 前缀。
func safeRedirectPath(redirect string) string {
	cleaned := strings.Map(func(r rune) rune {
		if r == '\t' || r == '\n' || r == '\r' {
			return -1
		}
		return r
	}, redirect)
	if !strings.HasPrefix(cleaned, "/") || strings.HasPrefix(cleaned, "//") || strings.HasPrefix(cleaned, "/\\") {
		return "/"
	}
	return cleaned
}

func RequestOrigin(r *http.Request) string {
	host := strings.TrimSpace(r.Header.Get("X-Forwarded-Host"))
	if host == "" {
		host = r.Host
	}
	proto := strings.TrimSpace(r.Header.Get("X-Forwarded-Proto"))
	if proto == "" {
		proto = "http"
	}
	return proto + "://" + host
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

// SSOClaims represents the JWT claims sent from new-api during SSO.
type SSOClaims struct {
	NewApiID     string
	Username     string
	Email        string
	DisplayName  string
	PasswordHash string
	APIToken     string
}

// ValidateSSOToken parses and validates the SSO JWT token using the shared secret.
func ValidateSSOToken(tokenText string) (SSOClaims, error) {
	if strings.TrimSpace(config.Cfg.SSOSecret) == "" {
		return SSOClaims{}, safeMessageError{message: "SSO is not configured"}
	}
	type ssoJWTClaims struct {
		Username     string `json:"username"`
		Email        string `json:"email"`
		DisplayName  string `json:"display_name"`
		PasswordHash string `json:"password_hash"`
		APIToken     string `json:"api_token"`
		jwt.RegisteredClaims
	}
	claims := ssoJWTClaims{}
	token, err := jwt.ParseWithClaims(tokenText, &claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(config.Cfg.SSOSecret), nil
	},
		jwt.WithIssuer("new-api"),
		jwt.WithAudience("infinite-canvas"),
		jwt.WithExpirationRequired(),
	)
	if err != nil || !token.Valid {
		return SSOClaims{}, safeMessageError{message: "SSO token is invalid or expired"}
	}
	newApiID := strings.TrimSpace(claims.Subject)
	if newApiID == "" {
		return SSOClaims{}, safeMessageError{message: "SSO token missing user id"}
	}
	return SSOClaims{
		NewApiID:     newApiID,
		Username:     strings.TrimSpace(claims.Username),
		Email:        strings.TrimSpace(claims.Email),
		DisplayName:  strings.TrimSpace(claims.DisplayName),
		PasswordHash: strings.TrimSpace(claims.PasswordHash),
		APIToken:     strings.TrimSpace(claims.APIToken),
	}, nil
}

// LoginWithSSO finds or creates a user from SSO claims and returns a canvas session.
// 画布账号的用户名与密码始终对齐 new-api 侧（两边同为 bcrypt，哈希可直接互通）：
// 新用户直接采用 new-api 的用户名和密码哈希；存量用户在每次 SSO 登录时同步，
// 因此在 new-api 改名或改密码后，重新进入画布即可自动对齐。
func LoginWithSSO(claims SSOClaims) (model.AuthSession, error) {
	user, ok, err := repository.GetUserByNewApiID(claims.NewApiID)
	if err != nil {
		return model.AuthSession{}, err
	}
	displayName := claims.DisplayName
	if displayName == "" {
		displayName = claims.Username
	}
	if !ok {
		// Create a new canvas user linked to the new-api account.
		// 用户名优先用 new-api 的用户名保持两边一致；被画布本地注册账号
		// 占用时回退到 newapi-{id} 派生名。
		username := claims.Username
		if username != "" {
			if _, exists, _ := repository.GetUserByUsername(username); exists {
				username = ""
			}
		}
		if username == "" {
			username = "newapi-" + claims.NewApiID
			if _, exists, _ := repository.GetUserByUsername(username); exists {
				username = username + "-" + uuid.NewString()[:8]
			}
		}
		// 密码直接采用 new-api 的 bcrypt 哈希，两边密码保持一致；
		// 旧版 token 未携带哈希时生成随机密码兜底。
		passwordHash := claims.PasswordHash
		if passwordHash == "" {
			randomPw := uuid.NewString() + uuid.NewString()
			passwordHash, err = hashPassword(randomPw)
			if err != nil {
				return model.AuthSession{}, err
			}
		}
		user = model.User{
			ID:          newID("user"),
			Username:    username,
			Password:    passwordHash,
			DisplayName: displayName,
			Email:       claims.Email,
			Role:        model.UserRoleUser,
			NewApiID:    claims.NewApiID,
			AffCode:     newAffCode(),
			Status:      model.UserStatusActive,
			CreatedAt:   now(),
		}
	} else {
		if user.Status == model.UserStatusBan {
			return model.AuthSession{}, safeMessageError{message: "账号已被禁用"}
		}
		// Update profile from the latest new-api data.
		if displayName != "" {
			user.DisplayName = displayName
		}
		if claims.Email != "" {
			user.Email = claims.Email
		}
		// 同步用户名：目标名被画布内其他账号占用时保留现名，下次 SSO 重试。
		if claims.Username != "" && user.Username != claims.Username {
			if _, exists, _ := repository.GetUserByUsername(claims.Username); !exists {
				user.Username = claims.Username
			}
		}
		if claims.PasswordHash != "" {
			user.Password = claims.PasswordHash
		}
	}
	user.LastLoginAt = now()
	user.UpdatedAt = now()
	// Refresh the new-api API token on every SSO login so user-token billing
	// keeps working after the token is rotated or re-created in new-api.
	if claims.APIToken != "" {
		user.NewApiToken = claims.APIToken
	}
	user, err = repository.SaveUser(user)
	if err != nil {
		return model.AuthSession{}, err
	}
	return newSession(user)
}

// UserNewApiToken 返回用户绑定的 new-api 访问令牌，供用户令牌计费渠道使用。
// 令牌由 SSO 登录时从 new-api 同步，缺失时提示用户重新进入画布。
func UserNewApiToken(userID string) (string, error) {
	user, ok, err := repository.GetUserByID(userID)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", safeMessageError{message: "用户不存在"}
	}
	if strings.TrimSpace(user.NewApiToken) == "" {
		return "", safeMessageError{message: "请从 new-api 重新进入画布以同步访问令牌"}
	}
	return user.NewApiToken, nil
}

func WarnDefaultSecurityConfig() {
	if config.Cfg.AdminUsername == "admin" && config.Cfg.AdminPassword == "infinite-canvas" {
		log.Println("WARNING: using default admin credentials, please set ADMIN_USERNAME and ADMIN_PASSWORD to safer values before deployment")
	}
}
