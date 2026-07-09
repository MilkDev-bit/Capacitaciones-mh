package repository

import (
	"context"
	"encoding/json"
	"time"

	forospb "Prueba-Go/gen/foros"

	"github.com/jmoiron/sqlx"
	"google.golang.org/grpc/metadata"
)

type ForoPost struct {
	ID        string    `db:"id"`
	LeccionID string    `db:"leccion_id"`
	UserID    string    `db:"user_id"`
	UserName  string    `db:"user_name"`
	Titulo    string    `db:"titulo"`
	Contenido string    `db:"contenido"`
	MediaURL  string    `db:"media_url"`
	MediaType string    `db:"media_type"`
	Reactions []byte    `db:"reactions"`
	CreatedAt time.Time `db:"created_at"`
}

func (p *ForoPost) ToProto() *forospb.PostResponse {
	return &forospb.PostResponse{
		Id: p.ID, LeccionId: p.LeccionID, UserId: p.UserID, UserName: p.UserName,
		Titulo: p.Titulo, Contenido: p.Contenido, MediaUrl: p.MediaURL,
		MediaType: p.MediaType, Reactions: parseReactions(p.Reactions),
		CreatedAt: p.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

type ForoComentario struct {
	ID        string    `db:"id"`
	PostID    string    `db:"post_id"`
	ParentID  *string   `db:"parent_id"`
	UserID    string    `db:"user_id"`
	UserName  string    `db:"user_name"`
	Contenido string    `db:"contenido"`
	Reactions []byte    `db:"reactions"`
	CreatedAt time.Time `db:"created_at"`
}

func (c *ForoComentario) ToProto() *forospb.ComentarioResponse {
	pid := ""
	if c.ParentID != nil {
		pid = *c.ParentID
	}
	return &forospb.ComentarioResponse{
		Id: c.ID, PostId: c.PostID, ParentId: pid, UserId: c.UserID, UserName: c.UserName,
		Contenido: c.Contenido, Reactions: parseReactions(c.Reactions),
		CreatedAt: c.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// ForosRepository contrato.
type ForosRepository interface {
	ListPosts(ctx context.Context, leccionID, userID string) ([]*ForoPost, error)
	CreatePost(ctx context.Context, req *forospb.CreatePostRequest) (*ForoPost, error)
	DeletePost(ctx context.Context, postID, userID string, isAdmin bool) error
	ListComentarios(ctx context.Context, postID, userID string) ([]*ForoComentario, error)
	CreateComentario(ctx context.Context, req *forospb.CreateComentarioRequest) (*ForoComentario, error)
	TogglePostReaction(ctx context.Context, req *forospb.PostReactionRequest) (*forospb.ReactionResponse, error)
	ToggleComentarioReaction(ctx context.Context, req *forospb.ComentarioReactionRequest) (*forospb.ReactionResponse, error)
}

type postgresForosRepository struct{ db *sqlx.DB }

func NewForosRepository(db *sqlx.DB) ForosRepository {
	return &postgresForosRepository{db: db}
}

// metaVal extrae un valor del gRPC incoming metadata.
func metaVal(ctx context.Context, key string) string {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if vals := md.Get(key); len(vals) > 0 {
			return vals[0]
		}
	}
	return ""
}

func (r *postgresForosRepository) ListPosts(ctx context.Context, leccionID, userID string) ([]*ForoPost, error) {
	query := `
		SELECT fp.id, fp.leccion_id, fp.user_id,
		       COALESCE(fp.user_name,'') user_name,
		       fp.titulo, fp.contenido,
		       COALESCE(fp.media_url,'') media_url,
		       COALESCE(fp.media_type,'') media_type,
		       fp.created_at,
		       COALESCE(
		         (SELECT json_agg(json_build_object(
		            'emoji', sub.emoji,
		            'count', sub.cnt,
		            'user_reacted', sub.user_reacted
		          ))
		          FROM (
		            SELECT emoji, COUNT(*) as cnt,
		                   BOOL_OR(user_id = NULLIF($2,'')::uuid) as user_reacted
		            FROM foro_post_reactions
		            WHERE post_id = fp.id
		            GROUP BY emoji
		          ) sub), '[]'::json
		       ) as reactions
		  FROM foro_posts fp
		  -- Isolate to same licencia / cohort:
		  -- A user sees posts only from people in the same licencia as them.
		  -- Admins/instructors see all posts in the leccion.
		  JOIN lecciones l ON l.id = fp.leccion_id
		  WHERE fp.leccion_id = $1::uuid AND fp.deleted_at IS NULL
		    AND (
		        -- The post author shares same licencia as the requester
		        EXISTS (
		          SELECT 1 FROM inscripciones i_req
		          JOIN inscripciones i_author ON
		            i_author.user_id = fp.user_id
		            AND i_author.capacitacion_id = i_req.capacitacion_id
		            AND i_author.licencia_id IS NOT DISTINCT FROM i_req.licencia_id
		          WHERE i_req.user_id = NULLIF($2,'')::uuid
		            AND i_req.capacitacion_id = l.capacitacion_id
		        )
		        OR
		        -- Or the user is admin / instructor (they see everything)
		        EXISTS (
		          SELECT 1 FROM users u WHERE u.id = NULLIF($2,'')::uuid
		            AND u.role IN ('admin', 'instructor')
		        )
		        -- Or the user is the post author
		        OR fp.user_id = NULLIF($2,'')::uuid
		    )
		 ORDER BY fp.created_at DESC`
	var posts []*ForoPost
	return posts, r.db.SelectContext(ctx, &posts, query, leccionID, userID)
}

func (r *postgresForosRepository) CreatePost(ctx context.Context, req *forospb.CreatePostRequest) (*ForoPost, error) {
	userName := metaVal(ctx, "x-user-name")
	var id string
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO foro_posts(leccion_id,user_id,user_name,titulo,contenido,media_url,media_type)
		 VALUES($1::uuid,$2::uuid,$3,$4,$5,$6,$7) RETURNING id`,
		req.LeccionId, req.UserId, userName, req.Titulo, req.Contenido, req.MediaUrl, req.MediaType,
	).Scan(&id)
	if err != nil {
		return nil, err
	}
	var p ForoPost
	return &p, r.db.GetContext(ctx, &p,
		`SELECT id, leccion_id, user_id,
		        COALESCE(user_name,'') user_name,
		        titulo, contenido,
		        COALESCE(media_url,'') media_url,
		        COALESCE(media_type,'') media_type,
		        '[]'::json as reactions, created_at
		   FROM foro_posts WHERE id=$1::uuid`, id)
}

func (r *postgresForosRepository) DeletePost(ctx context.Context, postID, userID string, isAdmin bool) error {
	var err error
	if isAdmin {
		_, err = r.db.ExecContext(ctx,
			`UPDATE foro_posts SET deleted_at=NOW() WHERE id=$1::uuid`, postID)
	} else {
		_, err = r.db.ExecContext(ctx,
			`UPDATE foro_posts SET deleted_at=NOW() WHERE id=$1::uuid AND user_id=$2::uuid`, postID, userID)
	}
	return err
}

func (r *postgresForosRepository) ListComentarios(ctx context.Context, postID, userID string) ([]*ForoComentario, error) {
	var cs []*ForoComentario
	query := `SELECT c.id, c.post_id, c.parent_id, c.user_id,
		        COALESCE(c.user_name,'') user_name,
		        c.contenido, c.created_at,
		        COALESCE(
		         (SELECT json_agg(json_build_object(
		            'emoji', sub.emoji,
		            'count', sub.cnt,
		            'user_reacted', sub.user_reacted
		          ))
		          FROM (
		            SELECT emoji, COUNT(*) as cnt,
		                   BOOL_OR(user_id = NULLIF($2,'')::uuid) as user_reacted
		            FROM foro_comentario_reactions
		            WHERE comentario_id = c.id
		            GROUP BY emoji
		          ) sub), '[]'::json
		       ) as reactions
		   FROM foro_comentarios c
		  WHERE c.post_id=$1::uuid ORDER BY c.created_at ASC`
	return cs, r.db.SelectContext(ctx, &cs, query, postID, userID)
}

func (r *postgresForosRepository) CreateComentario(ctx context.Context, req *forospb.CreateComentarioRequest) (*ForoComentario, error) {
	userName := metaVal(ctx, "x-user-name")
	var id string
	var parentID *string
	if req.ParentId != "" {
		parentID = &req.ParentId
	}
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO foro_comentarios(post_id,user_id,user_name,contenido,parent_id) VALUES($1::uuid,$2::uuid,$3,$4,$5::uuid) RETURNING id`,
		req.PostId, req.UserId, userName, req.Contenido, parentID,
	).Scan(&id)
	if err != nil {
		return nil, err
	}
	var c ForoComentario
	return &c, r.db.GetContext(ctx, &c,
		`SELECT id, post_id, parent_id, user_id,
		        COALESCE(user_name,'') user_name,
		        contenido, '[]'::json as reactions, created_at
		   FROM foro_comentarios WHERE id=$1::uuid`, id)
}

func (r *postgresForosRepository) TogglePostReaction(ctx context.Context, req *forospb.PostReactionRequest) (*forospb.ReactionResponse, error) {
	var exists bool
	r.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM foro_post_reactions WHERE post_id=$1::uuid AND user_id=$2::uuid AND emoji=$3)`,
		req.PostId, req.UserId, req.Emoji,
	).Scan(&exists)

	if exists {
		r.db.ExecContext(ctx, `DELETE FROM foro_post_reactions WHERE post_id=$1::uuid AND user_id=$2::uuid AND emoji=$3`, req.PostId, req.UserId, req.Emoji)
	} else {
		r.db.ExecContext(ctx, `INSERT INTO foro_post_reactions(post_id,user_id,emoji) VALUES($1::uuid,$2::uuid,$3)`, req.PostId, req.UserId, req.Emoji)
	}
	return r.getPostReactions(ctx, req.PostId, req.UserId)
}

func (r *postgresForosRepository) ToggleComentarioReaction(ctx context.Context, req *forospb.ComentarioReactionRequest) (*forospb.ReactionResponse, error) {
	var exists bool
	r.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM foro_comentario_reactions WHERE comentario_id=$1::uuid AND user_id=$2::uuid AND emoji=$3)`,
		req.ComentarioId, req.UserId, req.Emoji,
	).Scan(&exists)

	if exists {
		r.db.ExecContext(ctx, `DELETE FROM foro_comentario_reactions WHERE comentario_id=$1::uuid AND user_id=$2::uuid AND emoji=$3`, req.ComentarioId, req.UserId, req.Emoji)
	} else {
		r.db.ExecContext(ctx, `INSERT INTO foro_comentario_reactions(comentario_id,user_id,emoji) VALUES($1::uuid,$2::uuid,$3)`, req.ComentarioId, req.UserId, req.Emoji)
	}
	return r.getComentarioReactions(ctx, req.ComentarioId, req.UserId)
}

