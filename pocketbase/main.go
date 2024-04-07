package main

import (
	"log"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
)

func main() {
	app := pocketbase.New()

	configure(app)
	routes(app)

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}

func configure(app *pocketbase.PocketBase) {
	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		Automigrate: false,
	})
}

func routes(app *pocketbase.PocketBase) {
	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		e.Router.POST("", func(c echo.Context) error {
			return nil
		})
		return nil
	})
}
