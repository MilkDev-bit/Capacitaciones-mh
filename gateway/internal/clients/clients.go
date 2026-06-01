// Package clients gestiona las conexiones gRPC a cada microservicio.
// Las conexiones se crean una vez al arranque y se reutilizan.
package clients

import (
	"Prueba-Go/gateway/internal/config"
	authpb "Prueba-Go/gen/auth"
	cursospb "Prueba-Go/gen/cursos"
	examenespb "Prueba-Go/gen/examenes"
	forospb "Prueba-Go/gen/foros"
	leccionespb "Prueba-Go/gen/lecciones"
	mensajespb "Prueba-Go/gen/mensajes"
	usuariospb "Prueba-Go/gen/usuarios"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Clients agrupa todos los clientes gRPC del Gateway.
type Clients struct {
	Auth      authpb.AuthServiceClient
	Usuarios  usuariospb.UsuariosServiceClient
	Cursos    cursospb.CursosServiceClient
	Lecciones leccionespb.LeccionesServiceClient
	Examenes  examenespb.ExamenesServiceClient
	Foros     forospb.ForosServiceClient
	Mensajes  mensajespb.MensajesServiceClient

	// conns para cerrarlas en shutdown
	conns []*grpc.ClientConn
}

// Dial crea todas las conexiones gRPC. Usar insecure dentro del clúster;
// en producción con mTLS reemplazar por credentials.NewTLS(...)
func Dial(cfg *config.Config) (*Clients, error) {
	c := &Clients{}
	dials := []struct {
		addr string
		fn   func(*grpc.ClientConn)
	}{
		{cfg.AuthAddr, func(cc *grpc.ClientConn) { c.Auth = authpb.NewAuthServiceClient(cc) }},
		{cfg.UsuariosAddr, func(cc *grpc.ClientConn) { c.Usuarios = usuariospb.NewUsuariosServiceClient(cc) }},
		{cfg.CursosAddr, func(cc *grpc.ClientConn) { c.Cursos = cursospb.NewCursosServiceClient(cc) }},
		{cfg.LeccionesAddr, func(cc *grpc.ClientConn) { c.Lecciones = leccionespb.NewLeccionesServiceClient(cc) }},
		{cfg.ExamenesAddr, func(cc *grpc.ClientConn) { c.Examenes = examenespb.NewExamenesServiceClient(cc) }},
		{cfg.ForosAddr, func(cc *grpc.ClientConn) { c.Foros = forospb.NewForosServiceClient(cc) }},
		{cfg.MensajesAddr, func(cc *grpc.ClientConn) { c.Mensajes = mensajespb.NewMensajesServiceClient(cc) }},
	}

	for _, d := range dials {
		conn, err := grpc.NewClient(d.addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			c.Close()
			return nil, err
		}
		c.conns = append(c.conns, conn)
		d.fn(conn)
	}
	return c, nil
}

// Close cierra todas las conexiones gRPC.
func (c *Clients) Close() {
	for _, conn := range c.conns {
		_ = conn.Close()
	}
}
