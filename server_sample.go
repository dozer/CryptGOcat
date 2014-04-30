//CryptGOCat is a p2p client&server system. This will act as
// a client or server depending on how you use it. It allows for several flags
// indicating a cert setup & connection setup. It uses the goalng OTR curses & my 
// TLS wrapped libraries.
//Dean Galivn & Scott Stevenson
package main

import (
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"log"
	//"strconv"
	"code.google.com/p/go.crypto/otr"
	"code.google.com/p/goncurses"
	"flag"
	"github.com/FreekingDean/gotWrap"
	"strings"
)

var s *gotWrap.Server
var c *gotWrap.Client //Server & Client

var connection = 0 //Current connection status 0-NONE 1-On Server 2-On Client

var tlscon *tls.Conn //Client TLS Connection

var conv otr.Conversation //OTR Conversation

var chatWindow *goncurses.Window
var inputWindow *goncurses.Window //Cruses Windows

type NullWriter int
func (NullWriter) Write([]byte) (int, error) { return 0, nil } //NullWriter for the logger to supress logs

//main sets up a new server/client and will start a curses session while hiding the logger
func main() {
	//Parse flag input
	addr := flag.String("addr", "", "Listening address")
	port := flag.String("port", "8000", "Listening port")
	protocol := flag.String("protocol", "tcp", "Listening protocol")
	pem := flag.String("pem", "certs/server.pem", "Cert pem file")
	key := flag.String("key", "certs/server.key", "Cert key file")
	flag.Parse()

	//Supress logger
	log.SetOutput(new(NullWriter))

	//Setup Curses Windows
	var err error
	chatWindow, err = goncurses.Init()
	if err != nil {
		log.Fatal("init:", err)
	}
	inputY, inputX := chatWindow.MaxYX()
	inputWindow, err = goncurses.NewWindow(1, inputX, inputY-1, 0)
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
	goncurses.InitPair(2, goncurses.C_RED, -1)    //ERROR
	goncurses.InitPair(3, goncurses.C_WHITE, -1)  //UNENCRYPTED [YOU]
	goncurses.InitPair(4, goncurses.C_CYAN, -1)   //ENCRYPTED   [YOU]
	goncurses.InitPair(5, goncurses.C_GREEN, -1)  //UNENCRYPTED [THEM]
	goncurses.InitPair(6, goncurses.C_MAGENTA, -1)   //ENCRYPTED   [THEM]
	chatWindow.ColorOn(1)
	chatWindow.Println("Waiting on connection...")
	chatWindow.Refresh()

	//Setup OTR
	privKey := &otr.PrivateKey{}
	privKey.Generate(rand.Reader)
	conv = otr.Conversation{PrivateKey: privKey}
	_ = otr.Conversation{PrivateKey: privKey}

	//Setup Client & Server
	s = &gotWrap.Server{
		Protocol:     *protocol,
		ListenerAddr: *addr + ":" + *port,
		PemFile:      *pem,
		KeyFile:      *key,
		MessageRec:   serverRec,
	}
	c = &gotWrap.Client{
		Protocol:   *protocol,
		RemoteAddr: *addr + ":" + *port,
		PemFile:    *pem,
		KeyFile:    *key,
		MessageRec: clientRec,
	}
	
	go readInput()
	s.CreateServer()
}

//readInput reads user input and connects, starts encryption, or sends messages
// based on '/CONNECT [ADDR]' or '/ENCRYPT'
func readInput() {
	for {
		var response string
		//Erase previous input & waits for new input
		inputWindow.Erase()
		inputWindow.Refresh()
		inputWindow.Print(":")
		response, err := inputWindow.GetString(300)
		if err != nil {
			log.Println(err)
		}
		responses := strings.Split(response, " ")
		switch responses[0] {
		//Connect checks for address & uses the client to connect
		case "/CONNECT": 
			if len(responses) == 1 {
				chatWindow.ColorOn(2)
				chatWindow.Println("Need address")
			} else {
				c.RemoteAddr = responses[1]
				c.Connect()
				c.SendMessage("/CONNECTREC")
			}
		//Encrypt just sends the other client "?OTRv2?" an OTR encrypt query
		case "/ENCRYPT":
			sendMsg([]byte("?OTRv2?"))
		//In the defualt input just send the message over!
		default:
			var toSend [][]byte
			var err error
			if !conv.IsEncrypted() {
				chatWindow.ColorOn(3)
				chatWindow.Println("[You ]:" + response)
				toSend = [][]byte{[]byte(response)}
			} else {
				chatWindow.ColorOn(4)
				chatWindow.Println("[You ]:" + response)
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
	//Checks for what connection is active
	if connection == 1 || (connection == 2 && tls != tlscon) {
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
	//Checks for what connection is active
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
	//RECIEVED OTR QUERY
	case "?OTRv2?":
		encrypt(msg)
	//RECIEVED CONNECTION START
	case "/CONNECTSEND":
		break
	//RECIEVED CONNECTION END
	case "/CONNECTREC":
		sendMsg([]byte("/CONNECTSEND"))
	//otherwise recieve the otr message
	default:
		out, _, _, toSend, err := conv.Receive([]byte(msg))
		if err != nil {
			chatWindow.Color(goncurses.C_RED)
			chatWindow.Println(err)
			chatWindow.Refresh()
		}
		sendMsgs(toSend)
		if conv.IsEncrypted() {
			chatWindow.ColorOn(6)
		} else {
			chatWindow.ColorOn(5)
		}
		chatWindow.Println("[Them]:" + string(out))
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
		for _, msgs := range toSend {
			gotWrap.SendMessage(tlscon, msgs)
		}
	} else if connection == 1 {
		for _, msgs := range toSend {
			c.SendMessage(string(msgs))
		}
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
	for _, toSend := range msgs {
		sendMsg(toSend)
	}
}
