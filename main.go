package main

import (
	"fmt"
	"github.com/miekg/dns"
	"log"
	"time"
)

func resolver(domain string, qtype uint16) ([]dns.RR, *dns.Msg) {
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(domain), qtype)
	m.RecursionDesired = true

	c := &dns.Client{Timeout: 5 * time.Second}

	response, _, err := c.Exchange(m, "8.8.8.8:53")

	if err != nil {
		log.Fatalf("[ERROR] : %v\n", err)
		return nil, nil
	}

	if response == nil {
		log.Fatalf("[ERROR] : no response from server\n")
		return nil, nil
	}

	//for _, answer := range response.Answer {
	//	fmt.Printf("%s\n", answer.String())
	//}

	return response.Answer, response
}

type dnsHandler struct{}

func (h *dnsHandler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	msg := new(dns.Msg)
	msg.SetReply(r)
	msg.Authoritative = true

	for _, question := range r.Question {
		answers, resp := resolver(question.Name, question.Qtype)
		msg.Answer = append(msg.Answer, answers...)
		if resp.Ns != nil {
			msg.Ns = append(msg.Ns, resp.Ns...)
		}
	}

	w.WriteMsg(msg)
}

func StartDNSServer() {
	handler := new(dnsHandler)

	// UDP-сервер
	udpServer := &dns.Server{
		Addr:      ":5303",
		Net:       "udp",
		Handler:   handler,
		UDPSize:   65535,
		ReusePort: true,
	}

	// TCP-сервер
	tcpServer := &dns.Server{
		Addr:      ":5303",
		Net:       "tcp",
		Handler:   handler,
		ReusePort: true,
	}

	fmt.Println("Starting DNS server on port 5303 for both UDP and TCP")

	go func() {
		if err := udpServer.ListenAndServe(); err != nil {
			log.Fatalf("Failed to start UDP server: %s\n", err.Error())
		}
	}()

	if err := tcpServer.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start TCP server: %s\n", err.Error())
	}
}

func main() {
	StartDNSServer()
}
