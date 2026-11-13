package repository

import "testing"

// autorizaExamen decide quién puede leer las calificaciones de los alumnos de un
// examen. Antes de que existiera, InstructorGetResultados no comprobaba nada:
// bastaba con probar identificadores en la URL para leer las notas de los
// alumnos de otro instructor.
func TestAutorizaExamen(t *testing.T) {
	const (
		yo    = "u-instructor"
		otro  = "u-otro"
		curso = "c-walmart"
	)

	casos := []struct {
		nombre     string
		dueno      string
		curso      string
		usuario    string
		revisables []string
		quiere     bool
		porque     string
	}{
		{
			nombre: "el dueño del examen siempre pasa",
			dueno:  yo, curso: curso, usuario: yo,
			quiere: true,
			porque: "lo creó él",
		},
		{
			nombre: "un instructor ajeno no pasa",
			dueno:  otro, curso: curso, usuario: yo,
			quiere: false,
			porque: "ni lo creó ni le asignaron el curso",
		},
		{
			nombre: "el revisor del curso pasa aunque el examen sea de otro",
			dueno:  otro, curso: curso, usuario: yo, revisables: []string{curso},
			quiere: true,
			porque: "es justo lo que el administrador quiso al asignarle el curso",
		},
		{
			nombre: "revisar OTRO curso no sirve",
			dueno:  otro, curso: curso, usuario: yo, revisables: []string{"c-distinto"},
			quiere: false,
			porque: "la lista no incluye el curso de este examen",
		},
		{
			// Un examen suelto no está atado a ninguna asignación: por mucho que
			// le hayan dado cursos a revisar, ninguno lo cubre.
			nombre: "un examen sin curso solo lo ve su dueño",
			dueno:  otro, curso: "", usuario: yo, revisables: []string{curso, "c-otro"},
			quiere: false,
			porque: "no hay curso que relacione el examen con la asignación",
		},
		{
			nombre: "el dueño ve su propio examen sin curso",
			dueno:  yo, curso: "", usuario: yo,
			quiere: true,
			porque: "la titularidad no depende del curso",
		},
		{
			// Las llamadas internas del propio gateway no llevan instructor.
			nombre: "sin usuario se permite, para las llamadas internas",
			dueno:  otro, curso: curso, usuario: "",
			quiere: true,
			porque: "no hay un instructor a quien restringir",
		},
		{
			// Un examen antiguo puede tener instructor_id nulo. Eso NO debe
			// convertirse en "cualquiera lo ve": el string vacío no es un id.
			nombre: "un examen huérfano no se abre a cualquiera",
			dueno:  "", curso: curso, usuario: yo,
			quiere: false,
			porque: "dueño vacío no debe coincidir con nadie",
		},
		{
			// Este es el caso que hace falta el corte por capacitacionID vacío.
			//
			// Si la lista de cursos revisables trae una cadena vacía —un curso
			// mal construido aguas arriba— coincidiría con un examen sin curso y
			// abriría las notas de un examen ajeno. Comprobado: sin ese corte,
			// este caso pasa a devolver true.
			nombre: "un hueco en la lista no vale como comodín",
			dueno:  otro, curso: "", usuario: yo, revisables: []string{""},
			quiere: false,
			porque: "una cadena vacía no identifica ningún curso",
		},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			got := autorizaExamen(c.dueno, c.curso, c.usuario, c.revisables)
			if got != c.quiere {
				t.Fatalf("autorizaExamen(dueno=%q, curso=%q, usuario=%q, revisables=%v) = %v; se esperaba %v porque %s",
					c.dueno, c.curso, c.usuario, c.revisables, got, c.quiere, c.porque)
			}
		})
	}
}
