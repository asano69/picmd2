package hooks

import (
	"github.com/pocketbase/pocketbase/core"
	"github.com/sirupsen/logrus"
)

// RegisterViewCounter increments the "views" field each time a file from
// the "images" collection is served.
//
// Counting is best-effort and must never block or fail the actual file
// response: the increment runs in a goroutine after the file has already
// been served, and any error here is only logged. The view count is a
// rough indicator, not an exact metric.
func RegisterViewCounter(app core.App) {
	app.OnFileDownloadRequest("images").BindFunc(func(e *core.FileDownloadRequestEvent) error {
		if err := e.Next(); err != nil {
			return err
		}

		recordID := e.Record.Id
		go func() {
			record, err := app.FindRecordById("images", recordID)
			if err != nil {
				logrus.WithError(err).WithField("id", recordID).Warn("images: view count: record not found")
				return
			}
			record.Set("views", record.GetInt("views")+1)
			if err := app.Save(record); err != nil {
				logrus.WithError(err).WithField("id", recordID).Warn("images: view count: save failed")
			}
		}()

		return nil
	})
}
