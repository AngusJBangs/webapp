package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var templates = template.Must(template.ParseFiles("read.html", "readall.html", "update.html", "create.html"))
var validPath = regexp.MustCompile("^/(read|update|commitupdate|create|commitcreate|delete)/([a-zA-Z0-9]+)$")

type ToDo struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Due         time.Time `json:"due"`
	Priority    int       `json:"priority"`
	Status      string    `json:"status"`
}
type ToDoRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Due         string `json:"due"`
	Priority    int    `json:"priority"`
	Status      string `json:"status"`
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

func Jsonify(TD ToDoRequest) ([]byte, error) {
	jsonData, err := json.Marshal(TD)
	if err != nil {
		fmt.Println("Error marshaling to JSON:", err)
		return nil, err
	}
	return jsonData, nil
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
	resp, err := http.Get("http://localhost:8080/todo/" + title)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	td, err := ParseBodyToToDo(body)
	if err != nil {
		log.Fatal(err)
	}

	p := Page{Title: strings.ReplaceAll(td.Title, " ", ""), ExistingTitle: td.Title, ExistingDescription: td.Description, ExistingDueYear: strconv.Itoa(td.Due.Year()), ExistingDueMonth: strconv.Itoa(int(td.Due.Month())), ExistingDueDay: strconv.Itoa(td.Due.Day()), ExistingDueHour: strconv.Itoa(td.Due.Hour()), ExistingPriority: strconv.Itoa(td.Priority), ExistingStatus: td.Status}
	renderTemplate(w, "read", &p)
}

func createHandler(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	year := strconv.Itoa(now.Year())
	month := strconv.Itoa(int(now.Month()))
	day := strconv.Itoa(now.Day())
	p := Page{ExistingDueYear: year, ExistingDueMonth: month, ExistingDueDay: day}
	renderTemplate(w, "create", &p)
}

func deleteHandler(w http.ResponseWriter, r *http.Request, title string) {
	resp, err := http.Post("http://localhost:8080/delete"+title, "application/json", bytes.NewBuffer([]byte{}))
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode != 200 {
		log.Fatal(errors.New("failed to delete"))
	}
	http.Redirect(w, r, "/readall/", http.StatusFound)
}

func updateHandler(w http.ResponseWriter, r *http.Request, title string) {
	resp, err := http.Get("http://localhost:8080/todo/" + title)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	td, err := ParseBodyToToDo(body)
	if err != nil {
		log.Fatal(err)
	}
	p := Page{Title: strings.ReplaceAll(title, " ", ""), ExistingTitle: td.Title, ExistingDescription: td.Description, ExistingDueYear: strconv.Itoa(td.Due.Year()), ExistingDueMonth: strconv.Itoa(int(td.Due.Month())), ExistingDueDay: strconv.Itoa(td.Due.Day()), ExistingDueHour: strconv.Itoa(td.Due.Hour()), ExistingPriority: strconv.Itoa(td.Priority), ExistingStatus: td.Status}
	renderTemplate(w, "update", &p)
}

func updator(update ToDoRequest, oldTitle string) {
	data, err := Jsonify(update)
	if err != nil {
		log.Fatal(err)
	}
	resp, err := http.Post("http://localhost:8080/update/"+strings.ReplaceAll(oldTitle, " ", ""), "application/json", bytes.NewBuffer(data))
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode != 200 {
		log.Fatal(errors.New("failed to update"))
	}
}

func commitCreateHandler(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("title")
	description := r.FormValue("description")
	due := r.FormValue("dueyear") + " " + r.FormValue("duemonth") + " " + r.FormValue("dueday") + " " + r.FormValue("duehour")
	priority, err := strconv.Atoi(r.FormValue("priority"))
	if err != nil {
		fmt.Println("error!")
	}
	status := r.FormValue("status")
	newToDo := ToDoRequest{title, description, due, priority, status}
	data, err := Jsonify(newToDo)
	var x ToDoRequest
	json.Unmarshal(data, &x)
	if err != nil {
		log.Fatal(err)
	}
	resp, err := http.Post("http://localhost:8080/create/", "application/json", bytes.NewBuffer(data))
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode != 200 {
		log.Fatal(errors.New("failed to create"))
	}
	http.Redirect(w, r, "/read/"+strings.ReplaceAll(newToDo.Title, " ", ""), http.StatusFound)
}

func commitUpdateHandler(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("title")
	description := r.FormValue("description")
	due := r.FormValue("dueyear") + " " + r.FormValue("duemonth") + " " + r.FormValue("dueday") + " " + r.FormValue("duehour")
	priority, err := strconv.Atoi(r.FormValue("priority"))
	if err != nil {
		fmt.Println("error!")
	}
	existingtitle := r.FormValue("existingtitle")
	status := r.FormValue("status")
	updatedToDo := ToDoRequest{title, description, due, priority, status}
	updator(updatedToDo, existingtitle)
	http.Redirect(w, r, "/read/"+strings.ReplaceAll(updatedToDo.Title, " ", ""), http.StatusFound)
}

func readAllHandler(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Get("http://localhost:8080/todos")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	p := Page{Title: "All To-dos", Body: body}
	renderTemplate(w, "readall", &p)
}

func ParseDue(d string) (time.Time, error) {
	due := strings.Split(d, "-")
	if len(due) != 3 {
		return time.Date(0, 0, 0, 0, 0, 0, 0, time.Local), errors.New("due formatted incorrectly")
	}
	dayTime := strings.Split(due[2], "T")
	hhmmss := strings.Split(dayTime[1], ":")
	year, err := strconv.Atoi(due[0])
	if err != nil {
		return time.Date(0, 0, 0, 0, 0, 0, 0, time.Local), err
	}
	month, err := strconv.Atoi(due[1])
	if err != nil {
		return time.Date(0, 0, 0, 0, 0, 0, 0, time.Local), err
	}
	day, err := strconv.Atoi(dayTime[0])
	if err != nil {
		return time.Date(0, 0, 0, 0, 0, 0, 0, time.Local), err
	}
	hour, err := strconv.Atoi(strings.Split(hhmmss[0], ":")[0])
	if err != nil {
		return time.Date(0, 0, 0, 0, 0, 0, 0, time.Local), err
	}
	return time.Date(year, time.Month(month), day, hour, 0, 0, 0, time.Local), nil
}

func ParseBodyToToDo(td []byte) (ToDo, error) {
	todo := string(td)
	tds := strings.Split(todo, `": `)
	title := tds[1][1 : len(tds[1])-19]
	description := tds[2][1 : len(tds[2])-11]
	due, err := ParseDue(tds[3][1:20])
	if err != nil {
		return ToDo{}, err
	}
	priority, err := strconv.Atoi(tds[4][:1])
	if err != nil {
		return ToDo{}, err
	}
	status := tds[5][1 : len(tds[5])-3]
	return ToDo{title, description, due, priority, status}, nil

}

func main() {
	http.HandleFunc("/create/", createHandler)
	http.HandleFunc("/read/", makeHandler(readHandler))
	http.HandleFunc("/readall/", (readAllHandler))
	http.HandleFunc("/update/", makeHandler(updateHandler))
	http.HandleFunc("/delete/", makeHandler(deleteHandler))
	http.HandleFunc("/commitupdate/", commitUpdateHandler)
	http.HandleFunc("/commitcreate/", commitCreateHandler)
	log.Fatal(http.ListenAndServe(":3000", nil))
}
