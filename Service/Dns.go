package service

import (
	"fmt"
	Config "godnslog/Config"
	Db "godnslog/Database"
	"log"
	"net"
	"strings"
	"time"

	"golang.org/x/net/dns/dnsmessage"
)

func handle(addr *net.UDPAddr, conn *net.UDPConn, msg dnsmessage.Message) {
	if len(msg.Questions) < 1 {
		return
	}
	question := msg.Questions[0]
	var (
		queryName  = question.Name.String()
		queryType  = question.Type
		newName, _ = dnsmessage.NewName(queryName)
		resource   dnsmessage.Resource
	)
	log.Println(fmt.Sprintf("%s %s -> %s", queryType.String(), queryName, addr.IP.String()))
	if strings.Contains(queryName, Config.Config.Domain) {
		record := Db.DnsRecord{
			Record:   queryName[:len(queryName)-1],
			Type:     strings.TrimPrefix(queryType.String(), "Type"),
			Ip:       addr.IP.String(),
			Datetime: time.Now().Format("2006-01-02"),
		}
		_ = record.Insert()
	}
	switch queryType {
	case dnsmessage.TypeA:
		resource = dnsmessage.Resource{
			Header: dnsmessage.ResourceHeader{
				Name:  newName,
				Class: dnsmessage.ClassINET,
				TTL:   600,
			},
			Body: &dnsmessage.AResource{
				A: [4]byte{127, 0, 0, 1},
			},
		}
	default:
		return
	}
	msg.Response = true
	msg.Answers = append(msg.Answers, resource)
	if packed, err := msg.Pack(); err == nil {
		if _, err = conn.WriteToUDP(packed, addr); err != nil {
			log.Println(err)
		}
	}
}

func DnsServe() {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.ParseIP(Config.Config.Listen),
		Port: 53,
	})
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Close()
	log.Println("DNS service start ...")
	for {
		buf := make([]byte, 512)
		_, addr, _ := conn.ReadFromUDP(buf)
		var msg dnsmessage.Message
		if err := msg.Unpack(buf); err != nil {
			continue
		}
		go handle(addr, conn, msg)
	}
}
