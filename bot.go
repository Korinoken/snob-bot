package main

import (
	"flag"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	//	"time"
	"time"
)

func init() {
	flag.StringVar(&token, "t", "", "Bot Token")
	flag.Parse()
}

var token string

func main() {
	if token == "" {
		fmt.Println("No token provided. Please run: snob-bot -t <bot token>")
		return
	}
	dg, err := discordgo.New("Bot " + token)
	dg.AddHandler(messageCreate)
	if err != nil {
		fmt.Println("Error creating Discord session: ", err)
		return
	}
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	dg.Close()
}
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	if strings.HasPrefix(m.Content, "!help") {
		s.ChannelMessageSend(m.ChannelID, "Receiving help will only weaken you")
	}
	if strings.HasPrefix(m.Content, "!join") {
		err := listenChannel(s, m)

		if err != nil {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Cannot join channel: %v", err))
		}
	}

}
func listenChannel(s *discordgo.Session, m *discordgo.MessageCreate) error {
	c, err := s.State.Channel(m.ChannelID)
	if err != nil {
		// Could not find channel.
		return nil
	}
	g, err := s.State.Guild(c.GuildID)
	if err != nil {
		// Could not find guild.
		return nil
	}
	for _, vs := range g.VoiceStates {
		if vs.UserID == m.Author.ID {
			log.Printf("joined channel: %v", vs.ChannelID)
			vc, err := s.ChannelVoiceJoin(g.ID, vs.ChannelID, false, false)
			vc.AddHandler(listenVoice)
			return err
		}
	}

	return nil
}
func listenVoice(vc *discordgo.VoiceConnection, vs *discordgo.VoiceSpeakingUpdate) {
	if vc.UserID == vs.UserID {
		return
	}
	log.Print("Listen voice started")

	allVoice := make([]discordgo.Packet, 0)
LISTENING:
	for vs.Speaking == true {
		select {
		case voicePacket := <-vc.OpusRecv:
			allVoice = append(allVoice, *voicePacket)
			log.Printf("timestamp: %v, user: %v,event: %v , stored packets: %v",
				voicePacket.Timestamp,
				vs.UserID,
				vs.SSRC,
				len(allVoice))
		case <-time.After(500 * time.Millisecond):
			log.Println("timed out")
			break LISTENING
		}
		//packetCopy := make([]byte,len(voicePacket.Opus))
		//copy(packetCopy,voicePacket.Opus)

	}
	log.Printf("Speaking event ended, stored voice :%v", len(allVoice))
	time.Sleep(2 * time.Second)
	log.Printf("Done sleeping")
	if vc.Ready {
		if len(allVoice) > 5 {
			err := vc.Speaking(true)
			if err == nil {

				for _, voicePacket := range allVoice {
					log.Printf("Sending packet %v", voicePacket.Sequence)
					vc.OpusSend <- voicePacket.Opus
				}
				time.Sleep(500 * time.Millisecond)
				vc.Speaking(false)

			} else {
				log.Printf("Cannot start speaking %v", err)
			}
			log.Print("Done sending")
		}
	} else {
		log.Print("Sent nothing")
	}

}
