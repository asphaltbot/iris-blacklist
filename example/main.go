package example

import (
	"github.com/asphaltbot/blacklist-middleware"
	"github.com/kataras/iris/v12"
)

func main() {
	app := iris.New()

	// we dont specify BlockedResponse since blacklist.New will automatically download the template file
	// and use that if we dont specify one
	// you can always specify one yourself, it's just a byte array
	blacklistMiddleware := blacklist.New(blacklist.Options{
		Debug:             true,
		BlockedIPs:        []string{"127.0.0.1", "::1"},
		BlockedUserAgents: []string{},
	})

	app.Use(blacklistMiddleware)
	app.Get("/", IndexRoute)

	err := app.Run(iris.Addr(":8080"))

	if err != nil {
		panic(err)
	}

}

func IndexRoute(ctx iris.Context) {
	ctx.StatusCode(200)
	ctx.HTML("<h1>Hello world!</h1>")
}
