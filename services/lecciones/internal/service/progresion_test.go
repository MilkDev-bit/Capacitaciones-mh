package service

import (
	"errors"
	"strings"
	"testing"

	leccionespb "Prueba-Go/gen/lecciones"
)

func lec(id string, completada bool) *leccionespb.LeccionResponse {
	return &leccionespb.LeccionResponse{Id: id, Title: "Lección " + id, Completada: completada}
}

// Árbol de referencia: un módulo con dos lecciones y un submódulo con una,
// otro módulo con una, y al final una lección suelta.
func arbol(completadas ...string) *leccionespb.CursoTreeResponse {
	hecho := map[string]bool{}
	for _, id := range completadas {
		hecho[id] = true
	}
	return &leccionespb.CursoTreeResponse{
		Modulos: []*leccionespb.ModuloResponse{
			{
				Id:        "m1",
				Lecciones: []*leccionespb.LeccionResponse{lec("a", hecho["a"]), lec("b", hecho["b"])},
				Submodulos: []*leccionespb.SubmoduloResponse{
					{Id: "s1", Lecciones: []*leccionespb.LeccionResponse{lec("c", hecho["c"])}},
				},
			},
			{Id: "m2", Lecciones: []*leccionespb.LeccionResponse{lec("d", hecho["d"])}},
		},
		Lecciones: []*leccionespb.LeccionResponse{lec("z", hecho["z"])},
	}
}

func ids(ls []*leccionespb.LeccionResponse) []string {
	out := make([]string, 0, len(ls))
	for _, l := range ls {
		out = append(out, l.Id)
	}
	return out
}

// El orden es la regla entera: si el servidor ordenara distinto que la barra
// lateral, el alumno vería abierta una lección que el servidor rechaza y se
// quedaría sin forma de avanzar.
func TestAplanarArbolSigueElOrdenDelCurso(t *testing.T) {
	got := ids(AplanarArbol(arbol()))
	want := []string{"a", "b", "c", "d", "z"}

	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("orden = %v, se esperaba %v", got, want)
	}
}

func TestAplanarArbolTolerarNulos(t *testing.T) {
	if AplanarArbol(nil) != nil {
		t.Fatal("un árbol nil debe dar una lista vacía")
	}
	conNulos := &leccionespb.CursoTreeResponse{
		Modulos: []*leccionespb.ModuloResponse{
			nil,
			{Id: "m", Submodulos: []*leccionespb.SubmoduloResponse{nil}, Lecciones: []*leccionespb.LeccionResponse{lec("a", false)}},
		},
	}
	if got := ids(AplanarArbol(conNulos)); len(got) != 1 || got[0] != "a" {
		t.Fatalf("con nulos intercalados = %v", got)
	}
}

func TestPrimeraPendiente(t *testing.T) {
	casos := []struct {
		nombre      string
		completadas []string
		quiero      int
	}{
		{"curso sin empezar", nil, 0},
		{"dos hechas", []string{"a", "b"}, 2},
		{"curso terminado", []string{"a", "b", "c", "d", "z"}, 5},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			if got := PrimeraPendiente(AplanarArbol(arbol(c.completadas...))); got != c.quiero {
				t.Fatalf("frontera = %d, se esperaba %d", got, c.quiero)
			}
		})
	}
}

func TestValidarOrdenPermiteLaQueToca(t *testing.T) {
	if err := ValidarOrden(AplanarArbol(arbol()), "a"); err != nil {
		t.Fatalf("la primera lección debe permitirse: %v", err)
	}
	if err := ValidarOrden(AplanarArbol(arbol("a", "b")), "c"); err != nil {
		t.Fatalf("la lección en la frontera debe permitirse: %v", err)
	}
}

// El caso que motivó la validación de servidor: llamar directo al endpoint
// para saltarse el curso y salir con la constancia.
func TestValidarOrdenRechazaElSalto(t *testing.T) {
	err := ValidarOrden(AplanarArbol(arbol()), "z")
	if !errors.Is(err, ErrLeccionBloqueada) {
		t.Fatalf("saltar al final debía bloquearse, err = %v", err)
	}
	// El mensaje nombra la lección que falta para que el alumno sepa a dónde ir.
	if !strings.Contains(err.Error(), "Lección a") {
		t.Fatalf("el error debe nombrar la lección pendiente: %q", err.Error())
	}
}

func TestValidarOrdenRechazaSaltarUnModulo(t *testing.T) {
	// Terminado el primer módulo salvo su submódulo, "d" (módulo 2) no se abre.
	if err := ValidarOrden(AplanarArbol(arbol("a", "b")), "d"); !errors.Is(err, ErrLeccionBloqueada) {
		t.Fatalf("el módulo siguiente debía estar bloqueado, err = %v", err)
	}
	// Con el módulo 1 completo, sí.
	if err := ValidarOrden(AplanarArbol(arbol("a", "b", "c")), "d"); err != nil {
		t.Fatalf("con el módulo anterior completo debía permitirse: %v", err)
	}
}

// Repasar es un caso normal: el reproductor reintenta y el alumno revisa. Si
// repetir exigiera el orden, revisar la lección 2 obligaría a rehacer el curso.
func TestValidarOrdenPermiteRepetirLoYaHecho(t *testing.T) {
	if err := ValidarOrden(AplanarArbol(arbol("a", "b", "c")), "a"); err != nil {
		t.Fatalf("repetir una lección completada debe permitirse: %v", err)
	}
}

func TestValidarOrdenNoOpinaSobreLoQueNoEstaEnElArbol(t *testing.T) {
	// Una lección recién creada o un curso a medio migrar no debe dejar al
	// alumno encerrado por un problema de datos que no causó él.
	if err := ValidarOrden(AplanarArbol(arbol()), "fantasma"); err != nil {
		t.Fatalf("una lección fuera del árbol no debe bloquearse: %v", err)
	}
	if err := ValidarOrden(nil, "a"); err != nil {
		t.Fatalf("sin árbol no se bloquea: %v", err)
	}
}

func TestValidarOrdenCursoTerminado(t *testing.T) {
	planas := AplanarArbol(arbol("a", "b", "c", "d", "z"))
	for _, id := range []string{"a", "c", "z"} {
		if err := ValidarOrden(planas, id); err != nil {
			t.Fatalf("con el curso terminado nada se bloquea (%s): %v", id, err)
		}
	}
}
