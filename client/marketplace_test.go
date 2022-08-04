package client

import (
	"testing"
)

func TestGetAppTileModule(t *testing.T) {
	mockResponse := map[string]interface{}{
		"myModule": map[string]interface{}{
			"title":       "test title",
			"description": "test description",
			"version":     "1.0.0",
			"source": map[string]string{
				"id": "some_id",
			},
			"iconV2": map[string]string{
				"url":           "some_url",
				"fileName":      "fancy_file",
				"fileExtension": "png",
			},
		},
	}
	mockClient := MockClient{
		response: &mockResponse,
	}
	client := MarketplaceClient{
		client:     &mockClient,
		graphqlUrl: "marketplace-service:deployed/v1/marketplace/authenticated/graphql",
	}
	response, err := client.GetAppTileModule("some_module_id")
	if err != nil {
		t.Fatal("Unexpected error", err)
	}

	if !mockClient.hasBeenCalled {
		t.Fatal("Mock Client never called")
	}

	if response == nil {
		t.Fatal("Response should not be nil")
	}

	if response.Description != "test description" {
		t.Fatal("Did not get back currect response", response)
	}
}
