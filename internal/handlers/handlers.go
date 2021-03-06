package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/ShamuhammetYlyas/bookings/internal/config"
	"github.com/ShamuhammetYlyas/bookings/internal/driver"
	"github.com/ShamuhammetYlyas/bookings/internal/forms"
	"github.com/ShamuhammetYlyas/bookings/internal/helpers"
	"github.com/ShamuhammetYlyas/bookings/internal/models"
	"github.com/ShamuhammetYlyas/bookings/internal/render"
	"github.com/ShamuhammetYlyas/bookings/internal/repository"
	"github.com/ShamuhammetYlyas/bookings/internal/repository/dbrepo"
	"github.com/go-chi/chi/v5"
)

//Repo the repository used by the handlers
var Repo *Repository

// Repository is the repository type
// Repository pattern ulanmagymyzyn sebabi app configurationlaryn hem handlers package-de
// hemde render package-de ulanylyandygy ucin. Kop yerde app config gerek bolany ucin
// biz handler package ucin repository pattern ulandyk. Hokman shuny ulanmaly diyen zat yok
// App-i main.go-dan NewRepo funksiyasyna gelyan app-in adresine denledik
type Repository struct {
	App *config.AppConfig
	DB  repository.DatabaseRepo
}

// NewRepo creates a new repository
// main.go-da shu NewRepo, doredilen app configin adresini ugradypdyk.
// bu funksiya hem bir repository doredip shonun adresini return edyar.
// main.go-da hem shu doredilen repositoryn adresini alyp bir repo variable-a
// denledik yagny main.go-da repo=0xc213123123(doredilen repositoryn adresi)
func NewRepo(a *config.AppConfig, db *driver.DB) *Repository {
	return &Repository{
		App: a,
		DB:  dbrepo.NewPostgresRepo(db.SQL, a),
	}
}

func NewTestRepo(a *config.AppConfig) *Repository {
	return &Repository{
		App: a,
		DB:  dbrepo.NewTestingRepo(a),
	}
}

//NewHandlers sets the repository for the handlers
//NewHandlers funskiyasy main.go-da ulanylyar. Bu metod NewRepodan return edilen repositoryn adresine garashyar
//we sho adresi alyp yokarda doredilen Repo variable-a denleyar. yagny shuwagt Repo = 0xca1231239(doredilen reponyn adresi)
func NewHandlers(r *Repository) {
	Repo = r
}

//Home handleri indi repositoryn receiver funksiya boldy
// biz nirede Home handleri ulanjak bolsak Repo objectin usti bilen ulanmaly bolyarys.
func (m *Repository) Home(res http.ResponseWriter, req *http.Request) {
	// req.RemoteAddr request ugradyan clientin adresini beryar.
	// ony alyp remoteIP variable-a denleyaris
	// remoteIP := req.RemoteAddr

	// m, doredilen Repo.
	// App sho reponyn App propertisi.
	// Bu propertinin type-i hem *config.AppConfig bolany ucin onun icindaki propertyleri ulanyp bolyar property
	// yagny structyn icinde struct typeli property bar
	// Session *config.AppConfigin Session propertisi. Biz muny main.go-da beripdik ilki bashda
	// Put session managerin receiver funksiyasy
	// m.App.Session.Put(req.Context(), "remote_ip", remoteIP)

	render.Template(res, req, "home.page.tmpl", &models.TemplateData{})
}

func (m *Repository) About(res http.ResponseWriter, req *http.Request) {
	//some logic here
	render.Template(res, req, "about.page.tmpl", &models.TemplateData{})
}

func (m *Repository) Contact(res http.ResponseWriter, req *http.Request) {
	render.Template(res, req, "contact.page.tmpl", &models.TemplateData{})
}

func (m *Repository) Generals(res http.ResponseWriter, req *http.Request) {
	render.Template(res, req, "generals.page.tmpl", &models.TemplateData{})
}

func (m *Repository) Majors(res http.ResponseWriter, req *http.Request) {
	render.Template(res, req, "majors.page.tmpl", &models.TemplateData{})
}

