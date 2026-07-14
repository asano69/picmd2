package serve

import (
	"fmt"
	"github.com/asano69/picmd/internal/assets"
	"github.com/asano69/picmd/internal/config"
	"github.com/asano69/picmd/internal/hooks"
	"io/fs"
	"net/http"
	"strings"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"

	"github.com/sirupsen/logrus"
)

func Run(app *pocketbase.PocketBase, cfg *config.Config) error {
	hooks.RegisterImageCompression(app)
	hooks.RegisterUUIDAssignment(app)
	hooks.RegisterViewCounter(app)

	assetsFS, err := fs.Sub(assets.FS, "assets")
	if err != nil {
		return fmt.Errorf("sub assets fs: %w", err)
	}
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)

	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		e.Router.GET("/assets/{path...}", apis.Static(assetsFS, false))

		serveShell := func(re *core.RequestEvent) error {
			re.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
			http.ServeFileFS(re.Response, re.Request, assets.FS, "index.html")
			return nil
		}

		e.Router.GET("/", serveShell)

		e.Router.GET("/favicon.svg", func(re *core.RequestEvent) error {
			re.Response.Header().Set("Content-Type", "image/svg+xml")
			http.ServeFileFS(re.Response, re.Request, assets.FS, "favicon.svg")
			return nil
		})

		// serveImage redirects to PocketBase's native file URL for the
		// "images" record matching the {uuid} path value. It backs both
		// the "/img/{uuid}" (default) and "/i/{uuid}" (legacy alias)
		// routes below, so this is still the only place that needs to
		// change if storage ever moves away from PocketBase.
		//
		// The path value may come with an arbitrary extension appended
		// (e.g. "...aee41a768eda.webp"), typically added by Markdown
		// viewers or clients that expect image URLs to end in a file
		// extension. UUIDs never contain a dot, so trimming everything
		// from the first dot onward safely recovers the bare UUID.
		serveImage := func(re *core.RequestEvent) error {
			uuid := re.Request.PathValue("uuid")
			if i := strings.IndexByte(uuid, '.'); i != -1 {
				uuid = uuid[:i]
			}

			record, err := app.FindFirstRecordByFilter(
				"images", "uuid = {:uuid}",
				dbx.Params{"uuid": uuid},
			)
			if err != nil {
				return apis.NewNotFoundError("image not found", err)
			}
			target := fmt.Sprintf("/api/files/images/%s/%s", record.Id, record.GetString("image"))
			http.Redirect(re.Response, re.Request, target, http.StatusFound)
			return nil
		}

		// "/img/{uuid}" is the current default public link handed out for
		// pasting into Markdown. "/i/{uuid}" is kept as an alias so
		// links generated before the rename don't break.
		e.Router.GET("/img/{uuid}", serveImage)
		e.Router.GET("/i/{uuid}", serveImage)

		return e.Next()
	})

	logrus.WithField("addr", addr).Info("listening")
	return apis.Serve(app, apis.ServeConfig{
		HttpAddr:        addr,
		ShowStartBanner: false,
	})
}
