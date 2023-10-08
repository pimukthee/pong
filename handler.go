package main

import "net/http"

func serveHome(w http.ResponseWriter, r *http.Request) {
  if r.Method != http.MethodGet {
    http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    return
  }
  http.ServeFile(w, r, "home.html")
}

func newHandler() *http.ServeMux {
  handler := http.NewServeMux()
  handler.HandleFunc("/", serveHome)

  return handler
}
