package service

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/tigerowo/infinite-canvas/config"
	"github.com/tigerowo/infinite-canvas/model"
	"github.com/tigerowo/infinite-canvas/repository"
	"golang.org/x/crypto/bcrypt"
)

// TestMain 在进程级初始化 SSO 测试所需的配置与内存数据库。
// repository.DB() 是 sync.Once 单例，必须在任何测试触发它之前
// 先把 DSN 指向内存库，否则会意外创建 data/infinite-canvas.db。
func TestMain(m *testing.M) {
	config.Cfg.SSOSecret = "test-sso-secret"
	config.Cfg.JWTSecret = "test-jwt-secret"
	config.Cfg.JWTExpireHours = 1
	config.Cfg.StorageDriver = "sqlite"
	config.Cfg.DatabaseDSN = "file:sso_login_test?mode=memory&cache=shared"
	m.Run()
}

func clearSSOTestUsers(t *testing.T) {
	t.Helper()
	db, err := repository.DB()
	if err != nil {
		t.Fatalf("failed to init test db: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Exec("DELETE FROM users").Error; err != nil {
			t.Fatalf("failed to clear users: %v", err)
		}
	})
}

// TestValidateSSOTokenParsesPasswordHash 验证 SSO JWT 中的 password_hash
// claim 被完整解析，供 LoginWithSSO 对齐两侧密码。
func TestValidateSSOTokenParsesPasswordHash(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("shared-password"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	type ssoJWTClaimsForTest struct {
		Username     string `json:"username"`
		Email        string `json:"email"`
		DisplayName  string `json:"display_name"`
		PasswordHash string `json:"password_hash"`
		APIToken     string `json:"api_token"`
		jwt.RegisteredClaims
	}
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, ssoJWTClaimsForTest{
		Username:     " alice ",
		Email:        "alice@example.com",
		DisplayName:  "Alice",
		PasswordHash: string(hash) + " ",
		APIToken:     "sk-test",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "new-api",
			Subject:   "42",
			Audience:  jwt.ClaimStrings{"infinite-canvas"},
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Minute)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	})
	tokenString, err := token.SignedString([]byte("test-sso-secret"))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	claims, err := ValidateSSOToken(tokenString)
	if err != nil {
		t.Fatalf("ValidateSSOToken returned error: %v", err)
	}
	if claims.NewApiID != "42" {
		t.Fatalf("NewApiID = %q, want 42", claims.NewApiID)
	}
	if claims.Username != "alice" {
		t.Fatalf("Username = %q, want alice", claims.Username)
	}
	if claims.PasswordHash != string(hash) {
		t.Fatalf("PasswordHash = %q, want %q", claims.PasswordHash, string(hash))
	}
}

