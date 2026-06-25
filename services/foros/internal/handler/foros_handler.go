package handler

import (
	"context"

	forospb "Prueba-Go/gen/foros"
	"Prueba-Go/services/foros/internal/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ForosHandler struct {
	forospb.UnimplementedForosServiceServer
	svc *service.ForosService
}

func NewForosHandler(svc *service.ForosService) *ForosHandler {
	return &ForosHandler{svc: svc}
}

func (h *ForosHandler) ListForoPosts(ctx context.Context, req *forospb.LeccionRequest) (*forospb.ListPostsResponse, error) {
	list, err := h.svc.ListForoPosts(ctx, req.LeccionId, req.UserId)
	if err != nil {
		return nil, toGRPC(err)
	}
	return &forospb.ListPostsResponse{Posts: list}, nil
}

func (h *ForosHandler) CreateForoPost(ctx context.Context, req *forospb.CreatePostRequest) (*forospb.PostResponse, error) {
	p, err := h.svc.CreateForoPost(ctx, req)
	if err != nil {
		return nil, toGRPC(err)
	}
	return p, nil
}

func (h *ForosHandler) DeleteForoPost(ctx context.Context, req *forospb.PostUserRequest) (*forospb.EmptyResponse, error) {
	// El campo user_id vacío == admin (gateway pone rol en metadata; aquí simplificamos por user_id)
	isAdmin := req.UserId == ""
	if err := h.svc.DeleteForoPost(ctx, req.PostId, req.UserId, isAdmin); err != nil {
		return nil, toGRPC(err)
	}
	return &forospb.EmptyResponse{}, nil
}

func (h *ForosHandler) ListForoComentarios(ctx context.Context, req *forospb.PostUserRequest) (*forospb.ListComentariosResponse, error) {
	list, err := h.svc.ListForoComentarios(ctx, req.PostId, req.UserId)
	if err != nil {
		return nil, toGRPC(err)
	}
	return &forospb.ListComentariosResponse{Comentarios: list}, nil
}

func (h *ForosHandler) CreateForoComentario(ctx context.Context, req *forospb.CreateComentarioRequest) (*forospb.ComentarioResponse, error) {
	c, err := h.svc.CreateForoComentario(ctx, req)
	if err != nil {
		return nil, toGRPC(err)
	}
	return c, nil
}

func (h *ForosHandler) ToggleForoPostReaction(ctx context.Context, req *forospb.PostReactionRequest) (*forospb.ReactionResponse, error) {
	r, err := h.svc.ToggleForoPostReaction(ctx, req)
	if err != nil {
		return nil, toGRPC(err)
	}
	return r, nil
}

func (h *ForosHandler) ToggleForoComentarioReaction(ctx context.Context, req *forospb.ComentarioReactionRequest) (*forospb.ReactionResponse, error) {
	r, err := h.svc.ToggleForoComentarioReaction(ctx, req)
	if err != nil {
		return nil, toGRPC(err)
	}
	return r, nil
}

func toGRPC(err error) error {
	return status.Error(codes.Internal, err.Error())
}
