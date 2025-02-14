package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"myproject/web/components"
	"myproject/web/pages"
	"net/http"
	"time"

	"github.com/delaneyj/toolbelt"
	"github.com/delaneyj/toolbelt/embeddednats"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/sessions"
	"github.com/nats-io/nats.go/jetstream"
	datastar "github.com/starfederation/datastar/sdk/go"
)

func setupHomeRoute(router chi.Router, store sessions.Store, ns *embeddednats.Server) error {
	nc, err := ns.Client()
	if err != nil {
		return fmt.Errorf("error creating nats client: %w", err)
	}

	js, err := jetstream.New(nc)
	if err != nil {
		return fmt.Errorf("error creating jetstream client: %w", err)
	}

	kv, err := js.CreateOrUpdateKeyValue(context.Background(), jetstream.KeyValueConfig{
		Bucket:      "sheet",
		Description: "Sheet data",
		Compression: true,
		TTL:         time.Hour,
		MaxBytes:    16 * 1024 * 1024,
	})

	if err != nil {
		return fmt.Errorf("error creating key value: %w", err)
	}

	saveSheet := func(ctx context.Context, sessionID string, sd *components.SheetData) error {
		b, err := json.Marshal(sd)
		if err != nil {
			return fmt.Errorf("failed to marshal mvc: %w", err)
		}
		if _, err := kv.Put(ctx, sessionID, b); err != nil {
			return fmt.Errorf("failed to put key value: %w", err)
		}
		return nil
	}

	newSheet := func(sd *components.SheetData) {
		sd.Title = "My Sheet"
		sd.Content = "Hello, world!"

	}

	sheetSession := func(w http.ResponseWriter, r *http.Request) (string, *components.SheetData, error) {
		ctx := r.Context()
		sessionID, err := upsertSessionID(store, r, w)
		if err != nil {
			return "", nil, fmt.Errorf("failed to get session id: %w", err)
		}

		sd := &components.SheetData{}
		if entry, err := kv.Get(ctx, sessionID); err != nil {
			if err != jetstream.ErrKeyNotFound {
				return "", nil, fmt.Errorf("failed to get key value: %w", err)
			}
			newSheet(sd)

			if err := saveSheet(ctx, sessionID, sd); err != nil {
				return "", nil, fmt.Errorf("failed to save mvc: %w", err)
			}
		} else {
			if err := json.Unmarshal(entry.Value(), sd); err != nil {
				return "", nil, fmt.Errorf("failed to unmarshal mvc: %w", err)
			}
		}
		return sessionID, sd, nil
	}

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		pages.Index("best title").Render(r.Context(), w)
	})

	router.Route("/api", func(apiRouter chi.Router) {

		apiRouter.Route("/sheet", func(sheetRouter chi.Router) {

			sheetRouter.Get("/", func(w http.ResponseWriter, r *http.Request) {
				sessionID, sd, err := sheetSession(w, r)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				sse := datastar.NewSSE(w, r)

				// Watch for updates
				ctx := r.Context()
				watcher, err := kv.Watch(ctx, sessionID)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				defer watcher.Stop()
				for {
					select {
					case <-ctx.Done():
						return
					case entry := <-watcher.Updates():
						if entry == nil {
							continue
						}
						if err := json.Unmarshal(entry.Value(), sd); err != nil {
							http.Error(w, err.Error(), http.StatusInternalServerError)
							return
						}
						c := components.SheetView(sd)
						if err := sse.MergeFragmentTempl(c); err != nil {
							sse.ConsoleError(err)
							return
						}
					}
				}
			})

			sheetRouter.Put("/", func(w http.ResponseWriter, r *http.Request) {
				type Signals struct {
					Title string `json:"title"`
				}
				signals := &Signals{}

				if err := datastar.ReadSignals(r, signals); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}

				sessionID, sd, err := sheetSession(w, r)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				sd.Title = signals.Title
				saveSheet(r.Context(), sessionID, sd)
			})
		})
	})

	return nil
}

func upsertSessionID(store sessions.Store, r *http.Request, w http.ResponseWriter) (string, error) {

	sess, err := store.Get(r, "connections")
	if err != nil {
		return "", fmt.Errorf("failed to get session: %w", err)
	}
	id, ok := sess.Values["id"].(string)
	if !ok {
		id = toolbelt.NextEncodedID()
		sess.Values["id"] = id
		if err := sess.Save(r, w); err != nil {
			return "", fmt.Errorf("failed to save session: %w", err)
		}
	}
	return id, nil
}