// TestLoginWithSSOAlignsUsernameAndPassword 验证画布账号的用户名与密码
// 在 SSO 登录时与 new-api 对齐。
func TestLoginWithSSOAlignsUsernameAndPassword(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("shared-password"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	t.Run("new user adopts new-api username and password hash", func(t *testing.T) {
		clearSSOTestUsers(t)

		_, err := LoginWithSSO(SSOClaims{
			NewApiID:     "1",
			Username:     "alice",
			Email:        "alice@example.com",
			DisplayName:  "Alice",
			PasswordHash: string(hash),
		})
		if err != nil {
			t.Fatalf("LoginWithSSO returned error: %v", err)
		}

		user, ok, err := repository.GetUserByNewApiID("1")
		if err != nil || !ok {
			t.Fatalf("failed to load created user: ok=%v err=%v", ok, err)
		}
		if user.Username != "alice" {
			t.Fatalf("Username = %q, want alice", user.Username)
		}
		if user.Password != string(hash) {
			t.Fatalf("Password hash = %q, want the new-api hash", user.Password)
		}
		if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte("shared-password")) != nil {
			t.Fatal("the new-api password must verify against the stored hash")
		}
	})

	t.Run("taken username falls back to newapi id", func(t *testing.T) {
		clearSSOTestUsers(t)

		if _, err := repository.SaveUser(model.User{
			ID:       "user-local",
			Username: "alice",
			Password: "local-hash",
			Role:     model.UserRoleUser,
			Status:   model.UserStatusActive,
		}); err != nil {
			t.Fatalf("failed to seed local user: %v", err)
		}

		if _, err := LoginWithSSO(SSOClaims{
			NewApiID:     "2",
			Username:     "alice",
			PasswordHash: string(hash),
		}); err != nil {
			t.Fatalf("LoginWithSSO returned error: %v", err)
		}

		user, ok, err := repository.GetUserByNewApiID("2")
		if err != nil || !ok {
			t.Fatalf("failed to load created user: ok=%v err=%v", ok, err)
		}
		if user.Username != "newapi-2" {
			t.Fatalf("Username = %q, want newapi-2", user.Username)
		}
		if user.Password != string(hash) {
			t.Fatalf("Password hash = %q, want the new-api hash even with fallback username", user.Password)
		}
	})

	t.Run("existing legacy user syncs username and password", func(t *testing.T) {
		clearSSOTestUsers(t)

		legacyHash, err := bcrypt.GenerateFromPassword([]byte("random-legacy"), bcrypt.MinCost)
		if err != nil {
			t.Fatalf("failed to hash legacy password: %v", err)
		}
		if _, err := repository.SaveUser(model.User{
			ID:       "user-legacy",
			Username: "newapi-3",
			Password: string(legacyHash),
			Role:     model.UserRoleUser,
			NewApiID: "3",
			Status:   model.UserStatusActive,
		}); err != nil {
			t.Fatalf("failed to seed legacy user: %v", err)
		}

		if _, err := LoginWithSSO(SSOClaims{
			NewApiID:     "3",
			Username:     "bob",
			PasswordHash: string(hash),
		}); err != nil {
			t.Fatalf("LoginWithSSO returned error: %v", err)
		}

		user, ok, err := repository.GetUserByNewApiID("3")
		if err != nil || !ok {
			t.Fatalf("failed to load synced user: ok=%v err=%v", ok, err)
		}
		if user.Username != "bob" {
			t.Fatalf("Username = %q, want bob synced from new-api", user.Username)
		}
		if user.Password != string(hash) {
			t.Fatalf("Password hash = %q, want the new-api hash", user.Password)
		}
	})

	t.Run("conflicting rename keeps current username but still syncs password", func(t *testing.T) {
		clearSSOTestUsers(t)

		if _, err := repository.SaveUser(model.User{
			ID:       "user-other",
			Username: "carol",
			Password: "other-hash",
			AffCode:  "AFFOTHER",
			Role:     model.UserRoleUser,
			Status:   model.UserStatusActive,
		}); err != nil {
			t.Fatalf("failed to seed conflicting user: %v", err)
		}
		if _, err := repository.SaveUser(model.User{
			ID:       "user-sso",
			Username: "newapi-4",
			Password: "legacy-hash",
			AffCode:  "AFFSSO4",
			Role:     model.UserRoleUser,
			NewApiID: "4",
			Status:   model.UserStatusActive,
		}); err != nil {
			t.Fatalf("failed to seed sso user: %v", err)
		}

		if _, err := LoginWithSSO(SSOClaims{
			NewApiID:     "4",
			Username:     "carol",
			PasswordHash: string(hash),
		}); err != nil {
			t.Fatalf("LoginWithSSO returned error: %v", err)
		}

		user, ok, err := repository.GetUserByNewApiID("4")
		if err != nil || !ok {
			t.Fatalf("failed to load synced user: ok=%v err=%v", ok, err)
		}
		if user.Username != "newapi-4" {
			t.Fatalf("Username = %q, want newapi-4 kept when target name is taken", user.Username)
		}
		if user.Password != string(hash) {
			t.Fatalf("Password hash = %q, want the new-api hash", user.Password)
		}
	})
}
