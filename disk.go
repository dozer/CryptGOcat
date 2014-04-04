package main

import "os"

//Exists checks if the file or folder at the given path exists
func Exists(path string) (bool, error) {
    _, err := os.Stat(path)
    if err == nil { return true, nil }
    if os.IsNotExist(err) { return false, nil }
    return false, err
}

//AddFolder adds a new folder and any predesecors at the given path
func AddFolders(path string) error {
  return os.MkdirAll(path, 0755)
}