func (r *postgresForosRepository) getPostReactions(ctx context.Context, postID, userID string) (*forospb.ReactionResponse, error) {
	var data []byte
	err := r.db.QueryRowContext(ctx, `
		SELECT COALESCE(
		  (SELECT json_agg(json_build_object('emoji', sub.emoji, 'count', sub.cnt, 'user_reacted', sub.user_reacted))
		   FROM (
		     SELECT emoji, COUNT(*) as cnt, BOOL_OR(user_id = NULLIF($2,'')::uuid) as user_reacted
		     FROM foro_post_reactions WHERE post_id = $1::uuid GROUP BY emoji
		   ) sub), '[]'::json)`, postID, userID).Scan(&data)
	if err != nil {
		return nil, err
	}
	return &forospb.ReactionResponse{Reactions: parseReactions(data)}, nil
}

func (r *postgresForosRepository) getComentarioReactions(ctx context.Context, comentarioID, userID string) (*forospb.ReactionResponse, error) {
	var data []byte
	err := r.db.QueryRowContext(ctx, `
		SELECT COALESCE(
		  (SELECT json_agg(json_build_object('emoji', sub.emoji, 'count', sub.cnt, 'user_reacted', sub.user_reacted))
		   FROM (
		     SELECT emoji, COUNT(*) as cnt, BOOL_OR(user_id = NULLIF($2,'')::uuid) as user_reacted
		     FROM foro_comentario_reactions WHERE comentario_id = $1::uuid GROUP BY emoji
		   ) sub), '[]'::json)`, comentarioID, userID).Scan(&data)
	if err != nil {
		return nil, err
	}
	return &forospb.ReactionResponse{Reactions: parseReactions(data)}, nil
}

func parseReactions(data []byte) []*forospb.Reaction {
	if len(data) == 0 {
		return nil
	}
	var raw []struct {
		Emoji       string `json:"emoji"`
		Count       int32  `json:"count"`
		UserReacted bool   `json:"user_reacted"`
	}
	_ = json.Unmarshal(data, &raw)
	res := make([]*forospb.Reaction, len(raw))
	for i, r := range raw {
		res[i] = &forospb.Reaction{
			Emoji:       r.Emoji,
			Count:       r.Count,
			UserReacted: r.UserReacted,
		}
	}
	return res
}
