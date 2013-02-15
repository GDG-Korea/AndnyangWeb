package main

import (
	"database/sql"
	"fmt"
	_ "github.com/bmizerany/pq"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
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

type Log struct {
	Hour    int
	Min     int
	Nick    string
	Message string
}

func getSurfixQuery(year int, month time.Month, day int) string {
	const TF_SQL = "20060102 15:04:05"
	const TF_CALENDAR = "20060102 15:04:05 -0700"
	st, _ := time.Parse(TF_CALENDAR, fmt.Sprintf("%04d%02d%02d 00:00:00 +0900", year, month, day))
	st = st.UTC()
	et := st.AddDate(0, 0, 1)
	return fmt.Sprintf(" and date between '%s' and '%s'", st.Format(TF_SQL), et.Format(TF_SQL))
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

	if len(channel) == 0 {
		return
	}

	sqlString := fmt.Sprintf("select * from andnyang_log where channel='#%s'", channel)

	if len(queries) != 2 || len(queries[1]) != 8 {
		now := time.Now().Local()
		year := now.Year()
		month := now.Month()
		day := now.Day()
		path := fmt.Sprintf("/log/%s/%04d%02d%02d", channel, year, month, day)
		http.Redirect(w, r, path, http.StatusFound)
		return
	} else {
		dateQuery := queries[1]
		year, _ := strconv.Atoi(dateQuery[0:4])
		monthNumber, _ := strconv.Atoi(dateQuery[4:6])
		month := time.Month(monthNumber)
		day, _ := strconv.Atoi(dateQuery[6:])
		surfix := getSurfixQuery(year, month, day)
		sqlString = sqlString + surfix
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

		log := Log{
			Hour:    hour,
			Min:     min,
			Nick:    nick,
			Message: message,
		}
		logs = append(logs, log)
	}

	filename := "log.html"
	body, error := ioutil.ReadFile(filename)
	if error != nil {
		log.Print(error)
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
