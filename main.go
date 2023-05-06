package main

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"github.com/Tnze/go-mc/bot"
	"github.com/Tnze/go-mc/bot/basic"
	"github.com/Tnze/go-mc/bot/msg"
	"github.com/Tnze/go-mc/bot/playerlist"
	"github.com/Tnze/go-mc/chat"
	"github.com/Tnze/go-mc/data/packetid"
	pk "github.com/Tnze/go-mc/net/packet"
	"log"
	"time"
)

var client *bot.Client
var player *basic.Player
var chatHandler *msg.Manager
var playerList *playerlist.PlayerList
var name string
var server string
var tag string
var secret string
var color bool
var password string

func main() {
	//name := "Bot003"
	//server := "fastmc.zziyu.top:25565"
	//secret := "tLeTleTLe"

	flag.StringVar(&name, "name", "", "Bot's username")
	flag.StringVar(&server, "server", "", "Server to connect")
	flag.StringVar(&server, "Tag", "", "Server's tag, which used to calculate the password bot use. ")
	flag.StringVar(&secret, "secret", "", "Secret used to calc password")
	flag.BoolVar(&color, "color", false, "Turn it on if you want to see colors in messages. ")
	flag.Parse()

	if name == "" {
		fmt.Print("Input your bot's username: ")
		_, err := fmt.Scanln(&name)
		if err != nil {
			return
		}
	}
	if server == "" {
		fmt.Print("Input your server address: ")
		_, err := fmt.Scanln(&server)
		if err != nil {
			return
		}
	}
	if tag == "" {
		fmt.Print("Input the server's tag: ")
		_, err := fmt.Scanln(&server)
		if err != nil {
			return
		}
	}
	if secret == "" {
		fmt.Print("Input your secret: ")
		_, err := fmt.Scanln(&secret)
		if err != nil {
			return
		}
	}

	log.Println("Name: " + name)
	log.Println("Server: " + server)
	log.Println("Tag: " + tag)
	log.Println("Secret: " + secret)

	password = calcMd5(name + tag + secret)[:30]

	log.Println(password)

	client = bot.NewClient()
	client.Auth.Name = name
	player = basic.NewPlayer(client, basic.DefaultSettings, basic.EventsListener{Death: onDeath, GameStart: onGamestart})
	playerList = playerlist.New(client)
	chatHandler = msg.New(client, player, playerList, msg.EventsHandler{
		SystemChat:        onSystemMsg,
		PlayerChatMessage: onPlayerMsg,
		DisguisedChat:     onDisguisedMsg,
	})
	//chatHandler = msg.New(client, player, playerList, msg.EventsHandler{})

	for {
		err := client.JoinServer(server)
		if err != nil {
			log.Print(err)
			time.Sleep(5 * time.Second)
			continue
		}

		log.Println("Connected. ")

		var nerr error
		var perr bot.PacketHandlerError
		for {
			if nerr = client.HandleGame(); nerr == nil {
				panic("HandleGame never return nil")
			}
			if errors.As(nerr, &perr) {
				log.Print(perr)
			} else {
				break
			}
		}
		time.Sleep(1 * time.Second)
		log.Println("Reconnecting...")
	}
}
func onSystemMsg(msg chat.Message, overlay bool) error {
	if color {
		log.Printf("System: %v", msg.String())
	} else {
		log.Printf("System: %v", msg.ClearString())
	}
	return nil
}

func onPlayerMsg(msg chat.Message, validated bool) error {
	if color {
		log.Printf("Player: %s", msg.String())
	} else {
		log.Printf("Player: %s", msg.ClearString())
	}
	return nil
}

func onDisguisedMsg(msg chat.Message) error {
	if color {
		log.Printf("Disguised: %v", msg.String())
	} else {
		log.Printf("Disguised: %v", msg.ClearString())
	}
	return nil
}

func calcMd5(s string) string {
	sum := md5.Sum([]byte(s))
	return hex.EncodeToString(sum[:])
}

func onGamestart() error {
	if err := sendCommand("register " + password + " " + password); err != nil {
		return err
	}
	if err := sendCommand("login " + password); err != nil {
		return err
	}
	return nil
}

func sendCommand(cmd string) error {
	var salt int64
	if err := binary.Read(rand.Reader, binary.BigEndian, &salt); err != nil {
		return err
	}

	err := client.Conn.WritePacket(pk.Marshal(
		packetid.ServerboundChatCommand,
		pk.String(cmd),
		pk.Long(time.Now().UnixMilli()),
		pk.Long(salt),
		pk.VarInt(0), // signature
		pk.VarInt(0),
		pk.NewFixedBitSet(20),
	))
	return err
}

func onDeath() error {
	err := player.Respawn()
	return err
}
