package main

import (
  //"io"
  "errors"
  "strings"
  "os"
  "code.google.com/p/goconf"
  //"crypto/sha512"
  //"github.com/gokyle/keyvault"
  //"github.com/cryptobox/gocryptobox/strongbox"
)

//User is a basic user including his name and config file.
type User struct {
  Config    *conf.ConfigFile
  UserName  string
  authed    bool
}

//GetUser returns a user from the config or creates one if none exists
func GetUser(path string) (user *User, err error) {
  user = &User{}
  user.Config, err = OpenConfig(path)
  if err != nil { return nil, err }
  if !user.Config.AddSection("user") {
    //Asks to make the user
    if !AskUserBool("No user found, add new user?") { return nil, errors.New("Can't continue without user") }
    user.UserName = "CryptGoCatUsr1"
    exists := user.Config.AddOption("user", "user_name", user.UserName)
    //This should never run ever
    if exists {
      return nil, errors.New("Key exists but section does not")
    } else {
      user.UserName, err = user.Config.GetString("user", "user_name")
      if err != nil { return nil, err }
    }
  }
  return
}

//OpenConfig opens a config from a path and creates one along with
// its directories if they don't exist. Throws back any errors that
// are thrown to it.
func OpenConfig(path string) (config *conf.ConfigFile, err error) {
  config, err = conf.ReadConfigFile(path)
  if err == nil {
    return config, nil
  } else {
    if _, ok := err.(*os.PathError); ok {
      config, err = AddConfig(path)
      if err != nil { return nil, err }
      return config, nil
    } else {
      return nil, err
    }
  }
  return
}

//AddConfig adds a new config file with the specified path. If the
// path doesn't exist it will create nesecary folders in its path.
func AddConfig(path string) (config *conf.ConfigFile, err error) {
  folders := strings.Split(path, "/")
  folders = folders[0:len(folders)-1]
  configPath := strings.Join(folders, "/")
  found, err := Exists(configPath)
  if err != nil { return nil, err }
  if !found {
    err = AddFolders(configPath)
    if err != nil { return nil, err }
  }
  config = conf.NewConfigFile()
  return config, nil
}

/*
func not_main() {
  newKey, correct := strongbox.GenerateKey()
  fmt.Println(correct)
  hash := sha512.New()
  io.WriteString(hash, "The fog is getting thicker!")
  shasum := hash.Sum(nil)
  var vaultKey [144]byte
  for index, curByte := range shasum {
    vaultKey[index] = curByte
  }
  for index, curByte := range newKey {
    vaultKey[index+64] = curByte
  }
  k := keyvault.New("./test.store", vaultKey, []byte("SALT YER HASHES"))
  var md keyvault.Metadata
  md = make(map[string]string)
  md["password"] = "testPass"
  ctx := &keyvault.Context{
    Label: "TEST LABEL",
    Metadata: md,
  }
  correct = k.InitialContext(ctx)
  fmt.Println(correct)
  err := k.Close()
  fmt.Println(err)
}

func checkPassword(ctx keyvault.Context, authInfo ...interface{}) bool {
    var (
        ctxpass string
        ok       bool
    )

    if len(authInfo) != 1 {
        return false
    } else if ctx.Metadata == nil {
        return false
    } else if ctxpass, ok = ctx.Metadata["password"]; !ok {
            return false
    }

    switch authInfo[0].(type) {
    case string:
        return authInfo[0].(string) == ctxpass
    default:
        return false
    }
}
*/
