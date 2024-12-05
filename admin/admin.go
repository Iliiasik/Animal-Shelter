package admin

import (
	"Animals_Shelter/models"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/qor/admin"
	"github.com/qor/qor"
)

// InitAdmin инициализирует и возвращает QOR Admin
func InitAdmin(db *gorm.DB) *admin.Admin {
	// Создаём новый экземпляр QOR Admin
	Admin := admin.New(&admin.AdminConfig{DB: db})

	// Регистрируем ресурсы
	Admin.AddResource(&models.Adoption{})
	Admin.AddResource(&models.AdoptionStatus{})
	Admin.AddResource(&models.AnimalAge{})
	Admin.AddResource(&models.Animal{})
	Admin.AddResource(&models.AnimalStatus{})
	Admin.AddResource(&models.AnimalType{})
	Admin.AddResource(&models.Feedback{})
	Admin.AddResource(&models.Gender{})
	Admin.AddResource(&models.Like{})
	Admin.AddResource(&models.MedicalRecord{})
	Admin.AddResource(&models.PostImage{})
	Admin.AddResource(&models.Post{})
	Admin.AddResource(&models.Role{})
	Admin.AddResource(&models.Session{})
	Admin.AddResource(&models.Topic{})
	Admin.AddResource(&models.User{})
	Admin.AddResource(&models.UserDetail{}, &admin.Config{}).Meta(&admin.Meta{
		Name:  "User",
		Label: "Username",
		Type:  "string",
		Valuer: func(record interface{}, context *qor.Context) (result interface{}) {
			if userDetail, ok := record.(*models.UserDetail); ok {
				// Используем Where для загрузки данных конкретного пользователя
				if err := db.Where("id = ?", userDetail.UserID).First(&userDetail.User).Error; err != nil {
					fmt.Printf("Ошибка загрузки пользователя: %v\n", err)
				}
				fmt.Printf("User: %+v\n", userDetail.User)
				if userDetail.User.Username != "" {
					return userDetail.User.Username
				}
			}
			return "Not found"
		},
	})

	Admin.AddResource(&models.UserEmailConfirmation{})

	Admin.AddResource(&models.UserImage{}, &admin.Config{
		Menu: []string{"User Images"},
	}).Meta(&admin.Meta{
		Name:  "User",
		Label: "Username",
		Type:  "string",
		Valuer: func(record interface{}, ctx *qor.Context) interface{} {
			if userImage, ok := record.(*models.UserImage); ok {
				// Загружаем данные о пользователе
				if err := db.Model(userImage).Preload("User").First(userImage).Error; err != nil {
					fmt.Printf("Ошибка загрузки пользователя: %v\n", err)
				}
				fmt.Printf("UserImage: %+v\n", userImage.User) // Логируем содержимое
				if userImage.User.Username != "" {
					return userImage.User.Username
				}
			}
			return "Not found"
		},
	})

	return Admin
}
