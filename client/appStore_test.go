package client

import (
	"testing"
)

func TestGetAppStoreListing(t *testing.T) {
	mockResponse := map[string]interface{}{
		"app": map[string]interface{}{
			"name":          "test title",
			"description":   "test description",
			"authorDisplay": "Cool Author",
			"url":           "some_url",
			"image":         "some_image_url",
		},
	}
	mockClient := MockClient{
		response: &mockResponse,
	}
	client := AppStoreClient{
		client:     &mockClient,
		graphqlUrl: "marketplace-service:deployed/v1/marketplace/authenticated/graphql",
	}
	response, err := client.GetAppStoreListing("some_module_id")
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