func (m *Repository) Reservation(res http.ResponseWriter, req *http.Request) {
	reservation, ok := m.App.Session.Get(req.Context(), "reservation").(models.Reservation)
	if !ok {
		helpers.ServerError(res, errors.New("cannot get reservation from session"))
		return
	}

	room, err := m.DB.GetRoomByID(reservation.RoomID)
	if err != nil {
		helpers.ServerError(res, err)
		return
	}

	reservation.Room.RoomName = room.RoomName
	m.App.Session.Put(req.Context(), "reservation", reservation)

	sd := reservation.StartDate.Format("2006-01-02")
	ed := reservation.EndDate.Format("2006-01-02")

	stringMap := make(map[string]string)
	stringMap["start_date"] = sd
	stringMap["end_date"] = ed
	data := make(map[string]interface{})
	data["reservation"] = reservation
	render.Template(res, req, "make-reservation.page.tmpl", &models.TemplateData{
		// laravelidaki formdaky old value-lar ucin we formyn errorlaryny gorkezmek ucin
		// renderde-de form objecti gerek bolyar
		Form:      forms.New(nil),
		Data:      data,
		StringMap: stringMap,
	})
}

func (m *Repository) PostReservation(res http.ResponseWriter, req *http.Request) {
	reservation, ok := m.App.Session.Get(req.Context(), "reservation").(models.Reservation)
	if !ok {
		helpers.ServerError(res, errors.New("cannot get reservation from session"))
		return
	}
	err := req.ParseForm()
	if err != nil {
		helpers.ServerError(res, err)
		return
	}

	reservation.FirstName = req.Form.Get("first_name")
	reservation.LastName = req.Form.Get("last_name")
	reservation.Phone = req.Form.Get("phone")
	reservation.Email = req.Form.Get("email")

	//req.PostForm dine ParseForm metody cagyrylanyndan sonra ulayp bolyar
	//POST, PUT, PATCH metodlary bilen gelen formyn parsed edilen gornushini saklayar
	form := forms.New(req.PostForm)
	// form.Has("first_name", req)
	form.Required("first_name", "last_name", "email")
	form.MinLength("first_name", 3)
	// form.IsEmail("email")

	if !form.Valid() {
		data := make(map[string]interface{})
		data["reservation"] = reservation

		render.Template(res, req, "make-reservation.page.tmpl", &models.TemplateData{
			// laravelidaki formdaky old value-lar ucin we formyn errorlaryny gorkezmek ucin
			// renderde-de form objecti gerek bolyar
			Form: form,
			Data: data,
		})

		return
	}

	newReservationID, err := m.DB.InsertReservation(reservation)
	if err != nil {
		helpers.ServerError(res, err)
		return
	}

	m.App.Session.Put(req.Context(), "reservation", reservation)
	restriction := models.RoomRestriction{
		StartDate:     reservation.StartDate,
		EndDate:       reservation.EndDate,
		RoomID:        reservation.RoomID,
		ReservationID: newReservationID,
		RestrictionID: 1,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err = m.DB.InsertRoomRestriction(restriction)
	if err != nil {
		helpers.ServerError(res, err)
		return
	}

	m.App.Session.Put(req.Context(), "reservation", reservation)
	http.Redirect(res, req, "/reservation-summary", http.StatusSeeOther)
}

func (m *Repository) Availability(res http.ResponseWriter, req *http.Request) {
	render.Template(res, req, "search-availability.page.tmpl", &models.TemplateData{})
}

func (m *Repository) PostAvailability(res http.ResponseWriter, req *http.Request) {
	start := req.Form.Get("start")
	end := req.Form.Get("end")

	layout := "2006-01-02"
	startDate, err := time.Parse(layout, start)
	if err != nil {
		fmt.Println(err)
		// helpers.ServerError(res, err)
		return
	}

	endDate, err := time.Parse(layout, end)
	if err != nil {
		fmt.Println(err)
		// helpers.ServerError(res, err)
		return
	}

	rooms, err := m.DB.SearchAvailabilityForAllRooms(startDate, endDate)
	if err != nil {
		fmt.Println(err)
		// helpers.ServerError(res, err)
		return
	}

	if len(rooms) == 0 {
		m.App.Session.Put(req.Context(), "error", "No availability")
		http.Redirect(res, req, "/search-availability", http.StatusSeeOther)
		return
	}

	data := make(map[string]interface{})
	data["rooms"] = rooms

	reservation := models.Reservation{
		StartDate: startDate,
		EndDate:   endDate,
	}

	m.App.Session.Put(req.Context(), "reservation", reservation)

	render.Template(res, req, "choose-room.page.tmpl", &models.TemplateData{
		Data: data,
	})
}

type jsonResponse struct {
	OK        bool   `json:"ok"`
	Message   string `json:"message"`
	RoomID    string `json:"room_id"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

func (m *Repository) AvailabilityJSON(res http.ResponseWriter, req *http.Request) {
	sd := req.Form.Get("start")
	ed := req.Form.Get("end")

	layout := "2006-01-02"
	startDate, err := time.Parse(layout, sd)
	if err != nil {
		helpers.ServerError(res, err)
		return
	}
	endDate, err := time.Parse(layout, ed)
	if err != nil {
		helpers.ServerError(res, err)
		return
	}

	roomID, err := strconv.Atoi(req.Form.Get("room_id"))
	if err != nil {
		helpers.ServerError(res, err)
		return
	}

	available, err := m.DB.SearchAvailabilityByDatesByRoomID(startDate, endDate, roomID)
	if err != nil {
		helpers.ServerError(res, err)
		return
	}
	resp := jsonResponse{
		OK:        available,
		Message:   "",
		RoomID:    strconv.Itoa(roomID),
		StartDate: sd,
		EndDate:   ed,
	}

	out, err := json.MarshalIndent(resp, "", "     ")
	if err != nil {
		helpers.ServerError(res, err)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	// fmt.Fprint(res, out)
	res.Write(out)

}

// displays the reservation summary
func (m *Repository) ReservationSummary(res http.ResponseWriter, req *http.Request) {
	// m.App.Session.Get(req.Context(), "reservation") yazanymyz bilen sessiondan maglumat alyp bilemzok
	// sessiondan maglumat almak ucin type-ni bildirmeli. bu yerde type assertion ulanyldy
	reservation, ok := m.App.Session.Get(req.Context(), "reservation").(models.Reservation)
	if !ok {
		// log.Println("Cannot get item from session")
		m.App.ErrorLog.Println("Cannot get item from session")
		m.App.Session.Put(req.Context(), "error", "Cannot get reservation from session")
		http.Redirect(res, req, "/", http.StatusTemporaryRedirect)
		return
	}

	//sessiondaky reservationy ayyryar
	m.App.Session.Remove(req.Context(), "reservation")

	data := make(map[string]interface{})
	data["reservation"] = reservation

	sd := reservation.StartDate.Format("2006-01-02")
	ed := reservation.EndDate.Format("2006-01-02")
	stringMap := make(map[string]string)
	stringMap["start_date"] = sd
	stringMap["end_date"] = ed

	render.Template(res, req, "reservation-summary.page.tmpl", &models.TemplateData{
		Data:      data,
		StringMap: stringMap,
	})
}

// Displays available rooms
func (m *Repository) ChooseRoom(res http.ResponseWriter, req *http.Request) {
	roomID, err := strconv.Atoi(chi.URLParam(req, "id"))
	if err != nil {
		helpers.ServerError(res, err)
		return
	}

	m.App.Session.Get(req.Context(), "reservation")

	reservation, ok := m.App.Session.Get(req.Context(), "reservation").(models.Reservation)
	if !ok {
		helpers.ServerError(res, err)
		return
	}

	reservation.RoomID = roomID
	m.App.Session.Put(req.Context(), "reservation", reservation)

	http.Redirect(res, req, "/make-reservation", http.StatusSeeOther)

}

func (m *Repository) BookRoom(res http.ResponseWriter, req *http.Request) {
	//id, s, e
	roomID, _ := strconv.Atoi(req.URL.Query().Get("id"))
	sd := req.URL.Query().Get("s")
	ed := req.URL.Query().Get("e")

	layout := "2006-01-02"
	startDate, _ := time.Parse(layout, sd)
	endDate, _ := time.Parse(layout, ed)

	var reservation models.Reservation

	room, err := m.DB.GetRoomByID(roomID)
	if err != nil {
		helpers.ServerError(res, err)
		return
	}

	reservation.Room.RoomName = room.RoomName
	reservation.RoomID = roomID
	reservation.StartDate = startDate
	reservation.EndDate = endDate

	m.App.Session.Put(req.Context(), "reservation", reservation)

	http.Redirect(res, req, "/make-reservation", http.StatusSeeOther)
}
