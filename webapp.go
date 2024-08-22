package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var templates = template.Must(template.ParseFiles("read.html", "readall.html", "update.html"))

var validPath = regexp.MustCompile("^/(read|update|commitupdate)/([a-zA-Z0-9]+)$")
var sampleToDos = MakeSampleToDos()

type ToDo struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Due         time.Time `json:"due"`
	Priority    int       `json:"priortiy"`
	Status      string    `json:"status"`
}

type Page struct {
	Title               string
	Body                []byte
	ExistingTitle       string
	ExistingDescription string
	ExistingDueYear     string
	ExistingDueMonth    string
	ExistingDueDay      string
	ExistingDueHour     string
	ExistingPriority    string
	ExistingStatus      string
}

func MakeSampleToDos() []ToDo {
	return []ToDo{
		{"Clean room", "Do laundry, hoover, clean sheets, change bin", time.Date(2024, 9, 15, 12, 0, 0, 0, time.Local), 2, "not started"},
		{"Clean car", "Remove mess, hoover, wash outside", time.Date(2024, 9, 12, 12, 0, 0, 0, time.Local), 4, "not started"},
		{"Buy Present Mum", "Buy a birthday present for mum, maybe a massage(?)", time.Date(2024, 9, 25, 12, 0, 0, 0, time.Local), 9, "not started"},
		{"ChaseBuilders", "Chase for quote builder@builder.com", time.Date(2024, 9, 11, 12, 0, 0, 0, time.Local), 3, "in progress"},
		{"Finish Go Academy", "Complete ToDoApp inc testing", time.Date(2024, 8, 30, 12, 30, 0, 0, time.Local), 10, "not started"},
		{"Learn to fly", "Jump off progressively higher things until I figure it out", time.Date(2024, 9, 28, 12, 0, 0, 0, time.Local), 3, "not started"},
		{"Swallow spider", "Just something big enough to catch that fly", time.Date(2024, 8, 29, 12, 0, 0, 0, time.Local), 4, "complete"},
		{"Nigerian Prince", "Nigerian prince is driving around with a truck full of gold, waiting for me to send money sam.cam@hotmail.com", time.Date(2024, 9, 1, 12, 0, 0, 0, time.Local), 7, "not started"},
		{"Count to a billion", "Start at 1, count to a billion, tell theboy next door that I counted to a bigger number than him", time.Date(2024, 9, 9, 12, 0, 0, 0, time.Local), 10, "not started"},
		{"Leave the house", "Probably a bit of a stretch goal but worth a try", time.Date(2050, 1, 1, 00, 0, 0, 0, time.Local), 1, "not started"},
	}
}

func Jsonify(TD ...ToDo) ([]byte, bool) {
	jsonData, err := json.Marshal(TD)
	if err != nil {
		fmt.Println("Error marshaling to JSON:", err)
		return nil, false
	}
	return jsonData, true
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
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

func readHandler(w http.ResponseWriter, r *http.Request, title string) {
	for _, td := range sampleToDos {
		if strings.EqualFold(strings.ReplaceAll(td.Title, " ", ""), title) {
			p := Page{Title: strings.ReplaceAll(td.Title, " ", ""), ExistingTitle: td.Title, ExistingDescription: td.Description, ExistingDueYear: strconv.Itoa(td.Due.Year()), ExistingDueMonth: strconv.Itoa(int(td.Due.Month())), ExistingDueDay: strconv.Itoa(td.Due.Day()), ExistingDueHour: strconv.Itoa(td.Due.Hour()), ExistingPriority: strconv.Itoa(td.Priority), ExistingStatus: td.Status}
			renderTemplate(w, "read", &p)
		}
	}
}

func updateHandler(w http.ResponseWriter, r *http.Request, title string) {
	for _, td := range sampleToDos {
		if strings.EqualFold(strings.ReplaceAll(td.Title, " ", ""), title) {
			p := Page{Title: strings.ReplaceAll(title, " ", ""), ExistingTitle: td.Title, ExistingDescription: td.Description, ExistingDueYear: strconv.Itoa(td.Due.Year()), ExistingDueMonth: strconv.Itoa(int(td.Due.Month())), ExistingDueDay: strconv.Itoa(td.Due.Day()), ExistingDueHour: strconv.Itoa(td.Due.Hour()), ExistingPriority: strconv.Itoa(td.Priority), ExistingStatus: td.Status}
			renderTemplate(w, "update", &p)
		}
	}
}

func updator(update ToDo, oldTitle string) {
	fmt.Println("oldTitle")
	fmt.Println(oldTitle)
	for i, td := range sampleToDos {
		if td.Title == oldTitle {
			fmt.Println(update)
			sampleToDos[i] = update
		}
	}
}

func commitUpdateHandler(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("title")
	description := r.FormValue("description")
	fmt.Println("description:")
	fmt.Println(description)
	dueYear, err := strconv.Atoi(r.FormValue("dueyear"))
	if err != nil {
		fmt.Println("error!")
	}
	dueMonth, err := strconv.Atoi(r.FormValue("duemonth"))
	if err != nil {
		fmt.Println("error!")
	}
	dueDay, err := strconv.Atoi(r.FormValue("dueday"))
	if err != nil {
		fmt.Println("error!")
	}
	dueHour, err := strconv.Atoi(r.FormValue("duehour"))
	if err != nil {
		fmt.Println("error!")
	}
	priority, err := strconv.Atoi(r.FormValue("priority"))
	if err != nil {
		fmt.Println("error!")
	}
	existingtitle := r.FormValue("existingtitle")
	status := r.FormValue("status")
	due := time.Date(dueYear, time.Month(dueMonth), dueDay, dueHour, 0, 0, 0, time.Local)
	updatedToDo := ToDo{title, description, due, priority, status}
	fmt.Println(updatedToDo)
	updator(updatedToDo, existingtitle)
	http.Redirect(w, r, "/read/"+strings.ReplaceAll(updatedToDo.Title, " ", ""), http.StatusFound)
}

func readAllHandler(w http.ResponseWriter, r *http.Request) {
	body := []byte{}
	for _, td := range sampleToDos {
		body = append(body, []byte(td.Title)...)
		body = append(body, []byte(" - ")...)
	}
	p := Page{Title: "All To-dos", Body: body}
	renderTemplate(w, "readall", &p)
}

func main() {
	// http.HandleFunc("/create/", makeHandler(createHandler))
	http.HandleFunc("/read/", makeHandler(readHandler))
	http.HandleFunc("/readall/", (readAllHandler))
	http.HandleFunc("/update/", makeHandler(updateHandler))
	// http.HandleFunc("/delete/", makeHandler(deleteHandler))
	http.HandleFunc("/commitupdate/", commitUpdateHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
