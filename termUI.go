package main

import (
  "log"
  "os/user"
  "fmt"
  "flag"
)

var DEFAULT_CONFIG_PATH string = "/.config/cryptGOcat/" //Default user config file paht
var confPath string

//TODO - Graceully handle errors
//TermUI is a simple stub to test requests and handle users
func main() {
  //DECLARE FLAGS
  sysUsr, err := user.Current()
  if err != nil { log.Fatal(err) }
  flag.StringVar(&confPath, "conf", sysUsr.HomeDir + DEFAULT_CONFIG_PATH, "Default user config file path")

  //Get the new user
  user, err := GetUser(confPath+"user.conf")
  if err != nil { panic(err) }

  //Asks the user to input a new name
  err, newName := AskString("Please enter a user name:")
  if err != nil { panic(err) }

  user.UserName = newName

  err = user.SaveUser()
  if err != nil { panic(err) }
}

//
func AskUserString(question string) string {
  log.Println("termUI.go:", "AskUserBool:", "Asking User user stuffs")
  var and string
  fmt.Printf(question)
  fmt.Scanf("%s\n", &ans)
  return ans
}

//AskUserBool
func AskUserBool(question string) bool {
  log.Println("termUI.go:","AskUserBool:", "Asking User user stuffs")
  var ans string
  fmt.Printf("%s [y/n]: ", question)
  fmt.Scanf("%s\n", &ans)
  log.Println("termUI.go:","AskUserBool:", "read", fmt.Sprintf("%v", ans))
  for (ans != "y" && ans != "n") {
    fmt.Printf("Please use [y or n]: 111")
    fmt.Scanf("%s\n", &ans)
    log.Println("termUI.go:","AskUserBool:", "read", fmt.Sprintf("%v", ans))
  }
  if ans == "y" {
    return true
  } else {
    return false
  }
}
