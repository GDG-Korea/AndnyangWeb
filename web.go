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
	"io/ioutil"
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
<li>{{printf "%02d" .Hour}}:{{printf "%02d" .Min}} [{{.Nick}}] {{.Message}}</li>
{{end}}
`

type Log struct {
	Hour int
	Min int
	Nick string
	Message string
}

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

	logs := []Log{}

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
		//line := fmt.Sprintf("%d:%d [%s] %s", hour, min, nick, message)
		log := Log {
			Hour: hour,
			Min: min,
			Nick: nick,
			Message: message,
		}
		logs = append(logs, log)
	}

	filename := "log.html"
	body, error := ioutil.ReadFile(filename)
	if error != nil {
		body = []byte(page)
	}	

	t := template.New("hello template")
	t, error = t.Parse(string(body))
	if error != nil {
		log.Print(error)
	}

	t.Execute(w, logs)
}

func main() {
	http.HandleFunc("/", helloHandler)
	http.HandleFunc(logPrefix, logHandler)

	http.Handle("/css/", http.StripPrefix("/css", http.FileServer(http.Dir("./css"))))
	http.Handle("/js/", http.StripPrefix("/js", http.FileServer(http.Dir("./js"))))
	http.Handle("/img/", http.StripPrefix("/img", http.FileServer(http.Dir("./img"))))

	http.ListenAndServe(":5000", nil)
}
