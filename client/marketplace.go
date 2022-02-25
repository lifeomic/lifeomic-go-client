package client

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"

	"github.com/mitchellh/mapstructure"
)

type MarketplaceClient struct {
	graphqlUrl string
	client     graphqlClient
}

func (self *MarketplaceClient) Gql(query string, variables map[string]interface{}) (*map[string]interface{}, error) {
	return self.client.Gql(self.graphqlUrl, query, variables)
}

const GET_PUBLISHED_APP_TILE_MODULE = `
  query GetPublishedModule($id: ID!, $version: String) {
    myModule(moduleId: $id, version: $version) {
      title
      description
	  version
	  source {
		... on AppTile {
		  id
		}
	  }
	  iconV2 {
		url
		fileName
		fileExtension
	  }
    }
  }
`

const CREATE_DRAFT_MODULE = `
 mutation CreateDraftModule($input: CreateDraftModuleInput!) {
   createDraftModule(input: $input) {
     id
   }
 }
`

const SET_APP_TILE = `
  mutation SetAppTile($input: SetPublicAppTileDraftModuleSourceInput!) {
	setPublicAppTileDraftModuleSource(input: $input) {
	  moduleId
	}
  }
`

const PUBLISH_MODULE = `
  mutation PublishModule($input: PublishDraftModuleInputV2!) {
	publishDraftModuleV2(input: $input) {
	  id
	  version {
		version
	  }
	}
  }
`

const START_IMAGE_UPLOAD = `
  mutation StartImageUpload($input: StartUploadInput!) {
	startUpload(input: $input) {
	  id
	  url
	  fields
	}
  }
`

const FINALIZE_IMAGE_UPLOAD = `
  mutation FinalizeImageUpload($input: FinalizeUploadInput!) {
	finalizeUpload(input: $input) {
	  moduleId
	}
  }
`

type appTileModule struct {
	Title       string
	Description string
	Version     string
	Source      struct {
		Id string
	}
	IconV2 *struct {
		Url           string
		FileName      string
		FileExtension string
	}
}

func (self *MarketplaceClient) GetAppTileModule(id string) (*appTileModule, error) {
	res, err := self.Gql(GET_PUBLISHED_APP_TILE_MODULE, map[string]interface{}{"id": id})
	if err != nil {
		return nil, err
	}
	var data struct {
		MyModule *appTileModule
	}
	err = mapstructure.Decode(res, &data)
	if err != nil {
		return nil, err
	}
	return data.MyModule, nil
}

type appTileCreate struct {
	Name           string
	Description    string
	Image          string
	AppTileId      string
	Version        string
	ParentModuleId *string
}

func postImageToUrl(url string, image string, file_name string, fields map[string]string) error {
	file, err := os.Open(image)
	if err != nil {
		return err
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for key, val := range fields {
		err = writer.WriteField(key, val)
		if err != nil {
			return err
		}
	}
	part, err := writer.CreateFormFile("file", file_name)
	if err != nil {
		return err
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return err
	}

	err = writer.Close()
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if err != nil {
		return err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	responseBody := &bytes.Buffer{}
	responseBody.ReadFrom(resp.Body)
	resp.Body.Close()
	return nil
}

func (self *MarketplaceClient) AttachImageToDraftModule(moduleId string, image string) error {
	fileName := path.Base(image)
	startResponse, err := self.Gql(START_IMAGE_UPLOAD, map[string]interface{}{
		"input": map[string]interface{}{
			"fileName": fileName,
		},
	})
	if err != nil {
		return err
	}
	var startData struct {
		StartUpload struct {
			Fields map[string]string
			Url    string
			Id     string
		}
	}
	err = mapstructure.Decode(startResponse, &startData)
	if err != nil {
		return err
	}

	err = postImageToUrl(startData.StartUpload.Url, image, fileName, startData.StartUpload.Fields)
	if err != nil {
		return err
	}

	finalizeResponse, err := self.Gql(FINALIZE_IMAGE_UPLOAD, map[string]interface{}{
		"input": map[string]string{
			"id":       startData.StartUpload.Id,
			"moduleId": moduleId,
			"type":     "ICON",
		},
	})

	if err != nil {
		return nil
	}

	var finalizeData struct {
		FinalizeUpload struct {
			ModuleId string
		}
	}

	err = mapstructure.Decode(finalizeResponse, &finalizeData)
	return err
}

func (self *MarketplaceClient) CreateAppTileDraftModule(params appTileCreate) (*string, error) {
	res, err := self.Gql(CREATE_DRAFT_MODULE, map[string]interface{}{"input": map[string]interface{}{
		"title":       params.Name,
		"description": params.Description,
		// "iconV2":         params.Image, // Use upload
		"parentModuleId": params.ParentModuleId,
		"category":       "APP_TILE",
	}})

	if err != nil {
		return nil, err
	}

	var createDraftData struct {
		CreateDraftModule struct {
			Id string
		}
	}

	err = mapstructure.Decode(res, &createDraftData)
	if err != nil {
		return nil, err
	}

	moduleId := createDraftData.CreateDraftModule.Id

	res, err = self.Gql(SET_APP_TILE, map[string]interface{}{"input": map[string]interface{}{
		"moduleId": moduleId,
		"sourceInfo": map[string]string{
			"id": params.AppTileId,
		},
	}})

	if err != nil {
		return nil, err
	}

	var setAppTileData struct {
		SetPublicAppTileDraftModuleSource struct {
			ModuleId string
		}
	}
	err = mapstructure.Decode(res, &setAppTileData)
	if err != nil {
		return nil, err
	}

	err = self.AttachImageToDraftModule(moduleId, params.Image)

	if err != nil {
		return nil, err
	}

	return &moduleId, nil
}

func (self *MarketplaceClient) PublishNewAppTileModule(params appTileCreate) (*string, error) {
	draftModuleId, err := self.CreateAppTileDraftModule(params)
	if err != nil {
		return nil, err
	}
	publishRes, err := self.Gql(PUBLISH_MODULE, map[string]interface{}{"input": map[string]interface{}{
		"moduleId": draftModuleId,
		"version": map[string]string{
			"version": params.Version,
		},
	}})
	if err != nil {
		return nil, err
	}
	var publishModuleData struct {
		PublishDraftModuleV2 struct {
			Id      string
			Version struct {
				Version string
			}
		}
	}
	err = mapstructure.Decode(publishRes, &publishModuleData)
	if err != nil {
		return nil, err
	}
	return &publishModuleData.PublishDraftModuleV2.Id, nil
}
