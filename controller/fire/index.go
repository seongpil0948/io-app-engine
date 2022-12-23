package fire

import (
	"context"
	"fmt"
	"log"
	"os"

	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"
)

type FireApp struct {
	Ctx  context.Context
	Inst *firebase.App
}

func (f FireApp) ToString() string {
	_id, err := f.Inst.InstanceID(f.Ctx)
	return fmt.Sprintf("FireApp: %v, %v", _id, err)
}

var instance *FireApp

func newApp() *FireApp {
	appInst := new(FireApp)
	appInst.Ctx = context.Background()
	gcpCredentialPath := os.Getenv("GCP_CREDENTIALS_PATH")
	if gcpCredentialPath == "" {
		log.Fatalf("Not Specified env gcpCredentialPath: %v", gcpCredentialPath)
	}
	opt := option.WithCredentialsFile(gcpCredentialPath)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
	}
	appInst.Inst = app
	return appInst
}

func GetFireInstance() *FireApp {
	if instance == nil {
		instance = newApp()
	}
	return instance
}
