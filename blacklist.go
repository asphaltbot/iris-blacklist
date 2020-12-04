package blacklist

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"

	"github.com/kataras/iris/v12"
)

type Options struct {
	BlockedResponse       []byte
	BlockedIPs            []string
	BlockedIpRanges       []string
	BlockedUserAgents     []string
	ReplaceStrings        map[string]string
	BlacklistedStatusCode int
}

// Blacklist is the struct that we will use for all internal purposes
type Blacklist struct {
	blockedIPs        []string
	blockedUserAgents []string
	blockedIpRanges   []string
	blockedResponse   []byte
	statusCode        int
	replaceStrings    map[string]string
}

// New creates a new instance of the Blacklist middleware
func New(options Options) iris.Handler {
	b := &Blacklist{
		blockedIPs:        options.BlockedIPs,
		statusCode:        options.BlacklistedStatusCode,
		blockedIpRanges:   options.BlockedIpRanges,
		blockedUserAgents: options.BlockedUserAgents,
		blockedResponse:   options.BlockedResponse,
		replaceStrings:    options.ReplaceStrings,
	}

	// if there are no replace strings defined in the options struct, make a map so we can pass some default values
	if b.replaceStrings == nil || len(b.replaceStrings) == 0 {
		defaultReplaceStrings := make(map[string]string, 2)
		b.replaceStrings = defaultReplaceStrings
	}

	// get the IP addresses within ranges that have been blocked
	if b.blockedIpRanges != nil && len(b.blockedIpRanges) != 0 {
		for _, v := range b.blockedIpRanges {
			ipsInCidr, err := getIpsInCIDR(v)

			if err != nil {
				fmt.Println("Couldn't get IP addresses in range " + v + ", is it valid?")
				continue
			}

			for _, v := range ipsInCidr {
				b.blockedIPs = append(b.blockedIPs, v)
			}
		}
	}

	// if the user has not specified a blocked response, then download the template and use that
	if b.blockedResponse == nil {
		fileBytes, err := b.downloadTemplateFile("https://asphaltbot.com/middleware/blacklist/template.html")

		if err != nil {
			panic("[blacklist] unable to download file: " + err.Error())
		}

		b.blockedResponse = fileBytes
	}

	return b.Serve

}

func (b *Blacklist) returnWithValuesReplaced() []byte {
	blockedResponse := string(b.blockedResponse)

	for k, v := range b.replaceStrings {
		blockedResponse = strings.Replace(blockedResponse, fmt.Sprintf("{{%s}}", k), v, -1)
	}

	return []byte(blockedResponse)
}

func (b *Blacklist) downloadTemplateFile(url string) ([]byte, error) {
	resp, err := http.Get(url)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	responseBytes, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	return responseBytes, err

}

// Serve performs some checks whether the user is blocked or not
func (b *Blacklist) Serve(ctx iris.Context) {
	userAgent := ctx.Request().UserAgent()

	b.replaceStrings["ip"] = ctx.Request().RemoteAddr

	for _, v := range b.blockedUserAgents {
		if v == userAgent {
			ctx.StatusCode(b.statusCode)
			ctx.HTML(string(b.returnWithValuesReplaced()))

			ctx.StopExecution()
			return
		}
	}

	userIP := ctx.Request().RemoteAddr

	for _, v := range b.blockedIPs {
		if strings.Contains(userIP, v) {
			ctx.StatusCode(b.statusCode)
			ctx.HTML(string(b.returnWithValuesReplaced()))

			ctx.StopExecution()
			return
		}
	}

	// everything seems to check out.
	ctx.Next()

}

func getIpsInCIDR(cidr string) ([]string, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	ips := make([]string, 0)
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		ips = append(ips, ip.String())
	}

	return ips, nil

}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}
