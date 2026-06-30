package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"time"

	cursospb "Prueba-Go/gen/cursos"

	"github.com/google/uuid"
)

// ListSchedules obtiene la disponibilidad de los instructores
func (r *postgresCursosRepository) ListSchedules(ctx context.Context, instructorID *string) ([]*InstructorSchedule, error) {
	var schedules []*InstructorSchedule
	query := `SELECT id, instructor_id, start_time, end_time, status, created_at FROM instructor_schedules`
	var args []interface{}
	
	if instructorID != nil {
		query += ` WHERE instructor_id = $1`
		args = append(args, *instructorID)
	}
	query += ` ORDER BY start_time ASC`

	err := r.db.SelectContext(ctx, &schedules, query, args...)
	if err != nil {
		return nil, err
	}
	return schedules, nil
}

// CreateSchedule crea un horario disponible
func (r *postgresCursosRepository) CreateSchedule(ctx context.Context, req *cursospb.CreateScheduleRequest) (*InstructorSchedule, error) {
	var s InstructorSchedule
	query := `
		INSERT INTO instructor_schedules (instructor_id, start_time, end_time, status)
		VALUES ($1, $2, $3, 'available')
		RETURNING id, instructor_id, start_time, end_time, status, created_at
	`
	err := r.db.QueryRowxContext(ctx, query, req.InstructorId, req.StartTime, req.EndTime).StructScan(&s)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// UpdateSchedule actualiza un horario
func (r *postgresCursosRepository) UpdateSchedule(ctx context.Context, req *cursospb.UpdateScheduleRequest) (*InstructorSchedule, error) {
	var s InstructorSchedule
	query := `
		UPDATE instructor_schedules 
		SET start_time = COALESCE(NULLIF($2, ''), start_time::text)::timestamptz,
		    end_time = COALESCE(NULLIF($3, ''), end_time::text)::timestamptz,
		    status = COALESCE(NULLIF($4, ''), status)
		WHERE id = $1
		RETURNING id, instructor_id, start_time, end_time, status, created_at
	`
	err := r.db.QueryRowxContext(ctx, query, req.ScheduleId, req.StartTime, req.EndTime, req.Status).StructScan(&s)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// DeleteSchedule borra un horario
func (r *postgresCursosRepository) DeleteSchedule(ctx context.Context, scheduleID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM instructor_schedules WHERE id = $1`, scheduleID)
	return err
}

// CreateVideocallTickets crea N códigos únicos para acceso a videollamada
func (r *postgresCursosRepository) CreateVideocallTickets(ctx context.Context, capacitacionID string, licenciaID *string, scheduleID *string, count int) ([]*VideocallTicket, error) {
	var tickets []*VideocallTicket
	
	// Generar 'count' tickets en una sola transacción o batch
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO videocall_tickets (capacitacion_id, licencia_id, schedule_id, codigo)
		VALUES ($1, $2, $3, $4)
		RETURNING id, capacitacion_id, licencia_id, schedule_id, codigo, in_use_by_user_id, is_valid, created_at
	`
	
	for i := 0; i < count; i++ {
		// Generar un código único corto, ej. VC-UUID[:8]
		codigo := fmt.Sprintf("VC-%s", uuid.New().String()[:8])
		
		var t VideocallTicket
		var schID interface{}
		if scheduleID != nil && *scheduleID != "" {
			schID = *scheduleID
		}
		
		if err := tx.QueryRowxContext(ctx, query, capacitacionID, licenciaID, schID, codigo).StructScan(&t); err != nil {
			return nil, err
		}
		tickets = append(tickets, &t)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return tickets, nil
}

// JoinVideocall marca el ticket en uso y retorna el room de Jitsi
func (r *postgresCursosRepository) JoinVideocall(ctx context.Context, codigo, userID string) (string, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	var t VideocallTicket
	// Bloquear fila
	err = tx.QueryRowxContext(ctx, `SELECT * FROM videocall_tickets WHERE UPPER(codigo) = UPPER($1) FOR UPDATE`, codigo).StructScan(&t)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", errors.New("código no válido")
		}
		return "", err
	}

	if !t.IsValid {
		return "", errors.New("este código ya no es válido o la llamada terminó")
	}

	if t.InUseByUserID != nil && *t.InUseByUserID != userID {
		return "", errors.New("el código está siendo usado por otra persona en este momento")
	}

	// Verificar estado de la videollamada
	var cursoTitle string
	err = tx.QueryRowContext(ctx, `SELECT title FROM capacitaciones WHERE id = $1`, t.CapacitacionID).Scan(&cursoTitle)
	if err != nil {
		return "", err
	}
	
	re := regexp.MustCompile(`[^a-zA-Z0-9]+`)
	safeTitle := re.ReplaceAllString(cursoTitle, "")
	if len(safeTitle) > 30 {
		safeTitle = safeTitle[:30]
	}

	var roomName string
	if t.ScheduleID != nil && *t.ScheduleID != "" {
		var status string
		var endTime time.Time
		err = tx.QueryRowContext(ctx, `SELECT status, end_time FROM instructor_schedules WHERE id = $1`, *t.ScheduleID).Scan(&status, &endTime)
		if err != nil {
			return "", err
		}
		
		if status == "finished" || time.Now().After(endTime.Add(2*time.Hour)) {
			tx.ExecContext(ctx, `UPDATE videocall_tickets SET is_valid = false WHERE id = $1`, t.ID)
			return "", errors.New("esta sesión de videollamada ya ha finalizado")
		}
		
		roomName = fmt.Sprintf("%s-%s", safeTitle, (*t.ScheduleID)[:8])
	} else {
		roomName = fmt.Sprintf("%s-%s", safeTitle, t.CapacitacionID[:8])
	}

	_, err = tx.ExecContext(ctx, `UPDATE videocall_tickets SET in_use_by_user_id = $1 WHERE id = $2`, userID, t.ID)
	if err != nil {
		return "", err
	}

	if err := tx.Commit(); err != nil {
		return "", err
	}

	// El room name es el UUID de la capacitación
	return roomName, nil
}

// LeaveVideocall libera el uso del código
func (r *postgresCursosRepository) LeaveVideocall(ctx context.Context, codigo, userID string) error {
	res, err := r.db.ExecContext(ctx, `UPDATE videocall_tickets SET in_use_by_user_id = NULL WHERE UPPER(codigo) = UPPER($1) AND in_use_by_user_id = $2`, codigo, userID)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return errors.New("no estás en esta videollamada con este código")
	}
	return nil
}

