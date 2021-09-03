package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"

	firebase "firebase.google.com/go"
	"github.com/joho/godotenv"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type MetaStorage struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

func main() {
	err := godotenv.Load("/Users/adax/go/src/AdaBrain/adatools/.env.yaml")

	config := &firebase.Config{
		StorageBucket: "marketplace-testenv.appspot.com",
	}
	opt := option.WithCredentialsFile(os.Getenv("TOKEN_PATH"))
	app, err := firebase.NewApp(context.Background(), config, opt)
	if err != nil {
		log.Printf("error initializing app: %v", err)
	}

	ctx := context.Background()
	client, err := app.Storage(ctx)
	if err != nil {
		log.Fatalln(err)
	}

	bucket, err := client.DefaultBucket()
	if err != nil {
		log.Fatalln(err)
	}

	iter := bucket.Objects(ctx, nil)
	if err != nil {
		log.Fatalln(err)
	}
	// URL Formatting: firebaseURL + bucketURL + /o/ + nameURL + ?alt=media&token= + Token
	firebaseURL := "https://firebasestorage.googleapis.com/v0/b/"
	store := []MetaStorage{}

	for {
		obj, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Bucket(%q).Objects: %v", bucket, err)
		}
		token := obj.Metadata["firebaseStorageDownloadTokens"]
		filename := url.PathEscape(obj.Name)
		mediaURL := fmt.Sprintf("%s%s/o/%s?alt=media&token=%s", firebaseURL, obj.Bucket, filename, token)
		meta := &MetaStorage{
			Name: obj.Name,
			Url:  mediaURL,
		}

		store = append(store, *meta)
	}

	bf := bytes.NewBuffer([]byte{})
	jsonEncoder := json.NewEncoder(bf)
	jsonEncoder.SetEscapeHTML(false)
	jsonEncoder.SetIndent("", " ")
	jsonEncoder.Encode(store)

	if err := ioutil.WriteFile("mediaURLs.json", []byte(bf.String()), 0644); err != nil {
		log.Fatalln(err)
	}

	log.Println("========== Done Processing ==========")
}
