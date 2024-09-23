package ipdata

import (
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/btschwartz12/site/internal/slack"
	"github.com/ipinfo/go/v2/ipinfo"
)

var (
	ipinfoToken string
)

func init() {
	ipinfoToken = os.Getenv("IPINFO_TOKEN")
}

func GetIpinfoRecord(ip net.IP) *ipinfo.Core {
	client := ipinfo.NewClient(nil, nil, ipinfoToken)
	info, err := client.GetIPInfo(ip)
	if err != nil {
		return nil
	}
	return info
}

func GetIp(r *http.Request) net.IP {
	ip := r.Header.Get("X-Real-Ip")
	if ip == "" {
		ip = r.Header.Get("X-Forwarded-For")
	}
	if ip == "" {
		ip = r.RemoteAddr
	}
	return net.ParseIP(ip)
}

func GetVisitBlocks(r *http.Request, ip net.IP, info *ipinfo.Core) []slack.Block {
	blocks := []slack.Block{
		{
			Type: "header",
			Text: &slack.Element{
				Type:  "plain_text",
				Text:  fmt.Sprintf("visit from %s", "❓"),
				Emoji: true,
			},
		},
		{
			Type: "context",
			Elements: []slack.Element{
				{
					Type: "mrkdwn",
					Text: fmt.Sprintf("path: `%s`", r.URL.Path),
				},
			},
		},
	}

	if ip == nil || info == nil {
		ip = GetIp(r)
		info = GetIpinfoRecord(ip)
	}

	if info != nil {
		blocks[0].Text.Text = fmt.Sprintf("visit from %s", info.CountryFlag.Emoji)
		blocks[1].Elements = append(blocks[1].Elements,
			slack.Element{
				Type: "mrkdwn",
				Text: fmt.Sprintf("IP: %s", ip),
			},
			slack.Element{
				Type: "mrkdwn",
				Text: fmt.Sprintf("%s, %s, %s", info.City, info.Region, info.CountryName),
			},
		)
	}

	return blocks
}
