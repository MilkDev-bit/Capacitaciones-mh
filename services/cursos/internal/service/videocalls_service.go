package service

import (
	"context"
	"fmt"
	"strings"

	cursospb "Prueba-Go/gen/cursos"
	"Prueba-Go/services/cursos/internal/repository"
)

// JoinVideocall une a un usuario a una videollamada usando un código
func (s *CursosService) JoinVideocall(ctx context.Context, req *cursospb.JoinVideocallRequest) (*cursospb.JoinVideocallResponse, error) {
	roomName, err := s.repo.JoinVideocall(ctx, req.Codigo, req.UserId)
	if err != nil {
		if strings.Contains(err.Error(), "código no válido") {
			return nil, fmt.Errorf("%w: código no válido", ErrNotFound)
		}
		if strings.Contains(err.Error(), "ya no es válido") {
			return nil, fmt.Errorf("%w: este código ya no es válido o la llamada terminó", ErrForbidden)
		}
		if strings.Contains(err.Error(), "usado por otra persona") {
			return nil, fmt.Errorf("%w: el código está siendo usado por otra persona", ErrConflict)
		}
		return nil, err
	}

	return &cursospb.JoinVideocallResponse{
		RoomName: roomName,
		Token:    "", // Dejar vacío por ahora, se puede integrar JWT de Jitsi después si se requiere
	}, nil
}

// LeaveVideocall libera el código usado por un usuario
func (s *CursosService) LeaveVideocall(ctx context.Context, req *cursospb.LeaveVideocallRequest) error {
	return s.repo.LeaveVideocall(ctx, req.Codigo, req.UserId)
}

// EndVideocall finaliza la videollamada e invalida los tickets (solo instructor)
func (s *CursosService) EndVideocall(ctx context.Context, req *cursospb.CursoIDRequest) error {
	// Verificar que el usuario que lo solicita es el instructor
	curso, err := s.repo.FindByID(ctx, req.CursoId)
	if err != nil {
		return ErrNotFound
	}
	if curso.InstructorID == nil || *curso.InstructorID != req.UserId {
		return ErrForbidden
	}

	return s.repo.EndVideocall(ctx, req.CursoId)
}

// AdminListSchedules lista horarios
func (s *CursosService) AdminListSchedules(ctx context.Context, req *cursospb.UserRequest) (*cursospb.ListSchedulesResponse, error) {
	var filterID *string
	if req.UserId != "" {
		filterID = &req.UserId
	}
	schedules, err := s.repo.ListSchedules(ctx, filterID)
	if err != nil {
		return nil, err
	}

	resp := &cursospb.ListSchedulesResponse{}
	for _, sc := range schedules {
		resp.Schedules = append(resp.Schedules, sc.ToProto())
	}
	return resp, nil
}

// AdminCreateSchedule crea horario
func (s *CursosService) AdminCreateSchedule(ctx context.Context, req *cursospb.CreateScheduleRequest) (*cursospb.InstructorSchedule, error) {
	sc, err := s.repo.CreateSchedule(ctx, req)
	if err != nil {
		return nil, err
	}
	return sc.ToProto(), nil
}

// AdminUpdateSchedule actualiza horario
func (s *CursosService) AdminUpdateSchedule(ctx context.Context, req *cursospb.UpdateScheduleRequest) (*cursospb.InstructorSchedule, error) {
	sc, err := s.repo.UpdateSchedule(ctx, req)
	if err != nil {
		return nil, err
	}
	return sc.ToProto(), nil
}

// AdminDeleteSchedule elimina horario
func (s *CursosService) AdminDeleteSchedule(ctx context.Context, req *cursospb.ScheduleIDRequest) error {
	return s.repo.DeleteSchedule(ctx, req.ScheduleId)
}

// ListPublicSchedules lista horarios disponibles de un instructor
func (s *CursosService) ListPublicSchedules(ctx context.Context, req *cursospb.ListPublicSchedulesRequest) (*cursospb.ListPublicSchedulesResponse, error) {
	var filterID *string
	if req.InstructorId != "" {
		filterID = &req.InstructorId
	}
	schedules, err := s.repo.ListSchedules(ctx, filterID)
	if err != nil {
		return nil, err
	}

	resp := &cursospb.ListPublicSchedulesResponse{}
	for _, sc := range schedules {
		if sc.Status == "available" {
			resp.Schedules = append(resp.Schedules, sc.ToProto())
		}
	}
	return resp, nil
}

func (s *CursosService) ListTicketsByLicencia(ctx context.Context, licenciaID string) ([]*repository.VideocallTicket, error) { return s.repo.ListTicketsByLicencia(ctx, licenciaID) }
