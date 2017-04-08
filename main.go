package main

import (
	"html/template"
	"io/ioutil"
	"net/http"
	"errors"
	"regexp"
	"github.com/skip2/go-qrcode"
	"log"
	"fmt"
	"os"
)
var templates = template.Must(template.ParseFiles("temp/edit.html", "temp/view.html"))
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

func GenerateQRcode(phrase string) {
	
	err := qrcode.WriteFile(phrase, qrcode.Medium, 256, "public/img/qr-"+phrase+".png")
	
	if err!= nil {
		log.Fatal(err)
	}
}

func getTitle(w http.ResponseWriter, r *http.Request) (string, error) {
    m := validPath.FindStringSubmatch(r.URL.Path)
    if m == nil {
        http.NotFound(w, r)
        return "", errors.New("Invalid Page Title")
    }
    return m[2], nil // The title is the second subexpression.
}

type Page struct {
    Title string
    Body  []byte
}

func (p *Page) save() error {
    filename := p.Title + ".txt"
    return ioutil.WriteFile("data/"+filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
    filename := title + ".txt"
    body, err := ioutil.ReadFile("data/"+filename)
    if err != nil {
        return nil, err
    }
    return &Page{Title: title, Body: body}, nil
}
func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
    err := templates.ExecuteTemplate(w, tmpl+".html", p)
    if err != nil {
    	http.Error(w, err.Error(), http.StatusInternalServerError)
    }   
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {

    p, err := loadPage(title)
    if err != nil {
    	http.Redirect(w, r, "/edit/"+title, http.StatusFound)
    	return
    }
    renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
 
    p, err := loadPage(title)
    if err != nil {
        p = &Page{Title: title}
    }
    renderTemplate(w, "edit", p)
}
func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
    
    body := r.FormValue("body")
    p := &Page{Title: title, Body: []byte(body)}
    err := p.save()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    fmt.Printf(title)
    GenerateQRcode(title)
    fmt.Printf("done with:", title)
    http.Redirect(w, r, "/view/"+title, http.StatusFound)
}
func frontPageHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/view/FrontPage", http.StatusFound)
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        m := validPath.FindStringSubmatch(r.URL.Path)
        if m == nil {
            http.NotFound(w, r)
            return
        }
        fn(w, r, m[2])
    }
}


func main() {
	port := os.Getenv("PORT")

	if port == "" {
		port = ":8080"
	}
    http.HandleFunc("/view/", makeHandler(viewHandler))
    http.HandleFunc("/edit/", makeHandler(editHandler))
    http.HandleFunc("/save/", makeHandler(saveHandler))
    http.HandleFunc("/", frontPageHandler)
    http.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir("public"))))
    http.ListenAndServe(port, nil)
}