package admin

import (
	"Animals_Shelter/models"
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/qor/admin"
	"github.com/qor/qor"
	"github.com/qor/qor/resource"
	"log"
	"os"
)

// InitAdmin инициализирует и возвращает QOR Admin
func InitAdmin(db *gorm.DB) *admin.Admin {
	db.SetLogger(log.New(os.Stdout, "\r\n", log.LstdFlags))

	// Настроим уровень логирования для GORM v1 (SQL-запросы)
	db.LogMode(true)

	// Создаём новый экземпляр QOR Admin
	Admin := admin.New(&admin.AdminConfig{DB: db, SiteName: "Animal Shelter"})

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
		Type:  "string", // Тип поля — строка, чтобы отобразить Username
		Valuer: func(record interface{}, context *qor.Context) (result interface{}) {
			if userDetail, ok := record.(*models.UserDetail); ok {
				// Загружаем связанные данные пользователя
				if err := db.Where("id = ?", userDetail.UserID).First(&userDetail.User).Error; err != nil {
					fmt.Printf("Error loading user: %v\n", err)
				}
				if userDetail.User.Username != "" {
					return userDetail.User.Username // Возвращаем Username
				}
			}
			return ""
		},
		Setter: func(record interface{}, metaValue *resource.MetaValue, context *qor.Context) {
			if userDetail, ok := record.(*models.UserDetail); ok {
				// Преобразуем значение в срез строк
				if usernames, ok := metaValue.Value.([]string); ok && len(usernames) > 0 {
					username := usernames[0] // Берём первое значение из среза
					var user models.User
					if err := db.Where("username = ?", username).First(&user).Error; err != nil {
						if errors.Is(err, gorm.ErrRecordNotFound) {
							// Если пользователь не найден, добавляем ошибку в контекст
							context.AddError(fmt.Errorf("User with username '%s' not found", username))
						} else {
							// Обработка других ошибок
							fmt.Printf("Error searching user: %v\n", err)
						}
						return
					}
					// Устанавливаем UserID в записи
					userDetail.UserID = user.ID
				} else {
					fmt.Println("Setter: value is empty or incorrect")
				}
			}
		},
	})

	Admin.AddResource(&models.UserPrivacy{}, &admin.Config{
		Menu: []string{"Users menu"},
	}).Meta(&admin.Meta{
		Name:  "User",
		Label: "Username",
		Type:  "string",
		Valuer: func(record interface{}, context *qor.Context) (result interface{}) {
			if userPrivacy, ok := record.(*models.UserPrivacy); ok {
				// Загружаем связанные данные пользователя
				if err := db.Where("id = ?", userPrivacy.UserID).First(&userPrivacy.User).Error; err != nil {
					fmt.Printf("Error load user: %v\n", err)
				}
				if userPrivacy.User.Username != "" {
					return userPrivacy.User.Username // Возвращаем Username
				}
			}
			return ""
		},
		Setter: func(record interface{}, metaValue *resource.MetaValue, context *qor.Context) {
			if userPrivacy, ok := record.(*models.UserPrivacy); ok {
				// Преобразуем значение в срез строк
				if usernames, ok := metaValue.Value.([]string); ok && len(usernames) > 0 {
					username := usernames[0] // Берём первое значение из среза
					var user models.User
					if err := db.Where("username = ?", username).First(&user).Error; err != nil {
						if errors.Is(err, gorm.ErrRecordNotFound) {
							// Если пользователь не найден, добавляем ошибку в контекст
							context.AddError(fmt.Errorf("User with username '%s' not found", username))
						} else {
							// Обработка других ошибок
							fmt.Printf("Error searching user: %v\n", err)
						}
						return
					}
					// Устанавливаем UserID в записи
					userPrivacy.UserID = user.ID
				} else {
					fmt.Println("Setter: value is empty or incorrect")
				}
			}
		},
	})
	Admin.AddResource(&models.UserEmailConfirmation{}, &admin.Config{
		Menu: []string{"Users menu"},
	}).Meta(&admin.Meta{
		Name:  "User",
		Label: "Username",
		Type:  "string",
		Valuer: func(record interface{}, context *qor.Context) (result interface{}) {
			if userEmailConfirmation, ok := record.(*models.UserEmailConfirmation); ok {
				// Загружаем связанные данные пользователя
				if err := db.Where("id = ?", userEmailConfirmation.UserID).First(&userEmailConfirmation.User).Error; err != nil {
					fmt.Printf("Error load user: %v\n", err)
				}
				if userEmailConfirmation.User.Username != "" {
					return userEmailConfirmation.User.Username // Возвращаем Username
				}
			}
			return ""
		},
		Setter: func(record interface{}, metaValue *resource.MetaValue, context *qor.Context) {
			if userEmailConfirmation, ok := record.(*models.UserEmailConfirmation); ok {
				// Преобразуем значение в срез строк
				if usernames, ok := metaValue.Value.([]string); ok && len(usernames) > 0 {
					username := usernames[0] // Берём первое значение из среза
					var user models.User
					if err := db.Where("username = ?", username).First(&user).Error; err != nil {
						if errors.Is(err, gorm.ErrRecordNotFound) {
							// Если пользователь не найден, добавляем ошибку в контекст
							context.AddError(fmt.Errorf("User with username '%s' not found", username))
						} else {
							// Обработка других ошибок
							fmt.Printf("Error searching user: %v\n", err)
						}
						return
					}
					// Устанавливаем UserID в записи
					userEmailConfirmation.UserID = user.ID
				} else {
					fmt.Println("Setter: value is empty or incorrect")
				}
			}
		},
	})
	Admin.AddResource(&models.UserImage{}, &admin.Config{
		Menu: []string{"Users menu"},
	}).Meta(&admin.Meta{
		Name:  "User",
		Label: "Username",
		Type:  "string",
		Valuer: func(record interface{}, context *qor.Context) (result interface{}) {
			if userImage, ok := record.(*models.UserImage); ok {
				// Загружаем связанные данные пользователя
				if err := db.Where("id = ?", userImage.UserID).First(&userImage.User).Error; err != nil {
					fmt.Printf("Error load user: %v\n", err)
				}
				if userImage.User.Username != "" {
					return userImage.User.Username // Возвращаем Username
				}
			}
			return ""
		},
		Setter: func(record interface{}, metaValue *resource.MetaValue, context *qor.Context) {
			if userImage, ok := record.(*models.UserImage); ok {
				// Преобразуем значение в срез строк
				if usernames, ok := metaValue.Value.([]string); ok && len(usernames) > 0 {
					username := usernames[0] // Берём первое значение из среза
					var user models.User
					if err := db.Where("username = ?", username).First(&user).Error; err != nil {
						if errors.Is(err, gorm.ErrRecordNotFound) {
							// Если пользователь не найден, добавляем ошибку в контекст
							context.AddError(fmt.Errorf("User with username '%s' not found", username))
						} else {
							// Обработка других ошибок
							fmt.Printf("Error searching user: %v\n", err)
						}
						return
					}
					// Устанавливаем UserID в записи
					userImage.UserID = user.ID
				} else {
					fmt.Println("Setter: value is empty or incorrect")
				}
			}
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
			if userSession, ok := record.(*models.Session); ok {
				// Загружаем связанные данные пользователя
				if err := db.Where("id = ?", userSession.UserID).First(&userSession.User).Error; err != nil {
					fmt.Printf("Error load user: %v\n", err)
				}
				if userSession.User.Username != "" {
					return userSession.User.Username // Возвращаем Username
				}
			}
			return ""
		},
		Setter: func(record interface{}, metaValue *resource.MetaValue, context *qor.Context) {
			if userSession, ok := record.(*models.Session); ok {
				// Преобразуем значение в срез строк
				if usernames, ok := metaValue.Value.([]string); ok && len(usernames) > 0 {
					username := usernames[0] // Берём первое значение из среза
					var user models.User
					if err := db.Where("username = ?", username).First(&user).Error; err != nil {
						if errors.Is(err, gorm.ErrRecordNotFound) {
							// Если пользователь не найден, добавляем ошибку в контекст
							context.AddError(fmt.Errorf("User with username '%s' not found", username))
						} else {
							// Обработка других ошибок
							fmt.Printf("Error searching user: %v\n", err)
						}
						return
					}
					// Устанавливаем UserID в записи
					userSession.UserID = user.ID
				} else {
					fmt.Println("Setter: value is empty or incorrect")
				}
			}
		},
	})
	// Животные
	Admin.AddResource(&models.Animal{}, &admin.Config{
		Menu: []string{"Animals menu"},
	}).Meta(&admin.Meta{
		Name:  "User",
		Label: "Username",
		Type:  "string", // Тип поля — строка, чтобы отобразить Username
		Valuer: func(record interface{}, context *qor.Context) (result interface{}) {
			if animalUser, ok := record.(*models.Animal); ok {
				// Загружаем связанные данные пользователя
				if err := db.Where("id = ?", animalUser.UserID).First(&animalUser.User).Error; err != nil {
					fmt.Printf("Error loading user: %v\n", err)
				}
				if animalUser.User.Username != "" {
					return animalUser.User.Username // Возвращаем Username
				}
			}
			return ""
		},
		Setter: func(record interface{}, metaValue *resource.MetaValue, context *qor.Context) {
			if animalUser, ok := record.(*models.Animal); ok {
				// Преобразуем значение в срез строк
				if usernames, ok := metaValue.Value.([]string); ok && len(usernames) > 0 {
					username := usernames[0] // Берём первое значение из среза
					var user models.User
					if err := db.Where("username = ?", username).First(&user).Error; err != nil {
						if errors.Is(err, gorm.ErrRecordNotFound) {
							// Если пользователь не найден, добавляем ошибку в контекст
							context.AddError(fmt.Errorf("User with username '%s' not found", username))
						} else {
							// Обработка других ошибок
							fmt.Printf("Error searching user: %v\n", err)
						}
						return
					}
					// Устанавливаем UserID в записи
					animalUser.UserID = user.ID
				} else {
					fmt.Println("Setter: value is empty or incorrect")
				}
			}
		},
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
	}).Meta(&admin.Meta{
		Name:  "User",
		Label: "Username",
		Type:  "string",
		Valuer: func(record interface{}, context *qor.Context) (result interface{}) {
			if adoption, ok := record.(*models.Adoption); ok {
				// Загружаем связанные данные пользователя
				if err := db.Where("id = ?", adoption.UserID).First(&adoption.User).Error; err != nil {
					fmt.Printf("Error load user: %v\n", err)
				}
				if adoption.User.Username != "" {
					return adoption.User.Username // Возвращаем Username
				}
			}
			return ""
		},
		Setter: func(record interface{}, metaValue *resource.MetaValue, context *qor.Context) {
			if adoption, ok := record.(*models.Adoption); ok {
				// Преобразуем значение в срез строк
				if usernames, ok := metaValue.Value.([]string); ok && len(usernames) > 0 {
					username := usernames[0] // Берём первое значение из среза
					var user models.User
					if err := db.Where("username = ?", username).First(&user).Error; err != nil {
						if errors.Is(err, gorm.ErrRecordNotFound) {
							// Если пользователь не найден, добавляем ошибку в контекст
							context.AddError(fmt.Errorf("User with username '%s' not found", username))
						} else {
							// Обработка других ошибок
							fmt.Printf("Error searching user: %v\n", err)
						}
						return
					}
					// Устанавливаем UserID в записи
					adoption.UserID = user.ID
				} else {
					fmt.Println("Setter: value is empty or incorrect")
				}
			}
		},
	})
	Admin.AddResource(&models.AdoptionStatus{}, &admin.Config{
		Menu: []string{"Adoptions menu"},
	})
	Admin.AddResource(&models.AdoptionStatistic{}, &admin.Config{
		Menu: []string{"Adoptions menu"},
	})

	// Форум
	Admin.AddResource(&models.Topic{}, &admin.Config{
		Menu: []string{"Forum and community menu"},
	}).Meta(&admin.Meta{
		Name:  "User",
		Label: "Username",
		Type:  "string",
		Valuer: func(record interface{}, context *qor.Context) (result interface{}) {
			if topic, ok := record.(*models.Topic); ok {
				// Загружаем связанные данные пользователя
				if err := db.Where("id = ?", topic.UserID).First(&topic.User).Error; err != nil {
					fmt.Printf("Error load user: %v\n", err)
				}
				if topic.User.Username != "" {
					return topic.User.Username // Возвращаем Username
				}
			}
			return ""
		},
		Setter: func(record interface{}, metaValue *resource.MetaValue, context *qor.Context) {
			if topic, ok := record.(*models.Topic); ok {
				// Преобразуем значение в срез строк
				if usernames, ok := metaValue.Value.([]string); ok && len(usernames) > 0 {
					username := usernames[0] // Берём первое значение из среза
					var user models.User
					if err := db.Where("username = ?", username).First(&user).Error; err != nil {
						if errors.Is(err, gorm.ErrRecordNotFound) {
							// Если пользователь не найден, добавляем ошибку в контекст
							context.AddError(fmt.Errorf("User with username '%s' not found", username))
						} else {
							// Обработка других ошибок
							fmt.Printf("Error searching user: %v\n", err)
						}
						return
					}
					// Устанавливаем UserID в записи
					topic.UserID = user.ID
				} else {
					fmt.Println("Setter: value is empty or incorrect")
				}
			}
		},
	})
	Admin.AddResource(&models.Like{}, &admin.Config{
		Menu: []string{"Forum and community menu"},
	}).Meta(&admin.Meta{
		Name:  "User",
		Label: "Username",
		Type:  "string",
		Valuer: func(record interface{}, context *qor.Context) (result interface{}) {
			if like, ok := record.(*models.Like); ok {
				// Загружаем связанные данные пользователя
				if err := db.Where("id = ?", like.UserID).First(&like.User).Error; err != nil {
					fmt.Printf("Error load user: %v\n", err)
				}
				if like.User.Username != "" {
					return like.User.Username // Возвращаем Username
				}
			}
			return ""
		},
		Setter: func(record interface{}, metaValue *resource.MetaValue, context *qor.Context) {
			if like, ok := record.(*models.Like); ok {
				// Преобразуем значение в срез строк
				if usernames, ok := metaValue.Value.([]string); ok && len(usernames) > 0 {
					username := usernames[0] // Берём первое значение из среза
					var user models.User
					if err := db.Where("username = ?", username).First(&user).Error; err != nil {
						if errors.Is(err, gorm.ErrRecordNotFound) {
							// Если пользователь не найден, добавляем ошибку в контекст
							context.AddError(fmt.Errorf("User with username '%s' not found", username))
						} else {
							// Обработка других ошибок
							fmt.Printf("Error searching user: %v\n", err)
						}
						return
					}
					// Устанавливаем UserID в записи
					like.UserID = user.ID
				} else {
					fmt.Println("Setter: value is empty or incorrect")
				}
			}
		},
	})
	Admin.AddResource(&models.Post{}, &admin.Config{
		Menu: []string{"Forum and community menu"},
	}).Meta(&admin.Meta{
		Name:  "User",
		Label: "Username",
		Type:  "string",
		Valuer: func(record interface{}, context *qor.Context) (result interface{}) {
			if post, ok := record.(*models.Post); ok {
				// Загружаем связанные данные пользователя
				if err := db.Where("id = ?", post.UserID).First(&post.User).Error; err != nil {
					fmt.Printf("Error load user: %v\n", err)
				}
				if post.User.Username != "" {
					return post.User.Username // Возвращаем Username
				}
			}
			return ""
		},
		Setter: func(record interface{}, metaValue *resource.MetaValue, context *qor.Context) {
			if post, ok := record.(*models.Post); ok {
				// Преобразуем значение в срез строк
				if usernames, ok := metaValue.Value.([]string); ok && len(usernames) > 0 {
					username := usernames[0] // Берём первое значение из среза
					var user models.User
					if err := db.Where("username = ?", username).First(&user).Error; err != nil {
						if errors.Is(err, gorm.ErrRecordNotFound) {
							// Если пользователь не найден, добавляем ошибку в контекст
							context.AddError(fmt.Errorf("User with username '%s' not found", username))
						} else {
							// Обработка других ошибок
							fmt.Printf("Error searching user: %v\n", err)
						}
						return
					}
					// Устанавливаем UserID в записи
					post.UserID = user.ID
				} else {
					fmt.Println("Setter: value is empty or incorrect")
				}
			}
		},
	})
	Admin.AddResource(&models.PostLike{}, &admin.Config{
		Menu: []string{"Forum and community menu"},
	}).Meta(&admin.Meta{
		Name:  "User",
		Label: "Username",
		Type:  "string",
		Valuer: func(record interface{}, context *qor.Context) (result interface{}) {
			if postLike, ok := record.(*models.PostLike); ok {
				// Загружаем связанные данные пользователя
				if err := db.Where("id = ?", postLike.UserID).First(&postLike.User).Error; err != nil {
					fmt.Printf("Error load user: %v\n", err)
				}
				if postLike.User.Username != "" {
					return postLike.User.Username // Возвращаем Username
				}
			}
			return ""
		},
		Setter: func(record interface{}, metaValue *resource.MetaValue, context *qor.Context) {
			if postLike, ok := record.(*models.PostLike); ok {
				// Преобразуем значение в срез строк
				if usernames, ok := metaValue.Value.([]string); ok && len(usernames) > 0 {
					username := usernames[0] // Берём первое значение из среза
					var user models.User
					if err := db.Where("username = ?", username).First(&user).Error; err != nil {
						if errors.Is(err, gorm.ErrRecordNotFound) {
							// Если пользователь не найден, добавляем ошибку в контекст
							context.AddError(fmt.Errorf("User with username '%s' not found", username))
						} else {
							// Обработка других ошибок
							fmt.Printf("Error searching user: %v\n", err)
						}
						return
					}
					// Устанавливаем UserID в записи
					postLike.UserID = user.ID
				} else {
					fmt.Println("Setter: value is empty or incorrect")
				}
			}
		},
	})
	Admin.AddResource(&models.Feedback{}, &admin.Config{
		Menu: []string{"Forum and community menu"},
	}).Meta(&admin.Meta{
		Name:  "User",
		Label: "Username",
		Type:  "string",
		Valuer: func(record interface{}, context *qor.Context) (result interface{}) {
			if feedback, ok := record.(*models.Feedback); ok {
				// Загружаем связанные данные пользователя
				if err := db.Where("id = ?", feedback.UserID).First(&feedback.User).Error; err != nil {
					fmt.Printf("Error load user: %v\n", err)
				}
				if feedback.User.Username != "" {
					return feedback.User.Username // Возвращаем Username
				}
			}
			return ""
		},
		Setter: func(record interface{}, metaValue *resource.MetaValue, context *qor.Context) {
			if feedback, ok := record.(*models.Feedback); ok {
				// Преобразуем значение в срез строк
				if usernames, ok := metaValue.Value.([]string); ok && len(usernames) > 0 {
					username := usernames[0] // Берём первое значение из среза
					var user models.User
					if err := db.Where("username = ?", username).First(&user).Error; err != nil {
						if errors.Is(err, gorm.ErrRecordNotFound) {
							// Если пользователь не найден, добавляем ошибку в контекст
							context.AddError(fmt.Errorf("User with username '%s' not found", username))
						} else {
							// Обработка других ошибок
							fmt.Printf("Error searching user: %v\n", err)
						}
						return
					}
					// Устанавливаем UserID в записи
					feedback.UserID = user.ID
				} else {
					fmt.Println("Setter: value is empty or incorrect")
				}
			}
		},
	})

	return Admin
}
