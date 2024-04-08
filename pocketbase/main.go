package main

import (
	"log"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/models"
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
	app.OnRecordBeforeCreateRequest("readings").Add(func(e *core.RecordCreateEvent) error {
		maxBuffer := getMaxBuffer(app)
		targetLocationId := e.Record.GetString("location")
		readings := []models.Record{}
		err := app.Dao().
			RecordQuery("readings").
			Where(dbx.HashExp{"location": targetLocationId}).
			OrderBy("created ASC").
			All(&readings)
		if err != nil {
			app.Logger().Error(err.Error())
			return err
		}
		numberOfReadings := len(readings)
		if numberOfReadings == 0 || numberOfReadings < maxBuffer {
			return nil
		}

		earliestReading := readings[0]
		err = app.Dao().DeleteRecord(&earliestReading)
		if err != nil {
			app.Logger().Error(err.Error())
			return err
		}

		return nil
	})
}

func getMaxBuffer(app *pocketbase.PocketBase) int {
	record, err := app.Dao().FindRecordsByExpr("configs", dbx.HashExp{"name": "max_buffer"})
	if err != nil || len(record) == 0 {
		return 20
	}
	return record[0].GetInt("value")
}
