package main

import (
	"fmt"
	"net"
	"regexp"
	"math/rand"
	"github.com/miekg/dns"
)

type DNSProxy struct {
	Cache         *Cache
}

func getRandomResolver() string {
	resolvers := [5]string{"8.8.8.8:53", "208.67.222.222:53", "208.67.220.220:53","8.8.4.4:53","1.1.1.1:53"}
	return resolvers[rand.Intn(len(resolvers))]
}

func (proxy *DNSProxy) getResponse(requestMsg *dns.Msg) (*dns.Msg, error) {
	responseMsg := new(dns.Msg)
	if len(requestMsg.Question) > 0 {
		question := requestMsg.Question[0]
		dnsServer := getRandomResolver();
		switch question.Qtype {
		case dns.TypeA:
			answer, err := proxy.processTypeA(dnsServer, &question, requestMsg)
			if err != nil {
				return responseMsg, err
			}
			responseMsg.Answer = append(responseMsg.Answer, *answer)

		default:
			answer, err := proxy.processOtherTypes(dnsServer, &question, requestMsg)
			if err != nil {
				return responseMsg, err
			}
			responseMsg.Answer = append(responseMsg.Answer, *answer)
		}
	}
	return responseMsg, nil
}

func (proxy *DNSProxy) processOtherTypes(dnsServer string, q *dns.Question, requestMsg *dns.Msg) (*dns.RR, error) {

if checkAgainstBlacklist(q.Name) {
	return nil,fmt.Errorf("Blacklisted Host")
}

	queryMsg := new(dns.Msg)
	requestMsg.CopyTo(queryMsg)
	queryMsg.Question = []dns.Question{*q}

	msg, err := lookup(dnsServer, queryMsg)
	if err != nil {
		return nil, err
	}

	if len(msg.Answer) > 0 {
		return &msg.Answer[0], nil
	}
	return nil, fmt.Errorf("not found")
}

func (proxy *DNSProxy) processTypeA(dnsServer string, q *dns.Question, requestMsg *dns.Msg) (*dns.RR, error) {
	cacheMsg, found := proxy.Cache.Get(q.Name)
	ip := ""
if checkAgainstBlacklist(q.Name) {
	return nil,fmt.Errorf("Blacklisted Host")
}
	if !found {
		queryMsg := new(dns.Msg)
		requestMsg.CopyTo(queryMsg)
		queryMsg.Question = []dns.Question{*q}

		msg, err := lookup(dnsServer, queryMsg)
		if err != nil {
			return nil, err
		}

		if len(msg.Answer) > 0 {
			proxy.Cache.Set(q.Name, &msg.Answer[len(msg.Answer)-1])
			return &msg.Answer[len(msg.Answer)-1], nil
		}

	} else if found {
		return cacheMsg.(*dns.RR), nil
	} else if ip != "" {

		answer, err := dns.NewRR(fmt.Sprintf("%s A %s", q.Name, ip))
		if err != nil {
			return nil, err
		}
		return &answer, nil
	}
	return nil, fmt.Errorf("not found")
}

func (dnsProxy *DNSProxy) getIPFromConfigs(domain string, configs map[string]interface{}) string {

	for k, v := range configs {
		match, _ := regexp.MatchString(k+"\\.", domain)
		if match {
			return v.(string)
		}
	}
	return ""
}

func GetOutboundIP() (net.IP, error) {

	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP, nil
}

func lookup(server string, m *dns.Msg) (*dns.Msg, error) {
	dnsClient := new(dns.Client)
	dnsClient.Net = "udp"
	response, _, err := dnsClient.Exchange(m, server)
	if err != nil {
		return nil, err
	}

	return response, nil
}
