package render

import (
	"net/http"
	"testing"

	"github.com/ShamuhammetYlyas/bookings/internal/models"
)

func TestAddDefaultData(t *testing.T) {
	var td models.TemplateData
	req, err := getSession()
	if err != nil {
		t.Error(err)
	}
	session.Put(req.Context(), "flash", "123")
	result := AddDefaultData(&td, req)
	if result.Flash != "123" {
		t.Error("flash value of 123 not found in session")
	}
}

func TestRenderTemplate(t *testing.T) {
	// render.go-da pathToTemplates = "./templates"-di
	// yokarky dine render.go-da root folder-a cykyar yagny bookings folderin icine
	// yone biz testde doly yerini gorkezmeli bolyar yagny override edyaris CreateTemplateCache funksiya ishlemaka
	pathToTemplates = "./../../templates"
	tc, err := CreateTemplateCache()
	if err != nil {
		t.Error(err)
	}

	//aslynda shu yerde setupda doreden testApp-mizin TemplateCache-ni uuytgetdigimiz bolyar
	app.TemplateCache = tc

	req, err := getSession()
	if err != nil {
		t.Error(err)
	}

	var ww myWriter

	err = Template(&ww, req, "home.page.tmpl", &models.TemplateData{})
	if err != nil {
		t.Error(err)
	}

	err = Template(&ww, req, "non-existent.page.tmpl", &models.TemplateData{})
	if err == nil {
		t.Error("Rendered non existent template")
	}
}

func TestNewTemplate(t *testing.T) {
	NewRenderer(app)
}

func TestCreateTemplateCache(t *testing.T) {
	pathToTemplates = "./../../templates"
	_, err := CreateTemplateCache()
	if err != nil {
		t.Error(err)
	}
}

func getSession() (*http.Request, error) {
	req, err := http.NewRequest("GET", "/some-url", nil)
	if err != nil {
		return nil, err
	}

	ctx := req.Context()
	ctx, _ = session.Load(ctx, req.Header.Get("X-Session"))
	req = req.WithContext(ctx)

	return req, nil
}
