package example

import (
	blacklist "github.com/asphaltbot/iris-blacklist"
	"github.com/kataras/iris/v12"
)

func main() {
	app := iris.New()

	replaceValues := make(map[string]string, 1)
	replaceValues["message"] = "This is a test"

	// we dont specify BlockedResponse since blacklist.New will automatically download the template file
	// and use that if we dont specify one
	// you can always specify one yourself, it's just a byte array
	// by default, ReplaceStrings will have {{ip}} replaced with the user's IP address (if it exists in the specified template)
	// so you don't need to add that
	blacklistMiddleware := blacklist.New(blacklist.Options{
		BlockedIPs:        []string{"127.0.0.1", "::1"},
		BlockedUserAgents: []string{},
		ReplaceStrings:    replaceValues,
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
