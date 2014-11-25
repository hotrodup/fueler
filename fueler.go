package main

import (
    "net/http"
    "os"
    "io"
    "time"
    "io/ioutil"
    "fmt"
    fp "path/filepath"
)

const (
  UPDATE_FILE = ".hotrod-update"
  BASE_DIR = "/app/"
)

func handler(w http.ResponseWriter, r *http.Request, add, isFile bool) {
  // after 200000 bytes of file parts stored in memory
  // the remainder are persisted to disk in temporary files
  err := r.ParseMultipartForm(200000)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }

  formData := r.MultipartForm
  relPath := formData.Value["path"][0]

  filepath := fp.Join(BASE_DIR, relPath)

  if add {

    dir := filepath
    if isFile {
      dir = fp.Dir(filepath)
    }

    err = os.MkdirAll(dir, 0777)
    if err != nil {
      http.Error(w, "Unable to create the folder for writing. Check your write access privilege", http.StatusInternalServerError)
      return
    }

    if isFile {
      file := formData.File["file"][0]

      f, err := file.Open()
      defer f.Close()
      if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
      }

      dest, err := os.Create(filepath)
      defer dest.Close()
      if err != nil {
        http.Error(w, "Unable to create the file for writing. Check your write access privilege", http.StatusInternalServerError)
        return
      }

      _, err = io.Copy(dest, f)
      if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
      }
    }

  } else {

    err = os.RemoveAll(filepath)
    if err != nil {
      http.Error(w, "Unable to delete the file/folder. Check your write access privilege", http.StatusInternalServerError)
      return
    }
  }

  os.Chtimes(fp.Join(BASE_DIR, UPDATE_FILE), time.Now(), time.Now())
}

func baseHandler(w http.ResponseWriter, r *http.Request) {
  fmt.Fprintf(w, "Hello, server running.")
}

func addFileHandler(w http.ResponseWriter, r *http.Request) {
  handler(w, r, true, true)
}

func addFolderHandler(w http.ResponseWriter, r *http.Request) {
  handler(w, r, true, false)
}

func removeHandler(w http.ResponseWriter, r *http.Request) {
  handler(w, r, false, false)
}

func main() {
    ioutil.WriteFile(fp.Join(BASE_DIR, UPDATE_FILE), []byte(""), 0777)

    http.HandleFunc("/", baseHandler)
    http.HandleFunc("/addFile", addFileHandler)
    http.HandleFunc("/addFolder", addFolderHandler)
    http.HandleFunc("/remove", removeHandler)
    http.ListenAndServe(":8888", nil)
}
