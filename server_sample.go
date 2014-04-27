package main

import (
  "fmt"
  "log"
	"crypto/tls"
  "crypto/rand"
  //"strconv"
	"flag"
  "strings"
	"github.com/FreekingDean/gotWrap"
  "code.google.com/p/go.crypto/otr"
  "code.google.com/p/goncurses"
)

var s *gotWrap.Server
var c *gotWrap.Client
var connection = 0
var tlscon *tls.Conn
var conv otr.Conversation
var chatWindow *goncurses.Window
var inputWindow *goncurses.Window
type NullWriter int
func (NullWriter) Write([]byte) (int, error) { return 0, nil }

//main sets up a new server/client and will start a curses session while hiding the logger
func main() {
	addr := flag.String("addr", "", "Listening address")
	port := flag.String("port", "8000", "Listening port")
	protocol := flag.String("protocol", "tcp", "Listening protocol")
	pem := flag.String("pem", "certs/server.pem", "Cert pem file")
	key := flag.String("key", "certs/server.key", "Cert key file")
	flag.Parse()
	s = &gotWrap.Server{
		Protocol:     *protocol,
		ListenerAddr: *addr + ":" + *port,
		PemFile:      *pem,
		KeyFile:      *key,
		MessageRec:   serverRec,
	}
  c = &gotWrap.Client{
    Protocol:     *protocol,
    RemoteAddr:   *addr + ":" + *port,
    PemFile:      *pem,
    KeyFile:      *key,
    MessageRec:   clientRec,
  }
  privKey := &otr.PrivateKey{}
  privKey.Generate(rand.Reader)
  conv = otr.Conversation{PrivateKey: privKey}
  _ = otr.Conversation{PrivateKey: privKey}
  var err error
  chatWindow, err = goncurses.Init()
  if err != nil {
    log.Fatal("init:", err)
  }
  inputY, inputX := chatWindow.MaxYX()
  inputWindow, err = goncurses.NewWindow(1, inputX,  inputY-1, 0)
  chatWindow.Resize(inputY-1, inputX)
  chatWindow.ScrollOk(true)
  inputWindow.Println("Waiting...")
  inputWindow.Refresh()
  if err != nil {
    log.Fatal("init:", err)
  }
  defer goncurses.End()
  if err := goncurses.StartColor(); err != nil {
    log.Fatal(err)
  }
  goncurses.UseDefaultColors()
  goncurses.InitPair(1, goncurses.C_YELLOW, -1) //INFO
  goncurses.InitPair(2, goncurses.C_RED, -1) //ERROR
  goncurses.InitPair(3, goncurses.C_GREEN, -1) //UNENCRYPTED
  goncurses.InitPair(4, goncurses.C_CYAN, -1) //ENCRYPTED
  chatWindow.ColorOn(1)
  chatWindow.Println("Waiting on connection...")
  log.SetOutput(new(NullWriter))
  chatWindow.Refresh()
  go readInput()
	s.CreateServer()
}

//readInput reads user input and connects or starte encryption based
// on '/CONNECT [ADDR]' or '/ENCRYPT'
func readInput() {
	for {
		var response string
    inputWindow.Erase()
    inputWindow.Refresh()
    inputWindow.Print(":")
    response, err := inputWindow.GetString(300)
    if err != nil {
      log.Println(err)
    }
    responses := strings.Split(response, " ")
		switch responses[0] {
    case "/CONNECT":
      if len(responses) == 1 {
        chatWindow.ColorOn(2)
        chatWindow.Println("Need address")
      } else {
        c.RemoteAddr = responses[1]
        c.Connect()
        c.SendMessage("/CONNECTREC")
      }
    case "/ENCRYPT":
      sendMsg([]byte("?OTRv2?"))
    default:
      var toSend [][]byte
      var err error
      if !conv.IsEncrypted() {
        chatWindow.ColorOn(3)
        chatWindow.Println("[You ]:"+response)
        toSend = [][]byte{[]byte(response)}
      } else {
        chatWindow.ColorOn(4)
        chatWindow.Println("[You ]:"+response)
        toSend, err = conv.Send([]byte(response))
        if err != nil {
          log.Println(err)
        }
      }
      sendMsgs(toSend)
    }
    chatWindow.Refresh()
    inputWindow.Refresh()
  }
}

//serverRec recieves a message from a client an either ignores it for its
// current connection, starts a new one, or throws it to the reciever
func serverRec(tls *tls.Conn, msg string) {
  if connection == 1 || (connection == 2 && tls != tlscon){
    chatWindow.ColorOn(1)
    chatWindow.Println("Already connected on the client & getting attempts from server")
  } else if connection == 2 && tlscon == tls {
    recMsg(msg)
  } else if connection == 0 {
    tlscon = tls
    connection = 2
    recMsg(msg)
  }
}

//clientRec does the same as serverRec but from the client side 
func clientRec(msg string) {
  if connection == 2 {
    chatWindow.ColorOn(1)
    chatWindow.Println("Already connected on the server & getting attempts from client")
  } else if connection == 1 {
    recMsg(msg)
  } else if connection == 0 {
    connection = 1
    recMsg(msg)
  }
  chatWindow.Refresh()
  inputWindow.Refresh()
}

//recMsg recieves messages from the connection and relays a pingpong or sets up
// otr encryption
func recMsg(msg string) {
  message := strings.Split(msg, " ")
  switch message[0] {
  case "?OTRv2?":
    encrypt(msg)
  case "/CONNECTSEND":
    break
  case "/CONNECTREC":
    sendMsg([]byte("/CONNECTSEND"))
  default:
    out, _, _, toSend, err := conv.Receive([]byte(msg))
    if err != nil {
      chatWindow.Color(goncurses.C_RED)
      chatWindow.Println(err)
      chatWindow.Refresh()
    }
    sendMsgs(toSend)
    if conv.IsEncrypted() {
      chatWindow.ColorOn(4)
    } else {
      chatWindow.ColorOn(3)
    }
    chatWindow.Println("[Them]:"+string(out))
  }
  chatWindow.Refresh()
}

//encrypt start otr encryption based on the otr query
func encrypt(msg string) {
  _, _, _, toSend, err := conv.Receive([]byte(msg))
  if err != nil {
    fmt.Println(err)
  }
  if connection == 2 {
    for _, msgs := range toSend { gotWrap.SendMessage(tlscon, msgs) }
  } else if connection == 1 {
    for _, msgs := range toSend { c.SendMessage(string(msgs)) }
  }
}

//sendMsg sends a single message to the current connection
func sendMsg(msg []byte) {
  if connection == 0 {
    chatWindow.ColorOn(1)
    chatWindow.Println("NOT CONNECTED")
    chatWindow.Refresh()
  } else if connection == 1 {
    c.SendMessage(string(msg))
  } else if connection == 2 {
    gotWrap.SendMessage(tlscon, msg)
  }
}

//sendMsgs sends multipule messages to the current connection
func sendMsgs(msgs [][]byte) {
  for _, toSend := range msgs { sendMsg(toSend) }
}
