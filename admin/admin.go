package admin

import (
	"Animals_Shelter/models"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/qor/admin"
	"github.com/qor/qor"
	"log"
	"os"
)

// InitAdmin инициализирует и возвращает QOR Admin
func InitAdmin(db *gorm.DB) *admin.Admin {
	db.SetLogger(log.New(os.Stdout, "\r\n", log.LstdFlags))

	// Настроим уровень логирования для GORM v1 (SQL-запросы)
	db.LogMode(true)

	// Создаём новый экземпляр QOR Admin
	Admin := admin.New(&admin.AdminConfig{DB: db})
	// Регистрируем ресурсы
	// Пользователи
	Admin.AddResource(&models.User{}, &admin.Config{
		Menu: []string{"Users menu"},
	})
	Admin.AddResource(&models.UserDetail{}, &admin.Config{
		Menu: []string{"Users menu"},
	}).Meta(&admin.Meta{
		Name:  "User",
		Label: "Username",
		Type:  "string",
		Valuer: func(record interface{}, context *qor.Context) (result interface{}) { // Используем qor.Context
			// Преобразуем record в *models.UserDetail
			if userDetail, ok := record.(*models.UserDetail); ok {
				// Загружаем пользователя по UserID
				if err := db.Where("id = ?", userDetail.UserID).First(&userDetail.User).Error; err != nil {
					fmt.Printf("User Load Error: %v\n", err)
					return "Not found"
				}

				// Проверяем, что поле Username не пустое
				if userDetail.User.Username != "" {
					return userDetail.User.Username
				}
			}
			return "Not found"
		},
	})
	Admin.AddResource(&models.UserPrivacy{}, &admin.Config{
		Menu: []string{"Users menu"},
	}).Meta(&admin.Meta{
		Name:  "User",
		Label: "Username",
		Type:  "string",
		Valuer: func(record interface{}, context *qor.Context) (result interface{}) {
			// Преобразуем record в *models.UserPrivacy
			if userPrivacy, ok := record.(*models.UserPrivacy); ok {
				// Загружаем пользователя по UserID
				if err := db.Where("id = ?", userPrivacy.UserID).First(&userPrivacy.User).Error; err != nil {
					fmt.Printf("User Load Error: %v\n", err)
					return "Not found"
				}

				// Проверяем, что поле Username не пустое
				if userPrivacy.User.Username != "" {
					return userPrivacy.User.Username
				}
			}
			return "Not found"
		},
	})
	Admin.AddResource(&models.UserEmailConfirmation{}, &admin.Config{
		Menu: []string{"Users menu"},
	}).Meta(&admin.Meta{
		Name:  "User",
		Label: "Username",
		Type:  "string",
		Valuer: func(record interface{}, context *qor.Context) (result interface{}) {
			// Преобразуем record в *models.UserEmailConfirmation
			if userEmailConfirmation, ok := record.(*models.UserEmailConfirmation); ok {
				// Загружаем пользователя по UserID
				if err := db.Where("id = ?", userEmailConfirmation.UserID).First(&userEmailConfirmation.User).Error; err != nil {
					fmt.Printf("User Load Error: %v\n", err)
					return "Not found"
				}

				// Проверяем, что поле Username не пустое
				if userEmailConfirmation.User.Username != "" {
					return userEmailConfirmation.User.Username
				}
			}
			return "Not found"
		},
	})
	Admin.AddResource(&models.UserImage{}, &admin.Config{
		Menu: []string{"Users menu"},
	}).Meta(&admin.Meta{
		Name:  "User",
		Label: "Username",
		Type:  "string",
		Valuer: func(record interface{}, context *qor.Context) (result interface{}) {
			// Преобразуем record в *models.UserImage
			if userImage, ok := record.(*models.UserImage); ok {
				// Загружаем пользователя по UserID
				if err := db.Where("id = ?", userImage.UserID).First(&userImage.User).Error; err != nil {
					fmt.Printf("User Load Error: %v\n", err)
					return "Not found"
				}

				// Проверяем, что поле Username не пустое
				if userImage.User.Username != "" {
					return userImage.User.Username
				}
			}
			return "Not found"
		},
	})
	Admin.AddResource(&models.Role{}, &admin.Config{
		Menu: []string{"Users menu"},
	})
	Admin.AddResource(&models.Session{}, &admin.Config{
		Menu: []string{"Users menu"},
	}).Meta(&admin.Meta{
		Name:  "User",
		Label: "Username",
		Type:  "string",
		Valuer: func(record interface{}, context *qor.Context) (result interface{}) {
			// Преобразуем record в *models.UserImage
			if userSession, ok := record.(*models.Session); ok {
				// Загружаем пользователя по UserID
				if err := db.Where("id = ?", userSession.UserID).First(&userSession.User).Error; err != nil {
					fmt.Printf("User Load Error: %v\n", err)
					return "Not found"
				}

				// Проверяем, что поле Username не пустое
				if userSession.User.Username != "" {
					return userSession.User.Username
				}
			}
			return "Not found"
		},
	})
	// Животные
	Admin.AddResource(&models.Animal{}, &admin.Config{
		Menu: []string{"Animals menu"},
	})
	Admin.AddResource(&models.AnimalAge{}, &admin.Config{
		Menu: []string{"Animals menu"},
	})
	Admin.AddResource(&models.AnimalStatus{}, &admin.Config{
		Menu: []string{"Animals menu"},
	})
	Admin.AddResource(&models.AnimalType{}, &admin.Config{
		Menu: []string{"Animals menu"},
	})
	Admin.AddResource(&models.Gender{}, &admin.Config{
		Menu: []string{"Animals menu"},
	})
	Admin.AddResource(&models.PostImage{}, &admin.Config{
		Menu: []string{"Animals menu"},
	})
	Admin.AddResource(&models.MedicalRecord{}, &admin.Config{
		Menu: []string{"Animals menu"},
	})

	// Усыновления
	Admin.AddResource(&models.Adoption{}, &admin.Config{
		Menu: []string{"Adoptions menu"},
	})
	Admin.AddResource(&models.AdoptionStatus{}, &admin.Config{
		Menu: []string{"Adoptions menu"},
	})

	// Форум
	Admin.AddResource(&models.Topic{}, &admin.Config{
		Menu: []string{"Forum and community menu"},
	})
	Admin.AddResource(&models.Like{}, &admin.Config{
		Menu: []string{"Forum and community menu"},
	})
	Admin.AddResource(&models.Post{}, &admin.Config{
		Menu: []string{"Forum and community menu"},
	})
	Admin.AddResource(&models.PostLike{}, &admin.Config{
		Menu: []string{"Forum and community menu"},
	})
	Admin.AddResource(&models.Feedback{}, &admin.Config{
		Menu: []string{"Forum and community menu"},
	})

	return Admin
}
