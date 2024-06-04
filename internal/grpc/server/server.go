package server

import (
	"context"
	"errors"
	"net/url"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/keycubasic/shortener/internal/dto"
	pb "github.com/keycubasic/shortener/internal/grpc/proto"
	"github.com/keycubasic/shortener/internal/handler"
	"github.com/keycubasic/shortener/internal/storage"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/keycubasic/shortener/internal/config"
)

// GrpcServer stores unimplemented server, configuration and functions of the application.
type GrpcServer struct {
	pb.UnimplementedShortenerServer
	cfg *config.Cfg
	app handler.Application
}

// newGrpcServer creates a new GrpcServer object.
func newGrpcServer(cfg *config.Cfg, app handler.Application) *GrpcServer {
	return &GrpcServer{cfg: cfg, app: app}
}

// PingDB checks the availability of the data store.
func (s *GrpcServer) PingDB(ctx context.Context, _ *pb.PingRequest) (*pb.PingResponse, error) {
	if err := s.app.Ping(ctx); err != nil {
		return nil, err
	}
	return &pb.PingResponse{}, nil
}

// GetOriginalURL returns the original link by its shortening.
func (s *GrpcServer) GetOriginalURL(ctx context.Context, r *pb.GetLinkRequest) (*pb.GetLinkResponse, error) {
	var response pb.GetLinkResponse
	id := r.Id
	link, err := s.app.GetLink(ctx, id)
	if err != nil {
		return nil, status.Error(codes.NotFound, `target link not found`)
	}

	response.Link = link
	return &response, nil
}

// GetUserURLs returns user URLs.
func (s *GrpcServer) GetUserURLs(ctx context.Context, r *pb.GetUserURLsRequest) (*pb.GetUserURLsResponse, error) {
	res := &pb.GetUserURLsResponse{Links: make([]*pb.LinkDto, 0)}
	userID := int(r.UserId)
	linksDto, err := s.app.GetUserURLs(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if len(linksDto) == 0 {
		return nil, status.Error(codes.NotFound, codes.NotFound.String())
	}

	for _, linkDto := range linksDto {
		res.Links = append(res.Links, &pb.LinkDto{OriginalUrl: linkDto.OriginalURL, ShortUrl: linkDto.ShortURL})
	}

	return res, nil
}

// ShortenURL сreates a shortening for a link and returns it.
func (s *GrpcServer) ShortenURL(ctx context.Context, r *pb.CreateShortLinkRequest) (*pb.CreateShortLinkResponse, error) {
	res := &pb.CreateShortLinkResponse{}
	userID := int(r.UserId)
	link := r.Link

	if _, err := url.ParseRequestURI(link); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	key, err := s.app.CreateShortLink(ctx, userID, link)

	if err != nil && !errors.Is(err, storage.ErrUniqueViolation) {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if errors.Is(err, storage.ErrUniqueViolation) {
		key, err = s.app.GetShortLink(ctx, link)
		if err != nil {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}
	}

	res.ShortLink = strings.Join([]string{s.cfg.BaseURL, key}, "/")
	return res, nil
}

// ShortenURLBatch сreates a shortenings for a links in batch mode and returns it in JSON format.
func (s *GrpcServer) ShortenURLBatch(ctx context.Context, r *pb.CreateShortLinksBatchRequest) (*pb.CreateShortLinksBatchResponse, error) {
	res := &pb.CreateShortLinksBatchResponse{}
	userID := int(r.UserId)
	linkDto := make(dto.LinksDto, 0)
	for _, links := range res.Links {
		linkDto = append(linkDto, dto.LinkDto{CorrelationID: links.CorrelationId, OriginalURL: links.OriginalUrl})
	}

	resLinksDto, err := s.app.CreateShortLinksBatch(ctx, userID, linkDto)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	for _, linkDto := range resLinksDto {
		res.Links = append(res.Links, &pb.LinkDto{OriginalUrl: linkDto.OriginalURL, ShortUrl: linkDto.ShortURL})
	}

	return res, nil
}

// DeleteUserURLs deletes link-shortening pairs.
func (s *GrpcServer) DeleteUserURLs(ctx context.Context, r *pb.AppendChanRequest) (*pb.AppendChanResponse, error) {
	userID := int(r.UserId)
	slice := r.Links
	s.app.AppendChan(ctx, userID, slice)
	return &pb.AppendChanResponse{}, nil
}

// GetStats returns service statistics.
func (s *GrpcServer) GetStats(ctx context.Context, r *pb.GetStatsRequest) (*pb.GetStatsResponse, error) {
	stats, err := s.app.GetStats(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.GetStatsResponse{Urls: int32(stats.URLs), Users: int32(stats.Users)}, nil
}

// RegisterGrpcServer registers the gRPC server.
func RegisterGrpcServer(ctx context.Context, cfg *config.Cfg, app handler.Application) *grpc.Server {
	gs := newGrpcServer(cfg, app)
	s := grpc.NewServer()
	pb.RegisterShortenerServer(s, gs)
	log.Info("gRPC server registered")
	return s
}
