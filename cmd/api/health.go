package main

import (
	"log"
	"net/http"
	"time"
)

// healthcheckHandler godoc
//
//	@Summary		Healthcheck
//	@Description	Healthcheck endpoint
//	@Tags			ops
//	@Produce		json
//	@Success		200	{object}	string	"ok"
//	@Router			/health [get]
func (app *application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"status":  "ok",
		"env":     app.config.env,
		"version": version,
	}
	time.Sleep(time.Second * 3)
	if err := app.writeResponse(w, http.StatusOK, data); err != nil {
		app.internalServerError(w, r, err)
		log.Println(err)
	}
}
