package plugin

import (
	"context"

	pb "github.com/alvarotorresc/cortex/internal/plugin/proto"
)

// GRPCClient is the host-side client that talks to a plugin over gRPC.
// It implements the CortexPlugin interface by translating calls to gRPC.
type GRPCClient struct {
	client pb.CortexPluginClient
}

func (c *GRPCClient) GetManifest() (*Manifest, error) {
	response, err := c.client.GetManifest(context.Background(), &pb.Empty{})
	if err != nil {
		return nil, err
	}

	return &Manifest{
		ID:          response.Id,
		Name:        response.Name,
		Version:     response.Version,
		Description: response.Description,
		Icon:        response.Icon,
		Color:       response.Color,
		Permissions: response.Permissions,
	}, nil
}

func (c *GRPCClient) HandleAPI(request *APIRequest) (*APIResponse, error) {
	response, err := c.client.HandleAPI(context.Background(), &pb.APIRequest{
		Method: request.Method,
		Path:   request.Path,
		Body:   request.Body,
		Query:  request.Query,
	})
	if err != nil {
		return nil, err
	}

	return &APIResponse{
		StatusCode:  int(response.StatusCode),
		Body:        response.Body,
		ContentType: response.ContentType,
	}, nil
}

func (c *GRPCClient) GetWidgetData(slot string) ([]byte, error) {
	response, err := c.client.GetWidgetData(context.Background(), &pb.WidgetRequest{Slot: slot})
	if err != nil {
		return nil, err
	}

	return response.JsonData, nil
}

func (c *GRPCClient) Migrate(databasePath string) error {
	_, err := c.client.Migrate(context.Background(), &pb.MigrateRequest{DbPath: databasePath})
	return err
}

func (c *GRPCClient) Teardown() error {
	_, err := c.client.Teardown(context.Background(), &pb.Empty{})
	return err
}
