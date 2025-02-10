package routes

import (
	"fmt"
	"log"
	"myproject/web/components"
	"myproject/web/pages"
	"net/http"

	"github.com/go-chi/chi/v5"
	datastar "github.com/starfederation/datastar/sdk/go"
)

type searchSignal struct {
	Search string `json:"search"`
}

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
		sse.MergeFragmentTempl(components.Cell(cellId))
		sse.MarshalAndMergeSignals(map[string]any{
			"cell":       cellValue,
			"keyPressed": "",
		})
	})

	// router.Get("/cell/{cellId}/select", func(w http.ResponseWriter, r *http.Request) {
	// 	cellId := chi.URLParam(r, "cellId")
	// 	// Extract query parameter `value`
	// 	value := r.URL.Query().Get("value")

	// 	sse := datastar.NewSSE(w, r)
	// 	sse.MergeFragmentTempl(components.InactiveCell(cellId, value, true))
	// })

	router.Get("/search", func(w http.ResponseWriter, r *http.Request) {
		// sse := datastar.NewSSE(w, r)
		searchValue := &searchSignal{}
		if err := datastar.ReadSignals(r, searchValue); err != nil {

			http.Error(w, err.Error(), http.StatusBadRequest)

			log.Println("Failed to read signals:", err) // Updated to use log.Println
		}
		fmt.Println(searchValue.Search)
		// sse.MergeFragmentTempl(components.Cell(cellId, true))
	})
	return nil
}
