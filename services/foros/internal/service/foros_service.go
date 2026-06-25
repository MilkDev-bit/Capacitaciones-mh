package service

import (
	"context"

	forospb "Prueba-Go/gen/foros"
	"Prueba-Go/services/foros/internal/repository"
)

type ForosService struct {
	repo repository.ForosRepository
}

func NewForosService(repo repository.ForosRepository) *ForosService {
	return &ForosService{repo: repo}
}

func (s *ForosService) ListForoPosts(ctx context.Context, leccionID, userID string) ([]*forospb.PostResponse, error) {
	posts, err := s.repo.ListPosts(ctx, leccionID, userID)
	if err != nil {
		return nil, err
	}
	result := make([]*forospb.PostResponse, 0, len(posts))
	for _, p := range posts {
		result = append(result, p.ToProto())
	}
	return result, nil
}

func (s *ForosService) CreateForoPost(ctx context.Context, req *forospb.CreatePostRequest) (*forospb.PostResponse, error) {
	p, err := s.repo.CreatePost(ctx, req)
	if err != nil {
		return nil, err
	}
	return p.ToProto(), nil
}

func (s *ForosService) DeleteForoPost(ctx context.Context, postID, userID string, isAdmin bool) error {
	return s.repo.DeletePost(ctx, postID, userID, isAdmin)
}

func (s *ForosService) ListForoComentarios(ctx context.Context, postID, userID string) ([]*forospb.ComentarioResponse, error) {
	cs, err := s.repo.ListComentarios(ctx, postID, userID)
	if err != nil {
		return nil, err
	}
	result := make([]*forospb.ComentarioResponse, 0, len(cs))
	for _, c := range cs {
		result = append(result, c.ToProto())
	}
	return result, nil
}

func (s *ForosService) CreateForoComentario(ctx context.Context, req *forospb.CreateComentarioRequest) (*forospb.ComentarioResponse, error) {
	c, err := s.repo.CreateComentario(ctx, req)
	if err != nil {
		return nil, err
	}
	return c.ToProto(), nil
}

func (s *ForosService) ToggleForoPostReaction(ctx context.Context, req *forospb.PostReactionRequest) (*forospb.ReactionResponse, error) {
	return s.repo.TogglePostReaction(ctx, req)
}

func (s *ForosService) ToggleForoComentarioReaction(ctx context.Context, req *forospb.ComentarioReactionRequest) (*forospb.ReactionResponse, error) {
	return s.repo.ToggleComentarioReaction(ctx, req)
}
