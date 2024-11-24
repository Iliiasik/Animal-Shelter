package admin

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"Animals_Shelter/models"
	"github.com/jinzhu/gorm"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/ui/vuetify"
	h "github.com/theplant/htmlgo"
	"github.com/wcharczuk/go-chart/v2"
	"github.com/wcharczuk/go-chart/v2/drawing"
)

type TopicDashboard struct{}

// ConfigTopicDashboard настраивает дэшборд для топиков
func ConfigTopicDashboard(pb *presets.Builder, db *gorm.DB) {
	b := pb.Model(&TopicDashboard{}).Label("Topic Dashboard").URIName("topic-dashboard")

	lb := b.Listing()

	lb.PageFunc(func(ctx *web.EventContext) (r web.PageResponse, err error) {
		var topics []*models.Topic
		var userTopicsCount map[uint]int

		// Инициализация карты для подсчета
		userTopicsCount = make(map[uint]int)

		// Получение данных из базы
		if err = db.Model(&models.Topic{}).Find(&topics).Error; err != nil {
			r.Body = errorBody(err.Error())
			return
		}

		// Подсчет количества топиков по каждому пользователю
		for _, topic := range topics {
			userTopicsCount[uint(topic.UserID)]++
		}

		// Построение диаграммы
		var pie chart.PieChart
		pieBuffer := bytes.NewBuffer([]byte{})
		for userID, count := range userTopicsCount {
			pie.Values = append(pie.Values, chart.Value{
				Value: float64(count),
				Label: fmt.Sprintf("User ID %d", userID),
				Style: chart.Style{FillColor: drawing.ColorFromHex("#" + strconv.Itoa(0xFFFFFF / (len(userTopicsCount) + 1))[:6])},
			})
		}
		err = pie.Render(chart.SVG, pieBuffer)

		body := vuetify.VContainer(
			vuetify.VRow(
				h.Div(
					h.Div(h.Strong("User Topics Overview")).Class("mt-2 col col-12"),
					h.Div(
						h.RawHTML(
							strings.Replace(pieBuffer.String(), `width="1024" height="1024"`, `width="100%" height="100%" viewBox="-85 -80 1200 1200"`, -1)),
					).Class("v-card v-sheet theme--light").Style("height: 300px;"),
				).Class("col col-12"),
			),
		)

		r.Body = body
		r.PageTitle = "Topic Dashboard"
		return
	})
}

func errorBody(msg string) h.HTMLComponent {
	return vuetify.VContainer(
		h.P().Text(msg),
	)
}
