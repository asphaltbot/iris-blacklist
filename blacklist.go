package blacklist

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/kataras/iris/v12"
)

// Options is the struct that gets passed in by the user to create the blacklist struct with
type Options struct {
	Debug             bool
	BlockedResponse   []byte
	BlockedIPs        []string
	BlockedUserAgents []string
}

// Blacklist is the struct that we will use for all internal purposes
type Blacklist struct {
	log               *log.Logger
	blockedIPs        []string
	blockedUserAgents []string
	blockedResponse   []byte
}

// New creates a new instance of the Blacklist middleware
func New(options Options) iris.Handler {
	b := &Blacklist{
		blockedIPs:        options.BlockedIPs,
		blockedUserAgents: options.BlockedUserAgents,
		blockedResponse:   options.BlockedResponse,
	}

	if options.Debug {
		b.log = log.New(os.Stdout, "[blacklist] ", log.LstdFlags)
	}

	return b.Serve

}

// Serve performs some checks whether the user is blocked or not
func (b *Blacklist) Serve(ctx iris.Context) {
	userAgent := ctx.Request().UserAgent()
	b.log.Print(fmt.Sprintf("checking whether the user agent %s is blocked or not", userAgent))

	for _, v := range b.blockedUserAgents {
		if v == userAgent {
			b.log.Print(fmt.Sprintf("the user's user agent has been blocked. showing blocked page"))
			ctx.StatusCode(http.StatusForbidden)
			ctx.HTML(string(b.blockedResponse))

			ctx.StopExecution()
			return
		}
	}

	userIP := ctx.Request().RemoteAddr
	b.log.Print(fmt.Sprintf("checking whether the IP %s is blocked or not", userIP))

	for _, v := range b.blockedIPs {
		if strings.Contains(userIP, v) {
			b.log.Print(fmt.Sprintf("the user's IP has been blocked. showing blocked page"))
			ctx.StatusCode(http.StatusForbidden)
			ctx.HTML(string(b.blockedResponse))

			ctx.StopExecution()
			return
		}
	}

	// everything seems to check out.
	ctx.Next()

}
