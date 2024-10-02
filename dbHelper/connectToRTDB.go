package dbHelper

import (
	"context"
	"log"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/db"
	"google.golang.org/api/option"
)

func InitializeApp (ctx context.Context, path string, url string) *firebase.App {
	opt := option.WithCredentialsFile(path)
	
	conf := &firebase.Config{
		DatabaseURL: url,
	}

	app, err := firebase.NewApp(ctx, conf, opt)

	if err != nil {
		log.Fatalln("Error Initializing app: ", err)
	}

	return app
}

func ConnectToRTDB (app *firebase.App, ctx context.Context) *db.Client {
	client, err := app.Database(ctx)

	if err != nil {
		log.Fatalln("Error Initializing DB:", err)
	}

	return client
}