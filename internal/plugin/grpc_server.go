package plugin

import (
	"context"

	goplugin "github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"

	pb "github.com/alvarotorresc/cortex/internal/plugin/proto"
)

// CortexGRPCPlugin implements go-plugin's GRPCPlugin interface.
// It bridges the HashiCorp go-plugin system with the gRPC transport layer.
type CortexGRPCPlugin struct {
	goplugin.Plugin
	Impl CortexPlugin
}

func (p *CortexGRPCPlugin) GRPCServer(broker *goplugin.GRPCBroker, server *grpc.Server) error {
	pb.RegisterCortexPluginServer(server, &grpcServer{impl: p.Impl})
	return nil
}

func (p *CortexGRPCPlugin) GRPCClient(ctx context.Context, broker *goplugin.GRPCBroker, connection *grpc.ClientConn) (interface{}, error) {
	return &GRPCClient{client: pb.NewCortexPluginClient(connection)}, nil
}

// grpcServer wraps a CortexPlugin implementation to serve over gRPC (plugin side).
type grpcServer struct {
	pb.UnimplementedCortexPluginServer
	impl CortexPlugin
}

func (s *grpcServer) GetManifest(ctx context.Context, _ *pb.Empty) (*pb.PluginManifest, error) {
	manifest, err := s.impl.GetManifest()
	if err != nil {
		return nil, err
	}

	return &pb.PluginManifest{
		Id:          manifest.ID,
		Name:        manifest.Name,
		Version:     manifest.Version,
		Description: manifest.Description,
		Icon:        manifest.Icon,
		Color:       manifest.Color,
		Permissions: manifest.Permissions,
	}, nil
}

func (s *grpcServer) HandleAPI(ctx context.Context, request *pb.APIRequest) (*pb.APIResponse, error) {
	response, err := s.impl.HandleAPI(&APIRequest{
		Method: request.Method,
		Path:   request.Path,
		Body:   request.Body,
		Query:  request.Query,
	})
	if err != nil {
		return nil, err
	}

	return &pb.APIResponse{
		StatusCode:  int32(response.StatusCode),
		Body:        response.Body,
		ContentType: response.ContentType,
	}, nil
}

func (s *grpcServer) GetWidgetData(ctx context.Context, request *pb.WidgetRequest) (*pb.WidgetData, error) {
	data, err := s.impl.GetWidgetData(request.Slot)
	if err != nil {
		return nil, err
	}

	return &pb.WidgetData{JsonData: data}, nil
}

func (s *grpcServer) Migrate(ctx context.Context, request *pb.MigrateRequest) (*pb.MigrateResult, error) {
	err := s.impl.Migrate(request.DbPath)
	if err != nil {
		return &pb.MigrateResult{Success: false, Message: err.Error()}, nil
	}

	return &pb.MigrateResult{Success: true, Message: "ok"}, nil
}

func (s *grpcServer) Teardown(ctx context.Context, _ *pb.Empty) (*pb.Empty, error) {
	return &pb.Empty{}, s.impl.Teardown()
}
