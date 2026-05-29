package repository

import (
	"context"
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
	LikeCount int32     `db:"like_count"`
	UserLiked bool      `db:"user_liked"`
	CreatedAt time.Time `db:"created_at"`
}

func (p *ForoPost) ToProto() *forospb.PostResponse {
	return &forospb.PostResponse{
		Id: p.ID, LeccionId: p.LeccionID, UserId: p.UserID, UserName: p.UserName,
		Titulo: p.Titulo, Contenido: p.Contenido, MediaUrl: p.MediaURL,
		MediaType: p.MediaType, LikeCount: p.LikeCount, UserLiked: p.UserLiked,
		CreatedAt: p.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

type ForoComentario struct {
	ID        string    `db:"id"`
	PostID    string    `db:"post_id"`
	UserID    string    `db:"user_id"`
	UserName  string    `db:"user_name"`
	Contenido string    `db:"contenido"`
	CreatedAt time.Time `db:"created_at"`
}

func (c *ForoComentario) ToProto() *forospb.ComentarioResponse {
	return &forospb.ComentarioResponse{
		Id: c.ID, PostId: c.PostID, UserId: c.UserID, UserName: c.UserName,
		Contenido: c.Contenido, CreatedAt: c.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// ForosRepository contrato.
type ForosRepository interface {
	ListPosts(ctx context.Context, leccionID, userID string) ([]*ForoPost, error)
	CreatePost(ctx context.Context, req *forospb.CreatePostRequest) (*ForoPost, error)
	DeletePost(ctx context.Context, postID, userID string, isAdmin bool) error
	ListComentarios(ctx context.Context, postID string) ([]*ForoComentario, error)
	CreateComentario(ctx context.Context, req *forospb.CreateComentarioRequest) (*ForoComentario, error)
	ToggleLike(ctx context.Context, postID, userID string) (*forospb.LikeResponse, error)
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
		       COUNT(fl.id)::int like_count,
		       BOOL_OR(fl.user_id = $2) user_liked,
		       fp.created_at
		  FROM foro_posts fp
		  LEFT JOIN foro_likes fl ON fl.post_id = fp.id
		 WHERE fp.leccion_id = $1 AND fp.deleted_at IS NULL
		 GROUP BY fp.id
		 ORDER BY fp.created_at DESC`
	var posts []*ForoPost
	return posts, r.db.SelectContext(ctx, &posts, query, leccionID, userID)
}

func (r *postgresForosRepository) CreatePost(ctx context.Context, req *forospb.CreatePostRequest) (*ForoPost, error) {
	userName := metaVal(ctx, "x-user-name")
	var id string
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO foro_posts(leccion_id,user_id,user_name,titulo,contenido,media_url,media_type)
		 VALUES($1,$2,$3,$4,$5,$6,$7) RETURNING id`,
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
		        0::int like_count, false user_liked, created_at
		   FROM foro_posts WHERE id=$1`, id)
}

func (r *postgresForosRepository) DeletePost(ctx context.Context, postID, userID string, isAdmin bool) error {
	var err error
	if isAdmin {
		_, err = r.db.ExecContext(ctx,
			`UPDATE foro_posts SET deleted_at=NOW() WHERE id=$1`, postID)
	} else {
		_, err = r.db.ExecContext(ctx,
			`UPDATE foro_posts SET deleted_at=NOW() WHERE id=$1 AND user_id=$2`, postID, userID)
	}
	return err
}

func (r *postgresForosRepository) ListComentarios(ctx context.Context, postID string) ([]*ForoComentario, error) {
	var cs []*ForoComentario
	return cs, r.db.SelectContext(ctx, &cs,
		`SELECT id, post_id, user_id,
		        COALESCE(user_name,'') user_name,
		        contenido, created_at
		   FROM foro_comentarios
		  WHERE post_id=$1 ORDER BY created_at ASC`, postID)
}

func (r *postgresForosRepository) CreateComentario(ctx context.Context, req *forospb.CreateComentarioRequest) (*ForoComentario, error) {
	userName := metaVal(ctx, "x-user-name")
	var id string
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO foro_comentarios(post_id,user_id,user_name,contenido) VALUES($1,$2,$3,$4) RETURNING id`,
		req.PostId, req.UserId, userName, req.Contenido,
	).Scan(&id)
	if err != nil {
		return nil, err
	}
	var c ForoComentario
	return &c, r.db.GetContext(ctx, &c,
		`SELECT id, post_id, user_id,
		        COALESCE(user_name,'') user_name,
		        contenido, created_at
		   FROM foro_comentarios WHERE id=$1`, id)
}

func (r *postgresForosRepository) ToggleLike(ctx context.Context, postID, userID string) (*forospb.LikeResponse, error) {
	var exists bool
	r.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM foro_likes WHERE post_id=$1 AND user_id=$2)`,
		postID, userID,
	).Scan(&exists)

	if exists {
		r.db.ExecContext(ctx,
			`DELETE FROM foro_likes WHERE post_id=$1 AND user_id=$2`, postID, userID)
	} else {
		r.db.ExecContext(ctx,
			`INSERT INTO foro_likes(post_id,user_id) VALUES($1,$2)`, postID, userID)
	}

	var count int32
	r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM foro_likes WHERE post_id=$1`, postID,
	).Scan(&count)

	return &forospb.LikeResponse{LikeCount: count, UserLiked: !exists}, nil
}
