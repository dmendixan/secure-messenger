package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite" // 👈 используем вместо gorm.io/driver/sqlite
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"secure-messenger/config"
	"secure-messenger/internal/models"
	"secure-messenger/internal/services"
	"strings"
	"testing"
)

func setupTestDB() *gorm.DB {
	config.JWTSecret = "testsecret"
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	_ = db.AutoMigrate(&models.User{}, &models.RefreshToken{}) // ✅ теперь обе таблицы
	return db
}

func setupTestEnv() {
	config.JWTSecret = "testsecret"
}
func TestRegister(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestEnv()

	db := setupTestDB()
	router := gin.Default()

	// Используем обёртку RegisterWithDB
	router.POST("/register", RegisterWithDB(db))

	payload := `{
		"name": "Test User",
		"email": "test@example.com",
		"password": "123456"
	}`

	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "User registered successfully", response["message"])
}
func TestLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestEnv()

	db := setupTestDB()
	router := gin.Default()

	// Регистрируем пользователя напрямую в БД
	password := "mypassword"
	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	user := models.User{
		Name:         "Test User",
		Email:        "login@example.com",
		PasswordHash: string(hashed),
		Role:         "user",
	}
	db.Create(&user)

	// Регистрируем endpoint логина
	router.POST("/login", func(c *gin.Context) {
		configBackup := config.DB // если ты всё ещё используешь config.DB в Login
		config.DB = db
		defer func() { config.DB = configBackup }()

		Login(c)
	})

	// Формируем запрос логина
	payload := `{
		"email": "login@example.com",
		"password": "mypassword"
	}`

	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Contains(t, resp, "access_token")
	assert.Contains(t, resp, "refresh_token")
}
func TestRefreshToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestEnv()

	db := setupTestDB()
	router := gin.Default()

	// Создаём пользователя
	user := models.User{
		Name:         "Refresh User",
		Email:        "refresh@example.com",
		PasswordHash: "fake-hash", // не нужен для этого теста
		Role:         "user",
	}
	db.Create(&user)

	// Генерируем refresh token
	refreshToken, err := services.GenerateRefreshToken(db, user.ID)

	assert.NoError(t, err)
	assert.NotEmpty(t, refreshToken)

	// Регистрируем endpoint /refresh
	router.POST("/refresh", func(c *gin.Context) {
		configBackup := config.DB
		config.DB = db
		defer func() { config.DB = configBackup }()

		Refresh(c)
	})

	// Формируем запрос
	payload := fmt.Sprintf(`{"refresh_token": "%s"}`, refreshToken)
	req := httptest.NewRequest(http.MethodPost, "/refresh", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Contains(t, resp, "access_token")
	assert.Contains(t, resp, "refresh_token")
}
func TestProfileAccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestEnv()

	db := setupTestDB()
	router := gin.Default()

	// Создаём пользователя
	user := models.User{
		Name:         "Profile User",
		Email:        "profile@example.com",
		PasswordHash: "doesnt-matter", // пароль не нужен
		Role:         "user",
	}
	db.Create(&user)

	// Генерируем access token
	accessToken, err := services.GenerateJWT(user.ID, user.Role)

	assert.NoError(t, err)
	assert.NotEmpty(t, accessToken)

	// Регистрируем /profile endpoint
	router.GET("/profile", AuthMiddleware(""), ProfileHandler(db))

	// Создаём GET-запрос с заголовком Authorization
	req := httptest.NewRequest(http.MethodGet, "/profile", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, user.Email, resp["email"])
}
func TestAdminGetAllUsers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestEnv()

	db := setupTestDB()
	router := gin.Default()

	// Создаём admin-пользователя
	admin := models.User{
		Name:         "Admin User",
		Email:        "admin@example.com",
		PasswordHash: "doesnt-matter",
		Role:         "admin",
	}
	db.Create(&admin)

	// Добавим ещё одного обычного пользователя для проверки списка
	user := models.User{
		Name:         "Regular User",
		Email:        "user@example.com",
		PasswordHash: "irrelevant",
		Role:         "user",
	}
	db.Create(&user)

	// Генерируем access token для admin
	accessToken, err := services.GenerateJWT(admin.ID, admin.Role) // ✅ role: admin

	assert.NoError(t, err)

	// Регистрируем /admin/users endpoint
	adminGroup := router.Group("/admin")
	adminGroup.Use(AuthMiddleware("admin")) // только админ
	{
		adminGroup.GET("/users", GetAllUsersWithDB(db))

	}

	// Создаём GET-запрос с admin-токеном
	req := httptest.NewRequest(http.MethodGet, "/admin/users", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp []map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)

	// Ожидаем минимум двух пользователей
	assert.GreaterOrEqual(t, len(resp), 2)
}
func TestAdminDeleteUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestEnv()
	db := setupTestDB()
	router := gin.Default()

	// Создаём admin-пользователя
	admin := models.User{
		Name:         "Admin",
		Email:        "admin@delete.com",
		PasswordHash: "irrelevant",
		Role:         "admin",
	}
	db.Create(&admin)

	// Создаём обычного пользователя, которого будем удалять
	user := models.User{
		Name:         "ToDelete",
		Email:        "todelete@example.com",
		PasswordHash: "irrelevant",
		Role:         "user",
	}
	db.Create(&user)

	// Генерируем admin access token
	token, err := services.GenerateJWT(admin.ID, admin.Role)
	assert.NoError(t, err)

	// Регистрируем DELETE endpoint
	adminGroup := router.Group("/admin")
	adminGroup.Use(AuthMiddleware("admin"))
	{
		adminGroup.DELETE("/users/:id", DeleteUserWithDB(db))
	}

	// Создаём DELETE-запрос
	url := fmt.Sprintf("/admin/users/%d", user.ID)
	req := httptest.NewRequest(http.MethodDelete, url, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "User deleted successfully", resp["message"])

	// Проверка, что пользователь действительно удалён
	var deleted models.User
	err = db.First(&deleted, user.ID).Error
	assert.Error(t, err) // должен вернуть ошибку: record not found
}
func TestAdminUpdateUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestEnv()
	db := setupTestDB()
	router := gin.Default()

	// Админ
	admin := models.User{
		Name:         "Admin",
		Email:        "admin@update.com",
		PasswordHash: "irrelevant",
		Role:         "admin",
	}
	db.Create(&admin)

	// Пользователь, которого будем обновлять
	user := models.User{
		Name:         "Old Name",
		Email:        "old@example.com",
		PasswordHash: "irrelevant",
		Role:         "user",
	}
	db.Create(&user)

	// Токен для админа
	token, err := services.GenerateJWT(admin.ID, admin.Role)
	assert.NoError(t, err)

	// Регистрируем endpoint
	adminGroup := router.Group("/admin")
	adminGroup.Use(AuthMiddleware("admin"))
	{
		adminGroup.PUT("/users/:id", UpdateUserWithDB(db))
	}

	// JSON для обновления
	payload := `{
		"name": "New Name",
		"email": "new@example.com",
		"role": "admin"
	}`

	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/admin/users/%d", user.ID), strings.NewReader(payload))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "User updated successfully", resp["message"])

	// Проверка изменений в базе
	var updated models.User
	err = db.First(&updated, user.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, "New Name", updated.Name)
	assert.Equal(t, "new@example.com", updated.Email)
	assert.Equal(t, "admin", updated.Role)
}
