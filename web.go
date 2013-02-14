package main

import (
	"database/sql"
	"fmt"
	_ "github.com/bmizerany/pq"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"
)

func helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello world!")
}

const logPrefix string = "/log/"

const (
	TF_LOG = "12:30 PM"
)

const page = `
{{range .}}
<li>{{.}}</li>
{{end}}
`

func logHandler(w http.ResponseWriter, r *http.Request) {
	const lenPath = len(logPrefix)
	queryString := r.URL.Path[lenPath:]

	queries := strings.Split(queryString, "/")

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	db, error := sql.Open("postgres", "user=postgres password=gdg dbname=andnyang sslmode=disable")
	if error != nil {
		log.Print(error)
		fmt.Fprintf(w, error.Error())
	}

	channel := queries[0]

	var sqlString string
	if len(channel) == 0 {
		// sqlString = "select * from andnyang_log"
		fmt.Fprintf(w, "log viewer needs specified channel.")
	} else {
		sqlString = fmt.Sprintf("select * from andnyang_log where channel='#%s'", channel)
	}

	rows, error := db.Query(sqlString)
	if error != nil {
		fmt.Fprintf(w, error.Error())
	}

	templateRows := []string{}

	for rows.Next() {
		var id int
		var date time.Time
		var channel string
		var nick string
		var message string

		error := rows.Scan(&id, &date, &channel, &nick, &message)
		if error != nil {
			fmt.Fprintf(w, error.Error())
		}
		localTime := date.Local()
		hour := localTime.Hour()
		min := localTime.Minute()
		//fmt.Fprintf(w, "<li>%d:%d [%s] %s</li>", hour, min, nick, message)
		line := fmt.Sprintf("%d:%d [%s] %s", hour, min, nick, message)
		templateRows = append(templateRows, line)
	}

	t := template.New("hello template")
	t, error = t.Parse(page)
	if error != nil {
		log.Print(error)
	}
	t.Execute(w, templateRows)
}

func main() {
	http.HandleFunc("/", helloHandler)
	http.HandleFunc(logPrefix, logHandler)
	http.ListenAndServe(":5000", nil)
}
