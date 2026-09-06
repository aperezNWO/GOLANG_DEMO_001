package grpcservice

import (
	"context"
	"time"

	"go-ping-api/pkg/fractals"
	pb "go-ping-api/pkg/grpc/pb/proto"
)

// Limit to 2 concurrent fractal calculations to protect instance CPU limits
var sem = make(chan struct{}, 2)

type Server struct {
	pb.UnimplementedFractalServiceServer
	engine *fractals.Engine
}

func NewServer(engine *fractals.Engine) *Server {
	return &Server{engine: engine}
}

func (s *Server) GetFractal(ctx context.Context, req *pb.FractalRequest) (*pb.FractalResponse, error) {
	// Enforce a 10-second request timeout[cite: 5]
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Acquire a semaphore slot to regulate concurrent heavy calculations
	select {
	case sem <- struct{}{}:
		defer func() { <-sem }()
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	kind, err := fractals.ParseFractalKind(int(req.Kind))
	if err != nil {
		kind = fractals.Mandelbrot
	}

	maxIter := int(req.MaxIterations)
	if maxIter <= 0 {
		maxIter = 500
	}

	var bounds fractals.Bounds

	if req.XMin == 0 && req.XMax == 0 && req.YMin == 0 && req.YMax == 0 {
		bounds = fractals.Bounds{XMin: -2.0, XMax: 1.0, YMin: -1.2, YMax: 1.2}
	} else {
		bounds = fractals.Bounds{
			XMin: req.XMin, XMax: req.XMax,
			YMin: req.YMin, YMax: req.YMax,
		}
	}

	if bounds.XMin >= bounds.XMax || bounds.YMin >= bounds.YMax {
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
