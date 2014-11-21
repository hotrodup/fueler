package main

import (
    "net/http"
    "os"
    "io"
    "time"
    "io/ioutil"
    fp "path/filepath"
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

  baseDir := os.Getenv("APP_SRC_DIR")
  if len(baseDir) == 0 {
    baseDir = "/app/"
  }

  filepath := fp.Join(baseDir, relPath)

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

  os.Chtimes(fp.Join(baseDir, ".hotrod-update"), time.Now(), time.Now())

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

    baseDir := os.Getenv("APP_SRC_DIR")
    if len(baseDir) == 0 {
      baseDir = "/app/"
    }

    ioutil.WriteFile(fp.Join(baseDir, ".hotrod-update"), []byte(""), 0777)

    http.HandleFunc("/addFile", addFileHandler)
    http.HandleFunc("/addFolder", addFolderHandler)
    http.HandleFunc("/remove", removeHandler)
    http.ListenAndServe(":8888", nil)
}
