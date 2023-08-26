/*
Usage:

	go run server.go

See https://gist.github.com/winlinvip/e8665ba888e2fd489ccd5a449fadfa73
See https://stackoverflow.com/a/70576851/17679565
See https://github.com/ossrs/srs/issues/2843
*/
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/Acidic9/go-steam/steamid"
	"github.com/folospace/go-mysql-orm/orm"

	csgolog "github.com/janstuemmel/csgo-log"
)

var db, _ = orm.OpenMysql("root:@tcp(127.0.0.1:3306)/levelx?parseTime=true&charset=utf8mb4&loc=Asia%2FShanghai")

// *************** PLAYER OPERATIONS ***************/

// user table model
var PlayerTable = orm.NewQuery(Player{}, db)

type Player struct {
	Id       int    `json:"id"`
	TeamId   int    `json:"team_id"`
	StreamId string `json:"stream_id"`
	Side     string `json:"side"`
	Name     string `json:"player_name"`
}

// Table interface: implements two methods below
func (Player) TableName() string {
	return "players"
}
func (Player) DatabaseName() string {
	return "levelx"
}

// **************END OFPLAYER OPERATIONS ***************/
func main() {
	serverPort := 5001

	palyerid := steamid.NewID64(76561198174434951)

	fmt.Println("32-bit Value:", palyerid.To3().String())
	fmt.Println("32-bit Value:", palyerid.ToID())

	if len(os.Args) > 1 {
		if v, err := strconv.Atoi(os.Args[1]); err != nil {
			fmt.Printf("Invalid port %v, err %v", os.Args[1], err)
			os.Exit(-1)
		} else {
			serverPort = v
		}
	}

	addr := net.UDPAddr{
		Port: serverPort,
		IP:   net.ParseIP("0.0.0.0"),
	}
	server, err := net.ListenUDP("udp", &addr)
	if err != nil {
		fmt.Printf("Listen err %v\n", err)
		os.Exit(-1)
	}
	fmt.Printf("Listen at %v\n", addr.String())

	for {
		p := make([]byte, 1024)
		nn, raddr, err := server.ReadFromUDP(p)
		if err != nil {
			fmt.Printf("Read err  %v", err)
			continue
		}

		msg := p[:nn]
		fmt.Printf("Received %v : %s\n", raddr, msg)

		go response(server, raddr, msg)
	}
}

func response(udpServer net.PacketConn, addr net.Addr, buf []byte) {
	line := string(buf)
	data := parse(line)
	jsn := csgolog.ToJSON(data)
	handleEvent(data.GetType(), data)
	jsn = jsn + "\n"
	writeToFile(jsn)
	udpServer.WriteTo(cleanByteData([]byte(jsn)), addr)
}

func parse(line string) csgolog.Message {
	var msg csgolog.Message
	msg, err := csgolog.Parse(line)
	if err != nil {
		fmt.Println(err)
		return nil
	} else {
		fmt.Println(msg.GetType(), msg.GetTime().String())
		return msg
	}
}

func cleanByteData(input []byte) []byte {
	cleanedData := make([]byte, 0, len(input))
	for _, b := range input {
		if b != 0 {
			cleanedData = append(cleanedData, b)
		}
	}
	return cleanedData
}

func writeToFile(content string) {
	currentDate := time.Now()
	dateString := currentDate.Format("2006-01-02") // Format: YYYY-MM-DD
	// currentDate := time.Now().Format("2023-08-01") // Format: YYYY-MM-DD

	// Create the file name with today's date
	fileName := dateString + "-log.log"
	// Open the file in append mode
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	// Append the content to the file
	_, err = file.WriteString(content)
	if err != nil {
		fmt.Println("Error appending content:", err)
		return
	}
}

func convertTo32Bit(playerStreamID string) uint32 {
	// Calculate the SHA-256 hash of the player stream ID
	sha256Hash := sha256.Sum256([]byte(playerStreamID))

	// Take the first 4 bytes of the SHA-256 hash and convert them to a 32-bit unsigned integer
	value32Bit := binary.LittleEndian.Uint32(sha256Hash[:4])

	return value32Bit
}

func handleEvent(eventKey string, json csgolog.Message) {

	switch eventKey {
	case "ServerMessage":
	case "FreezTimeStart":
	case "WorldMatchStart":
		httpApi("match-start", csgolog.ToJSON(json))

	case "WorldRoundRestart":
		httpApi("round-start", csgolog.ToJSON(json))

	case "WorldRoundStart":
		httpApi("round-start", csgolog.ToJSON(json))
	case "WorldRoundEnd":
		httpApi("round-end", csgolog.ToJSON(json))
	case "WorldGameCommencing":
	case "TeamScored":
	case "TeamNotice":
	case "PlayerConnected":
		httpApi("player-connect", csgolog.ToJSON(json))

	case "PlayerDisconnected":
		httpApi("player-disconnect", csgolog.ToJSON(json))
	case "PlayerEntered":
		httpApi("player-connect", csgolog.ToJSON(json))
	case "PlayerBanned":
	case "PlayerSwitched":
		httpApi("player-connect", csgolog.ToJSON(json))
	case "PlayerSay":
	case "PlayerPurchase":
	case "PlayerKill":
		httpApi("player-kill", csgolog.ToJSON(json))
	case "PlayerKillAssist":
		httpApi("player-assist", csgolog.ToJSON(json))
	case "PlayerAttack":
		httpApi("player-attack", csgolog.ToJSON(json))
	case "PlayerKilledBomb":
	case "PlayerKilledSuicide":
	case "PlayerPickedUp":
	case "PlayerDropped":
	case "PlayerMoneyChange":
	case "PlayerBombGot":
	case "PlayerBombPlanted":
	case "PlayerBombDropped":
	case "PlayerBombBeginDefuse":
	case "PlayerBombDefused":
	case "PlayerThrew":
	case "PlayerBlinded":
	case "ProjectileSpawned":
	case "Unknown":
	case "GameOver":
		httpApi("match-end", csgolog.ToJSON(json))
	}
}

func httpApi(endpoint string, jsonData string) {
	// url := "https://vccpvc59e6.execute-api.ap-south-1.amazonaws.com/api/admin/" + endpoint

	url := "http://127.0.0.1:8000/api/admin/" + endpoint
	jsonRequest := []byte(jsonData)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonRequest))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	fmt.Println("Response Status:", resp.Status)
	fmt.Println("Response Body:", string(body))

}
