package main

import (
  "fmt"
  "net/http"
  "time"
)

func timeHandler(w http.ResponseWriter, r *http.Request) {
  t := time.Now()
  if r.Method == http.MethodPost {
    w.WriteHeader(http.StatusCreated)
  } else {
    w.WriteHeader(http.StatusOK)
  }
  fmt.Fprintf(w, "The current time is: %s\n", t.Format(time.RFC1123))
}

func main() {
  http.HandleFunc("/time", timeHandler)
  http.ListenAndServe(":8001", nil)
}
