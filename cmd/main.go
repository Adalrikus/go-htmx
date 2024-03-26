package main

import (
	"html/template"
	"io"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type (
	Templates struct {
		templates *template.Template
	}

	Contact struct {
		Name  string
		Email string
		ID    int
	}

	Contacts []Contact

	Count struct {
		Count int
	}

	Data struct {
		Contacts Contacts
	}

	FormData struct {
		Values map[string]string
		Errors map[string]string
	}

	Page struct {
		Data Data
		Form FormData
	}
)

var id = 0

func newTemplates() *Templates {
	return &Templates{
		templates: template.Must(template.ParseGlob("web/templates/*.html")),
	}
}

func (t *Templates) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func newContact(name string, email string) Contact {
	id++
	return Contact{
		Name:  name,
		Email: email,
		ID:    id,
	}
}

func newData(contacts Contacts) Data {
	return Data{
		Contacts: contacts,
	}
}

func (d *Data) hasEmail(email string) bool {
	for _, contact := range d.Contacts {
		if contact.Email == email {
			return true
		}
	}
	return false
}

func (d *Data) indexOfContact(id int) int {
	for i, contact := range d.Contacts {
		if contact.ID == id {
			return i
		}
	}
	return -1
}

func newFormData() FormData {
	return FormData{
		Values: make(map[string]string),
		Errors: make(map[string]string),
	}
}

func newPage(data Data, form FormData) Page {
	return Page{
		Data: data,
		Form: form,
	}
}

func main() {
	e := echo.New()
	e.Use(middleware.Logger())

	data := newData(Contacts{
		newContact("Adil", "a@a.com"),
		newContact("Khan", "k@k.com"),
	})
	page := newPage(data, newFormData())
	e.Renderer = newTemplates()

	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "index", page)
	})

	e.POST("/contacts", func(c echo.Context) error {
		name := c.FormValue("name")
		email := c.FormValue("email")

		if page.Data.hasEmail(email) {
			page.Form.Values["name"] = name
			page.Form.Values["email"] = email

			page.Form.Errors["email"] = "Email already exists"

			return c.Render(http.StatusUnprocessableEntity, "form", page.Form)
		}

		contact := newContact(name, email)
		page.Data.Contacts = append(page.Data.Contacts, contact)

		c.Render(http.StatusOK, "form", page.Form)
		return c.Render(http.StatusOK, "oob-contact", contact)
	})

	e.DELETE("/contacts/:id", func(c echo.Context) error {
		idString := c.Param("id")
		id, err := strconv.Atoi(idString)
		if err != nil {
			return c.String(http.StatusBadRequest, "Invalid ID")
		}

		index := page.Data.indexOfContact(id)
		if index == -1 {
			return c.String(http.StatusNotFound, "Contact not found")
		}

		page.Data.Contacts = append(page.Data.Contacts[:index], page.Data.Contacts[index+1:]...)

		return c.NoContent(http.StatusOK)
	})

	e.Logger.Fatal(e.Start(":8080"))
}
