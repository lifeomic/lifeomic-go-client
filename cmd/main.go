package main

import (
	"encoding/json"
	"io/ioutil"
	"log"

	"github.com/alexflint/go-arg"
	"github.com/lifeomic/phc-sdk-go/client"
	"github.com/mitchellh/mapstructure"
)

func main() {
	var args struct {
		Query     string `arg:"required"`
		Variables string `arg:"required"`
		Uri       string `arg:"required"`
		User      string `arg:"required"`
	}
	arg.MustParse(&args)

	query, err := ioutil.ReadFile(args.Query)
	if err != nil {
		log.Fatal(err)
	}

	variablesFile, err := ioutil.ReadFile(args.Variables)
	if err != nil {
		log.Fatal(err)
	}

	var variables map[string]interface{}
	err = json.Unmarshal(variablesFile, &variables)
	if err != nil {
		log.Fatal(err)
	}

	phcClient, err := client.BuildClient("lifeomic", args.User, map[string]bool{})
	if err != nil {
		log.Fatal(err)
	}

	resp, err := phcClient.Gql(args.Uri, string(query), variables)
	if err != nil {
		log.Fatal(err)
	}

	var module struct {
		MyModule struct {
			Description string
			Title       string
			Source      struct {
				Id string
			}
		}
	}

	err = mapstructure.Decode(resp, &module)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(module.MyModule)
}
