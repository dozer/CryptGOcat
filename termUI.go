package main

import (
  "log"
  "os/user"
  "fmt"
  "flag"
)

var DEFAULT_CONFIG_PATH string = "/.config/cryptGOcat/user.conf" //Default user config file paht
var confPath string

func main() {
  //DECLARE FLAGS
  usr, err := user.Current()
  if err != nil { log.Fatal(err) }
  flag.StringVar(&confPath, "conf", usr.HomeDir + DEFAULT_CONFIG_PATH, "Default user config file path")

  //Get the new user
  _, err = GetUser(confPath)
  if err != nil { panic(err) }
}

func AskUserBool(question string) bool {
  var ans byte
  fmt.Scanf(question + " [y/n]: %q", &ans)
  for ans != 'y' || ans != 'n' {
    fmt.Scanf("Please use [y/n]: %q", &ans)
  }
  if ans == 'y' {
    return true
  } else {
    return false
  }
}
