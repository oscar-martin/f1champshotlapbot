package servers

import (
	"fmt"
	"net/http"
)

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Hello World!\n")
	fmt.Fprintf(w, "No hay sesiones activas")
}
