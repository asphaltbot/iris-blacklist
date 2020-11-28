package blacklist

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/kataras/iris/v12"
)

// Options is the struct that gets passed in by the user to create the blacklist struct with
// If BlockedResponse is null, we will download and use the created template (see examples/template.html)
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

	// if the user has not specified a blocked response, then download the template and use that
	if b.blockedResponse == nil {
		b.log.Println("no blocked response specified, downloading template file")

		fileBytes, err := b.downloadTemplateFile("https://asphaltbot.com/middleware/blacklist/template.html")

		if err != nil {
			panic("[blacklist] unable to download file: " + err.Error())
		}

		b.log.Println(fmt.Sprintf("successfully downloaded template file"))
		b.blockedResponse = fileBytes
	}

	return b.Serve

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
