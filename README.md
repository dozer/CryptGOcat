#CryptGOcat

An encrypted instant messaging and file sharing service. Inspired by Cryptocat, written in Golang.

##Installation

###Download
`wget -O ChatServer.go https://raw.githubusercontent.com/oodabaga/CryptGOcat/master/server_sample.go`

or

`curl -L https://raw.githubusercontent.com/oodabaga/CryptGOcat/master/server_sample.go >> ChatServer.go`

###Build
`go build ChatServer.go`

##Use

###Connect
`/CONNECT [ip:port]`

###Encrypt
`/ENCRYPT`
