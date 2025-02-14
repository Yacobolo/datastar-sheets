package routes

import (
	"log"
	"myproject/web/components"
	"myproject/web/pages"
	"net/http"

	"github.com/go-chi/chi/v5"
	datastar "github.com/starfederation/datastar/sdk/go"
)

type cellSignal struct {
	Cell string `json:"cell"`
	Key  string `json:"keyPressed"`
}

func setupIndexRoute(router chi.Router) error {
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		pages.Index("best title").Render(r.Context(), w)
	})

	router.Get("/cell/{cellId}", func(w http.ResponseWriter, r *http.Request) {
		cellId := chi.URLParam(r, "cellId")
		// read signals into bytes

		cellData := &cellSignal{}
		if err := datastar.ReadSignals(r, cellData); err != nil {

			http.Error(w, err.Error(), http.StatusBadRequest)

			log.Println("Failed to read signals:", err) // Updated to use log.Println
		}
		sse := datastar.NewSSE(w, r)
		sse.MergeFragmentTempl(components.InactiveCell(cellId, cellData.Cell))
	})

	router.Get("/cell/{cellId}/edit", func(w http.ResponseWriter, r *http.Request) {
		cellId := chi.URLParam(r, "cellId")
		// Extract query parameter `value`
		value := r.URL.Query().Get("value")

		cellData := &cellSignal{}
		if err := datastar.ReadSignals(r, cellData); err != nil {

			http.Error(w, err.Error(), http.StatusBadRequest)

			log.Println("Failed to read signals:", err) // Updated to use log.Println
		}
		// print celldata
		cellValue := value + cellData.Key
		sse := datastar.NewSSE(w, r)
		sse.MergeFragmentTempl(components.Cell(cellId, []string{}))

		sse.MarshalAndMergeSignals(map[string]any{
			"cell":       cellValue,
			"keyPressed": "",
		})
	})

	router.Get("/cell/{cellId}/search", func(w http.ResponseWriter, r *http.Request) {
		cellId := chi.URLParam(r, "cellId")
		// Extract query parameter `value`

		cellData := &cellSignal{}
		if err := datastar.ReadSignals(r, cellData); err != nil {

			http.Error(w, err.Error(), http.StatusBadRequest)

			log.Println("Failed to read signals:", err) // Updated to use log.Println
		}
		// print celldata
		sse := datastar.NewSSE(w, r)
		// searchKey := cellData.Cell
		// startswith =
		if cellData.Cell == "=" {

			options := []string{"hi", "there"}
			sse.MergeFragmentTempl(components.Cell(cellId, options))
		}
	})
	return nil
}
