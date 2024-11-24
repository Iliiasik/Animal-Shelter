package admin

import (
	"Animals_Shelter/models"
	"github.com/jinzhu/gorm"
	"github.com/qor/admin"
	"github.com/qor5/admin/v3/presets"
)

// InitAdmin инициализирует и возвращает QOR Admin
func InitAdmin(db *gorm.DB) *admin.Admin {
	// Создаём новый экземпляр QOR Admin
	Admin := admin.New(&admin.AdminConfig{DB: db})
	pb := presets.New()
	// Регистрируем ресурсы
	Admin.AddResource(&models.AnimalStatus{})
	Admin.AddResource(&models.AnimalType{})
	Admin.AddResource(&models.Animal{})
	Admin.AddResource(&models.MedicalRecord{})
	Admin.AddResource(&models.PostImage{})
	Admin.AddResource(&models.Session{})
	Admin.AddResource(&models.Adoption{})
	Admin.AddResource(&models.Topic{})
	Admin.AddResource(&models.Post{})
	Admin.AddResource(&models.Like{})
	Admin.AddResource(&models.User{})

	// Добавляем кастомный дашборд в меню
	Admin.AddMenu(&admin.Menu{
		Name: "Dashboard",
		Link: "/admin/dashboard",
	})

	// Обработка маршрута для дашборда
	Admin.GetRouter().Get("/dashboard", func(ctx *admin.Context) {
		// Пример кастомного HTML контента
		ctx.Writer.Write([]byte(`
			<h1>Дашборд</h1>
			<p>Добро пожаловать в ваш кастомный дашборд!</p>
			<div>
				<a href="/admin/resources/animals">Список животных</a>
				<a href="/admin/resources/users">Пользователи</a>
			</div>
		`))
	})
	ConfigTopicDashboard(pb, db)

	return Admin
}
