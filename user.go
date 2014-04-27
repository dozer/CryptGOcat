package main

import (
  //"io"
  "log"
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
  Config      *conf.ConfigFile
  ConfigPath  string
  UserName    string
  password    string
  BuddyList   map[string]string
  authed      bool
}

//GetUser returns a user from the config or creates one if none exists
func GetUser(path string) (user *User, err error) {
  log.Println("user.go:GetUser:", "Getting new user")

  user = &User{}
  user.password = AskUserString("please enter a password"
  user.ConfigPath = path
  user.BuddyList = make(map[string]string)
  user.Config, err = OpenConfig(path)
  if err != nil { return nil, err }

  if user.Config.AddSection("user") {
    log.Println("user.go:GetUser:", "No Section \"user\" found, adding...")
    if !AskUserBool("No user found, add new user?") { return nil, errors.New("Can't continue without user") }
    user.UserName = "CryptGoCatUsr1"
    exists := user.Config.AddOption("user", "user_name", user.UserName)
    log.Println("user.go:GetUser:", "Adding option \"user_name\"")
    if !exists { return nil, errors.New("Key exists but section does not") }
  } else {
    log.Println("user.go:GetUser:", "Section \"user\" found")
    log.Println("user.go:GetUser:", "Getting option \"user_name\"")
    user.UserName, err = user.Config.GetString("user", "user_name")
    if err != nil { return nil, err }
  }

  if user.Config.AddSection("buddies") {
    log.Println("user.go:GetUser:", "No Section \"buddies\" found, adding...")
  } else {
    log.Println("user.go:GetUser:", "Section \"user\" found")
    log.Println("user.go:GetUser:", "Getting buddyList")
    user.GetBuddies()
    if err != nil { return nil, err }
  }
  return
}

//GetBuddies looks through the section "buddies" in the config file
// and creates a hash map of aliases/buddies
func (user *User) GetBuddies() error {
  buddies, err := user.Config.GetOptions()
  if err != nil { return err }
  for _, address := range buddies {
    user.BuddyList[address], err  = user.Config.GetOption("buddies", address)
    if err != nil { return err }
  }
  return
}

//SaveUser saves the current user to the config path with permissions 0655
func (user *User) SaveUser() error {
  //user.saveStruct()
  return user.Config.WriteConfigFile(user.ConfigPath, 0655, "User config for CryptGOcat")
}

//TODO - Save other items in struct
//SetName sets the UserName for the given user
func (user *User) saveStruct() error {
  user.Config.AddOption("user", "user_name", user.UserName)
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

func GetKeyVault(path string) {
  if !Exists(path) {
    setupNewKeyVault()
    return
  }
}

func SetupKeyVault() {
  newKey, correct := strongbox.GenerateKey()
  fmt.Println(correct)
  hash := sha512.New()
  io.WriteString(hash, )
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
