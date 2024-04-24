package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
)

var CLASSIFIER_URL string
var CLASSIFIER_KEY string

func main() {
	classifierUrl, isClassifierUrlPresent := os.LookupEnv("CLASSIFIER_URL")
	if !isClassifierUrlPresent {
		log.Fatal("CLASSIFIER_URL is not set")
	}
	CLASSIFIER_URL = classifierUrl

	classifierKey := os.Getenv("CLASSIFIER_KEY")
	CLASSIFIER_KEY = classifierKey

	app := pocketbase.New()

	configure(app)
	hooks(app)

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}

func configure(app *pocketbase.PocketBase) {
	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		Automigrate: false,
	})
}

func hooks(app *pocketbase.PocketBase) {
	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		go func() {
			tMinimumWait := 2 * time.Second
			for {
				tStart := time.Now()

				latestReading := models.Record{}

				err := app.Dao().
					RecordQuery("readings").
					OrderBy("created ASC").
					One(&latestReading)
				if err != nil {
					time.Sleep(tMinimumWait)
					continue
				}

				dataDir := app.DataDir()
				recordFilePath := latestReading.BaseFilesPath()
				fileName := latestReading.Get("content")

				fileBytes, err := os.ReadFile(fmt.Sprintf("%s/storage/%s/%s", dataDir, recordFilePath, fileName))
				if err != nil {
					app.Logger().Error("Unable to read audio file", "filename", fileName)
					time.Sleep(tMinimumWait)
					continue
				}

				result, err := classify(&http.Client{}, CLASSIFIER_URL+"/classify", fileBytes)
				if err != nil {
					app.Logger().Error("Unable to classify audio file", "filename", fileName, "error", err)
					time.Sleep(tMinimumWait)
					continue
				}

				app.Logger().Debug("classification of audio file", "results", result)
				app.Dao().DeleteRecord(&latestReading)

				tEnd := time.Now()
				time.Sleep(min(1*time.Second, tMinimumWait-tEnd.Sub(tStart)))
			}
		}()

		return nil
	})

	app.OnRecordBeforeCreateRequest("readings").Add(func(e *core.RecordCreateEvent) error {
		maxBuffer := getMaxBuffer(app)
		targetLocationId := e.Record.GetString("location")
		readings := []models.Record{}
		err := app.Dao().
			RecordQuery("readings").
			Where(dbx.HashExp{"location": targetLocationId}).
			OrderBy("created DESC").
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

func classify(client *http.Client, endpoint string, fileBytes []byte) ([]interface{}, error) {
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(fileBytes))
	if err != nil {
		return nil, err
	}

	// Set headers
	req.Header.Set("Content-Type", "audio/ogg")
	req.Header.Set("Authorization", "Bearer "+CLASSIFIER_KEY)

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Check if the response status code is not 200 OK
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("response failed: %s", resp.Status)
	}

	// Parse the JSON response
	var result []interface{} // Change interface{} to your specific type
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, fmt.Errorf("error parsing JSON: %s", err.Error())
	}

	return result, nil
}
