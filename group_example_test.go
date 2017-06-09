package xmux_test

import (
	"fmt"
	"log"
	"net/http"

	"context"

	"github.com/rs/xhandler"
	"github.com/rs/xmux"
)

func ExampleMux_NewGroup() {
	mux := xmux.New()

	api := mux.NewGroup("/api")

	api.GET("/users/:name", xhandler.HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "GET /api/users/%s", xmux.Param(ctx, "name"))
	}))

	api.POST("/users/:name", xhandler.HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "POST /api/users/%s", xmux.Param(ctx, "name"))
	}))

	if err := http.ListenAndServe(":8080", xhandler.New(context.Background(), mux)); err != nil {
		log.Fatal(err)
	}
}
