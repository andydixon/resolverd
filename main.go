package main

import (
	"regexp"

	"github.com/miekg/dns"
)

func main() {

	dnsCache := InitCache(300000000000)

	dnsProxy := DNSProxy{
		Cache:         &dnsCache,
	}

	logger := NewLogger("info")
	host := "0.0.0.0:61053"
	
	dns.HandleFunc(".", func(w dns.ResponseWriter, r *dns.Msg) {
		switch r.Opcode {
		case dns.OpcodeQuery:
			m, err := dnsProxy.getResponse(r)
			if err != nil {
				m.SetReply(r)
				w.WriteMsg(m)
				return
			}
			if len(m.Answer) > 0 {
				pattern := regexp.MustCompile(`(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3}`)
				ipAddress := pattern.FindAllString(m.Answer[0].String(), -1)

				if len(ipAddress) > 0 {
					logger.Infof("Lookup for %s with ip %s\n", m.Answer[0].Header().Name, ipAddress[0])
				} else {
					logger.Infof("Lookup for %s with response %s\n", m.Answer[0].Header().Name, m.Answer[0])
				}
			}
			m.SetReply(r)
			w.WriteMsg(m)
		}
	})

	server := &dns.Server{Addr: host, Net: "udp"}
logger.Infof("                      _               _ ")
logger.Infof("  _ __ ___  ___  ___ | |_   ___ __ __| |")
logger.Infof(" | '__/ _ \\/ __|/ _ \\| \\ \\ / / '__/ _` |")
logger.Infof(" | | |  __/\\__ \\ (_) | |\\ V /| | | (_| |")
logger.Infof(" |_|  \\___||___/\\___/|_| \\_/ |_|  \\__,_|")
logger.Infof(" Version: 2.0   (c) 2020-2021 Andy Dixon")
	logger.Infof("Starting up and binding to %s\n", host)
	err := server.ListenAndServe()
	if err != nil {
		logger.Errorf("Failed to start server: %s\n ", err.Error())
	}
}
