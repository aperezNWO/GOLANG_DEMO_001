package grpcservice

import (
	"context"
	"go-ping-api/pkg/fractals"
	 pb "go-ping-api/pkg/grpc/pb/proto"
)

type Server struct {
	pb.UnimplementedFractalServiceServer
	engine *fractals.Engine
}

func NewServer(engine *fractals.Engine) *Server {
	return &Server{engine: engine}
}

func (s *Server) GetFractal(ctx context.Context, req *pb.FractalRequest) (*pb.FractalResponse, error) {
	kind, err := fractals.ParseFractalKind(int(req.Kind))
	if err != nil {
		kind = fractals.Mandelbrot
	}

	maxIter := int(req.MaxIterations)
	if maxIter <= 0 {
		maxIter = 500
	}

	bounds := fractals.Bounds{
		XMin: req.XMin, XMax: req.XMax,
		YMin: req.YMin, YMax: req.YMax,
	}

	if bounds.XMin == 0 && bounds.XMax == 0 {
		bounds = fractals.Bounds{XMin: -2.0, XMax: 1.0, YMin: -1.2, YMax: 1.2}
	}

	points := s.engine.GetFractal(kind, bounds, maxIter)

	pbPoints := make([]*pb.Point, len(points))
	for i, p := range points {
		pbPoints[i] = &pb.Point{
			X:         p.X,
			Y:         p.Y,
			Intensity: int32(p.Intensity),
		}
	}

	return &pb.FractalResponse{Points: pbPoints}, nil
}