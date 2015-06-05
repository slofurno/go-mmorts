package main

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"
)

type OpCode int

const (
	Ping OpCode = iota
	Close
	Text
	G
)

type Websocket struct {
	Done chan bool
	conn net.Conn
	rw   *bufio.ReadWriter
}

func (ws *Websocket) Write(b []byte) (n int, err error) {
	rw := ws.rw
	length := len(b)

	rw.Write([]byte{129, byte(length)})
	rw.Write(b)
	rw.Flush()

	return length, nil

}

func ReadFrame(reader *bufio.Reader) (string, OpCode, error) {
	// | isfinal? | x x x | opcode(4) |
	// | ismask? | length(7) |
	// | mask (32) |
	header := make([]byte, 2)

	hlen, err := reader.Read(header)

	if err != nil {
		return "", G, err
	}

	fmt.Printf("header length read : %d \n", hlen)

	var isFinal = header[0] >> 7
	var opcode = header[0] & 15
	var isMasked = header[1] >> 7
	var length = int(header[1] & 127)
	fmt.Printf("raw header : %b %b \n", header[0], header[1])
	fmt.Printf("header : %d %d %d %d \n", isFinal, opcode, isMasked, length)

	if opcode == 8 {
		return "", Close, nil
	}

	//client to server always has a mask
	mask := make([]byte, 4)
	_, _ = reader.Read(mask)

	body := make([]byte, length)

	_, _ = reader.Read(body)

	for i := 0; i < length; i++ {
		/*
		   next,err := reader.ReadByte()
		   if err!=nil{
		     log.Printf("error reading frame: %v", err)
		     break
		   }*/
		//unmask
		body[i] = body[i] ^ mask[i%4]
	}

	s := string(body[:length])
	fmt.Printf("string : %s \n", s)
	return s, Text, nil

}

func WebsocketServer(w http.ResponseWriter, req *http.Request) {

	pid := req.URL.Query().Get("id")

	var guid = []byte("258EAFA5-E914-47DA-95CA-C5AB0DC85B11")
	key := req.Header.Get("Sec-WebSocket-Key")

	hash := sha1.New()
	hash.Write([]byte(key))
	hash.Write(guid)

	shaed := hash.Sum(nil)
	var challengeresponse = base64.StdEncoding.EncodeToString(shaed)

	h, _ := w.(http.Hijacker)
	conn, rw, _ := h.Hijack()
	defer conn.Close()

	buf := make([]byte, 0, 4096)

	buf = append(buf, "HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Accept: "...)
	buf = append(buf, challengeresponse...)
	buf = append(buf, "\r\n"...)
	buf = append(buf, "\r\n"...)

	rw.Write(buf)
	rw.Flush()
	conn.SetDeadline(time.Time{})

	var msg = []byte("welcome buddy")
	rw.Write([]byte{129, byte(len(msg))})
	rw.Write(msg)
	rw.Flush()

	incoming := make(chan string)
	done := make(chan bool, 2)

	//player := &Player{Out: outgoing, Done: done}
	ws := &Websocket{conn: conn, rw: rw, Done: done}

	//oldplayer, ok := connectedPlayers[pid]

	connectedPlayers[pid] = ws
	/*
		go func() {
			_, err := rw.ReadByte()
			if err == io.EOF {
				close(disconnected)
			}
		}()
	*/
	disconnected := func(reader *bufio.Reader) <-chan struct{} {
		d := make(chan struct{})
		go func() {
			for {
				fmt.Println("tevs")
				select {
				case <-done:
					fmt.Println("exiting listener")
					return
				default:
					frame, code, err := ReadFrame(reader)
					if err != nil {
						fmt.Println(err.Error())
						//done <- true
						close(d)
						return
					} else if code == Close {

						close(d)
						return
					}
					incoming <- frame
				}
			}
		}()
		return d
	}(rw.Reader)

	for {

		select {
		case <-done:
			//delete(connectedPlayers, pid)
			fmt.Println("done")
			return
		case <-disconnected:
			//delete(connectedPlayers, pid)
			fmt.Println("disconnected")
			return
		case rawcommand := <-incoming:
			fmt.Println("command", rawcommand)
			command, err := parseCommand(rawcommand)

			if err != nil {
				//fmt.Println(err.Error())
			} else {
				commandQueue <- command

			}
		}
	}

}

func parseCommand(msg string) (Command, error) {
	ct := CommandType(msg[0:3])
	rawcommand := msg[4:len(msg)]
	var command Command
	var err error

	switch ct {
	case Move:
		var move MoveCommand
		err = json.Unmarshal([]byte(rawcommand), &move)
		command = &move
	case Reinforce:
		var reinforce ReinforceCommand
		err = json.Unmarshal([]byte(rawcommand), &reinforce)
		command = &reinforce
	case Build:
		var build BuildCommand
		err = json.Unmarshal([]byte(rawcommand), &build)
		command = &build
	}

	return command, err
}
