package main

import (
	"database/sql"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func createTable() {
	connStr := "user=postgres dbname=s2 sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Println(err)
	}

	const query = `
		CREATE TABLE IF NOT EXISTS users (
		  id SERIAL PRIMARY KEY,
		  first_name TEXT,
		  last_name TEXT
	)`

	_, err = db.Exec(query)
	if err != nil {
		fmt.Println(err)
		return
	}
	db.Close()
}

func dropTable() {
	connStr := "user=postgres dbname=s2 sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = db.Exec("DROP TABLE IF EXISTS users")
	if err != nil {
		fmt.Println(err)
		return
	}
	db.Close()
}

func insertRecord(query string) {
	connStr := "user=postgres dbname=s2 sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = db.Exec(query)
	if err != nil {
		fmt.Println(err)
		return
	}
	db.Close()
}

func Test_count(t *testing.T) {
	var count int
	createTable()

	insertRecord("INSERT INTO users (first_name, last_name) VALUES ('John', 'Doe')")
	insertRecord("INSERT INTO users (first_name, last_name) VALUES ('Mihalis', 'Tsoukalos')")
	insertRecord("INSERT INTO users (first_name, last_name) VALUES ('Marko', 'Anastasov')")

	connStr := "user=postgres dbname=s2 sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Println(err)
		return
	}

	row := db.QueryRow("SELECT COUNT(*) FROM users")
	err = row.Scan(&count)
	db.Close()

	if count != 3 {
		t.Errorf("Select query returned %d", count)
	}
	dropTable()
}

func Test_queryDB(t *testing.T) {
	createTable()

	connStr := "user=postgres dbname=s2 sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Println(err)
		return
	}

	query := "INSERT INTO users (first_name, last_name) VALUES ('Random Text', '123456')"
	insertRecord(query)

	rows, err := db.Query(`SELECT * FROM users WHERE last_name=$1`, `123456`)
	if err != nil {
		fmt.Println(err)
		return
	}
	var col1 int
	var col2 string
	var col3 string
	for rows.Next() {
		rows.Scan(&col1, &col2, &col3)
	}
	if col2 != "Random Text" {
		t.Errorf("first_name returned %s", col2)
	}

	if col3 != "123456" {
		t.Errorf("last_name returned %s", col3)
	}

	db.Close()
	dropTable()
}

func Test_record(t *testing.T) {
	createTable()
	insertRecord("INSERT INTO users (first_name, last_name) VALUES ('John', 'Doe')")

	req, err := http.NewRequest("GET", "/getdata", nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(getData)
	handler.ServeHTTP(rr, req)

	status := rr.Code
	if status != http.StatusOK {
		t.Errorf("Handler returned %v", status)
	}

	if rr.Body.String() != "<h3 align=\"center\">1, John, Doe</h3>\n" {
		t.Errorf("Wrong server response!")
	}
	dropTable()
}

func Test_myHandler_root(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	myHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
	expected := "Serving: /\n"
	if rr.Body.String() != expected {
		t.Errorf("expected %q, got %q", expected, rr.Body.String())
	}
}

func Test_myHandler_customPath(t *testing.T) {
	req := httptest.NewRequest("GET", "/hello/world", nil)
	rr := httptest.NewRecorder()
	myHandler(rr, req)

	if !strings.Contains(rr.Body.String(), "/hello/world") {
		t.Errorf("response should contain the request path, got %q", rr.Body.String())
	}
}

func Test_timeHandler_containsTime(t *testing.T) {
	req := httptest.NewRequest("GET", "/time", nil)
	rr := httptest.NewRecorder()
	timeHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "The current time is:") {
		t.Errorf("response should contain time header, got %q", body)
	}
	if !strings.Contains(body, "Serving: /time") {
		t.Errorf("response should contain serving path, got %q", body)
	}
}

func Test_timeHandler_HTMLStructure(t *testing.T) {
	req := httptest.NewRequest("GET", "/time", nil)
	rr := httptest.NewRecorder()
	timeHandler(rr, req)

	body := rr.Body.String()
	if !strings.Contains(body, "<h1") {
		t.Errorf("response should contain h1 tag, got %q", body)
	}
	if !strings.Contains(body, "<h2") {
		t.Errorf("response should contain h2 tag, got %q", body)
	}
}

func Test_myHandler_methodPOST(t *testing.T) {
	req := httptest.NewRequest("POST", "/submit", nil)
	rr := httptest.NewRecorder()
	myHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200 for POST, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "/submit") {
		t.Errorf("response should contain path for POST request")
	}
}

func Test_timeHandler_statusCode(t *testing.T) {
	req := httptest.NewRequest("GET", "/time", nil)
	rr := httptest.NewRecorder()
	timeHandler(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, rr.Code)
	}
}

func Test_timeHandler_responseFreshness(t *testing.T) {
	before := time.Now()

	jitter := time.Duration(rand.Intn(1200)) * time.Millisecond
	time.Sleep(jitter)

	req := httptest.NewRequest("GET", "/time", nil)
	rr := httptest.NewRecorder()
	timeHandler(rr, req)

	after := time.Now()
	body := rr.Body.String()

	start := strings.Index(body, "<h2 align=\"center\">")
	end := strings.Index(body, "</h2>")
	if start == -1 || end == -1 {
		t.Fatal("could not find time in response body")
	}
	timeStr := body[start+len("<h2 align=\"center\">") : end]

	parsedTime, err := time.Parse(time.RFC1123, timeStr)
	if err != nil {
		t.Fatalf("could not parse time from response: %v", err)
	}

	if parsedTime.Before(before.Truncate(time.Second)) || parsedTime.After(after.Add(1*time.Second)) {
		t.Errorf("response time %v is not between %v and %v", parsedTime, before, after)
	}
}
