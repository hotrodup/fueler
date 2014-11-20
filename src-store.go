package main

import (
    "fmt"
    "net/http"
    "os"
    "io"
    fp "path/filepath"
)

func uploadHandler(w http.ResponseWriter, r *http.Request) {
  // after 200000 bytes of file parts stored in memory
  // the remainder are persisted to disk in temporary files
  err := r.ParseMultipartForm(200000)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }

  // get all the *fileheaders uploaded to the input field
  // named "src"
  formData := r.MultipartForm
  files := formData.File["src"]

  // loop through files
  for i, _ := range files {
    file, err := files[i].Open()
    defer file.Close()
    if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
    }

    filename := files[i].Filename
    path := r.URL.Path[1:]
    baseDir := os.Getenv("APP_SRC_DIR")
    if len(baseDir) == 0 {
      baseDir = "/app/"
    }
    fullPath := fp.Join(baseDir, path, filename)

    err = os.MkdirAll(fp.Join(baseDir, path), 0777)
    if err != nil {
      http.Error(w, "Unable to create the folder for writing. Check your write access privilege", http.StatusInternalServerError)
      return
    }

    dest, err := os.Create(fullPath)
    defer dest.Close()
    if err != nil {
      http.Error(w, "Unable to create the file for writing. Check your write access privilege", http.StatusInternalServerError)
      return
    }

    _, err = io.Copy(dest, file)

    if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
    }

    fmt.Fprintf(w,"Files uploaded successfully : ")
    fmt.Fprintf(w, "%s\n", fullPath)

  }
}

func handler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func main() {
    http.HandleFunc("/", uploadHandler)
    http.ListenAndServe(":8080", nil)
}