// EndVideocall invalida todos los códigos y la videollamada
func (r *postgresCursosRepository) EndVideocall(ctx context.Context, cursoID string, scheduleID *string) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if scheduleID != nil && *scheduleID != "" {
		_, err = tx.ExecContext(ctx, `UPDATE videocall_tickets SET is_valid = false, in_use_by_user_id = NULL WHERE capacitacion_id = $1 AND schedule_id = $2`, cursoID, *scheduleID)
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `UPDATE instructor_schedules SET status = 'finished' WHERE id = $1`, *scheduleID)
		if err != nil {
			return err
		}
	} else {
		_, err = tx.ExecContext(ctx, `UPDATE capacitaciones SET videocall_status = 'finished' WHERE id = $1`, cursoID)
		if err != nil {
			return err
		}

		_, err = tx.ExecContext(ctx, `UPDATE videocall_tickets SET is_valid = false, in_use_by_user_id = NULL WHERE capacitacion_id = $1`, cursoID)
		if err != nil {
			return err
		}
	}
	
	// Aquí podríamos agregar lógica para marcar inscripciones/progreso, 
	// pero por simplicidad con que is_valid=false los usuarios ya no pueden entrar.
	// La asistencia puede considerarse como: Si tienes un ticket, estás inscrito.
	// Haremos las asignaciones.
	
	// Obtener los usuarios que estuvieron en la llamada alguna vez (tickets asignados)
	// Pero no guardamos historial de quienes entraron si in_use se pone en NULL al salir.
	// Como compran los tickets, la Licencia ya les da acceso.
	// No es necesario "completar" explicitamente nada para videollamadas, simplemente ya no pueden entrar.

	return tx.Commit()
}

// ListTicketsByLicencia obtiene todos los tickets de una licencia B2B
func (r *postgresCursosRepository) ListTicketsByLicencia(ctx context.Context, licenciaID string) ([]*VideocallTicket, error) {
	var tickets []*VideocallTicket
	query := `SELECT id, capacitacion_id, licencia_id, schedule_id, owner_id, codigo, in_use_by_user_id, is_valid, created_at 
	          FROM videocall_tickets WHERE licencia_id = $1`
	err := r.db.SelectContext(ctx, &tickets, query, licenciaID)
	return tickets, err
}

func (r *postgresCursosRepository) AssignTicketToUser(ctx context.Context, ticketID, userID string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE videocall_tickets SET owner_id = $1 WHERE id = $2`, userID, ticketID)
	return err
}

func (r *postgresCursosRepository) GetTicketForUserAndCourse(ctx context.Context, userID, cursoID string) (*VideocallTicket, error) {
	var t VideocallTicket
	query := `SELECT id, capacitacion_id, licencia_id, schedule_id, owner_id, codigo, in_use_by_user_id, is_valid, created_at 
	          FROM videocall_tickets WHERE owner_id = $1 AND capacitacion_id = $2 AND is_valid = true LIMIT 1`
	err := r.db.GetContext(ctx, &t, query, userID, cursoID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, err
	}
	return &t, nil
}

func (r *postgresCursosRepository) GetCurrentScheduleForInstructor(ctx context.Context, instructorID, cursoID string) (*string, error) {
	var scheduleID string
	// Find the most recently created active/booked schedule for this instructor
	// We don't link schedules to courses in the DB directly, wait!
	// Oh! `instructor_schedules` doesn't have `curso_id`! It's just an instructor schedule!
	// But `videocall_tickets` links `schedule_id` to `capacitacion_id`.
	query := `SELECT s.id FROM instructor_schedules s
	          JOIN videocall_tickets t ON t.schedule_id = s.id
	          WHERE s.instructor_id = $1 AND t.capacitacion_id = $2 
			  AND (s.status = 'booked' OR s.status = 'active')
	          ORDER BY s.created_at DESC LIMIT 1`
	err := r.db.GetContext(ctx, &scheduleID, query, instructorID, cursoID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, err
	}
	return &scheduleID, nil
}
