package main

import (
	"net/http"

	_ "github.com/Mirnda/mirandaclin/docs" // registra a spec gerada pelo swag init
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func registerSwagger(mux *http.ServeMux) {
	mux.Handle("GET /swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))
}
