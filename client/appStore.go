package client

import (
	"errors"

	"github.com/mitchellh/mapstructure"
)

const GET_APP_STORE_LISTING = `
  query GetAppStoreListing($id: ID!) {
    app(id: $id) {
      name
      description
      authorDisplay
      image
      ... on AppStoreWebApplication {
        url
      }
    }
  }
`

const DELETE_APP_STORE_LISTING = `
  mutation DeleteAppStoreListing($id: ID!) {
	deleteApp(id: $id)
  }
`

const CREATE_APP_STORE_LISTING = `
  mutation CreateAppStoreListing($input: CreateWebAppInput!) {
    createWebApp(input: $input) {
      id
    }
  }
`

const EDIT_APP_STORE_LISTING = `
  mutation EditAppStoreListing($id: ID!, $edits: EditWebAppInput!) {
    editWebApp(id: $id, edits: $edits) 
  }
`

const GRAPHQL_URL = "/graphql"

type AppStoreClient struct {
	graphqlUrl string
	client     graphqlClient
}

type app struct {
	Name          string
	Description   string
	AuthorDisplay string
	Image         string
	Url           string
}

func (self *AppStoreClient) Gql(query string, variables map[string]interface{}) (*map[string]interface{}, error) {
	return self.client.Gql(self.graphqlUrl, query, variables)
}

func (self *AppStoreClient) GetAppStoreListing(id string) (*app, error) {
	res, err := self.Gql(GET_APP_STORE_LISTING, map[string]interface{}{"id": id})
	if err != nil {
		return nil, err
	}

	var data struct {
		App app
	}
	err = mapstructure.Decode(res, &data)
	if err != nil {
		return nil, err
	}
	return &data.App, nil
}

type appStoreCreate struct {
	Name          string
	AuthorDisplay string
	Url           string
	Description   string
	Image         string
}

func (self *AppStoreClient) CreateAppStoreListing(params appStoreCreate) (*string, error) {
	res, err := self.Gql(CREATE_APP_STORE_LISTING, map[string]interface{}{"input": map[string]string{
		"name":          params.Name,
		"authorDisplay": params.AuthorDisplay,
		"url":           params.Url,
		"description":   params.Description,
		"image":         params.Image,
		"product":       "LX",
	}})
	if err != nil {
		return nil, err
	}
	var data struct {
		CreateWebApp struct {
			Id string
		}
	}
	err = mapstructure.Decode(res, &data)
	if err != nil {
		return nil, err
	}
	return &data.CreateWebApp.Id, nil
}

func (self *AppStoreClient) EditAppStoreListing(id string, params appStoreCreate) error {
	res, err := self.Gql(EDIT_APP_STORE_LISTING, map[string]interface{}{
		"id": id,
		"edits": map[string]string{
			"name":          params.Name,
			"authorDisplay": params.AuthorDisplay,
			"url":           params.Url,
			"description":   params.Description,
			"image":         params.Image,
		}})
	if err != nil {
		return err
	}

	var data struct {
		EditWebApp bool
	}
	err = mapstructure.Decode(res, &data)
	if err != nil {
		return err
	}
	if !data.EditWebApp {
		return errors.New("The app you're trying to edit does not exist")
	}
	return nil
}

func (self *AppStoreClient) DeleteAppStoreListing(id string) error {
	res, err := self.Gql(DELETE_APP_STORE_LISTING, map[string]interface{}{
		"id": id,
	})
	if err != nil {
		return err
	}

	var data struct {
		DeleteApp bool
	}
	err = mapstructure.Decode(res, &data)
	if err != nil {
		return err
	}
	if !data.DeleteApp {
		return errors.New("The app you're trying to delete does not exist")
	}
	return nil
}
