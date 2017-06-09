// Forked from https://github.com/julienschmidt/go-http-routing-benchmark
//
package bench

import (
	"io"
	"net/http"
	"regexp"
	"testing"

	"context"

	"github.com/julienschmidt/httprouter"
	"github.com/rs/xhandler"
	"github.com/rs/xmux"
	goji "github.com/zenazn/goji/web"
)

var benchRe *regexp.Regexp

type route struct {
	method string
	path   string
}

type mockResponseWriter struct{}

func (m *mockResponseWriter) Header() (h http.Header) {
	return http.Header{}
}

func (m *mockResponseWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (m *mockResponseWriter) WriteString(s string) (n int, err error) {
	return len(s), nil
}

func (m *mockResponseWriter) WriteHeader(int) {}

type testHandler struct{}

func (n testHandler) ServeHTTPC(ctx context.Context, w http.ResponseWriter, r *http.Request) {}

var httpHandlerC = xhandler.HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {})

var xhandlerWrite = xhandler.HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, xmux.Params(ctx).Get("name"))
})

func loadXmux(routes []route) xhandler.HandlerC {
	h := testHandler{}
	mux := xmux.New()
	for _, route := range routes {
		mux.HandleC(route.method, route.path, h)
	}
	return mux
}

func loadXmuxSingle(method, path string, h xhandler.HandlerC) xhandler.HandlerC {
	mux := xmux.New()
	mux.HandleC(method, path, h)
	return mux
}

func httpHandlerFunc(w http.ResponseWriter, r *http.Request) {}

func gojiFuncWrite(c goji.C, w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, c.URLParams["name"])
}

func loadGoji(routes []route) http.Handler {
	h := httpHandlerFunc

	mux := goji.New()
	for _, route := range routes {
		switch route.method {
		case "GET":
			mux.Get(route.path, h)
		case "POST":
			mux.Post(route.path, h)
		case "PUT":
			mux.Put(route.path, h)
		case "PATCH":
			mux.Patch(route.path, h)
		case "DELETE":
			mux.Delete(route.path, h)
		default:
			panic("Unknown HTTP method: " + route.method)
		}
	}
	return mux
}

func loadGojiSingle(method, path string, handler interface{}) http.Handler {
	mux := goji.New()
	switch method {
	case "GET":
		mux.Get(path, handler)
	case "POST":
		mux.Post(path, handler)
	case "PUT":
		mux.Put(path, handler)
	case "PATCH":
		mux.Patch(path, handler)
	case "DELETE":
		mux.Delete(path, handler)
	default:
		panic("Unknow HTTP method: " + method)
	}
	return mux
}

func httpRouterHandle(_ http.ResponseWriter, _ *http.Request, _ httprouter.Params) {}

func httpRouterHandleWrite(w http.ResponseWriter, _ *http.Request, ps httprouter.Params) {
	io.WriteString(w, ps.ByName("name"))
}

func loadHTTPRouter(routes []route) http.Handler {
	h := httpRouterHandle

	router := httprouter.New()
	for _, route := range routes {
		router.Handle(route.method, route.path, h)
	}
	return router
}

func loadHTTPRouterSingle(method, path string, handle httprouter.Handle) http.Handler {
	router := httprouter.New()
	router.Handle(method, path, handle)
	return router
}

func benchRequest(b *testing.B, router http.Handler, r *http.Request) {
	w := new(mockResponseWriter)
	u := r.URL
	rq := u.RawQuery
	r.RequestURI = u.RequestURI()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		u.RawQuery = rq
		router.ServeHTTP(w, r)
	}
}

func benchRequestC(b *testing.B, router xhandler.HandlerC, ctx context.Context, r *http.Request) {
	w := new(mockResponseWriter)
	u := r.URL
	rq := u.RawQuery
	r.RequestURI = u.RequestURI()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		u.RawQuery = rq
		router.ServeHTTPC(ctx, w, r)
	}
}

func benchRoutes(b *testing.B, router http.Handler, routes []route) {
	w := new(mockResponseWriter)
	r, _ := http.NewRequest("GET", "/", nil)
	u := r.URL
	rq := u.RawQuery

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, route := range routes {
			r.Method = route.method
			r.RequestURI = route.path
			u.Path = route.path
			u.RawQuery = rq
			router.ServeHTTP(w, r)
		}
	}
}

func benchRoutesC(b *testing.B, router xhandler.HandlerC, ctx context.Context, routes []route) {
	w := new(mockResponseWriter)
	r, _ := http.NewRequest("GET", "/", nil)
	u := r.URL
	rq := u.RawQuery

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, route := range routes {
			r.Method = route.method
			r.RequestURI = route.path
			u.Path = route.path
			u.RawQuery = rq
			router.ServeHTTPC(ctx, w, r)
		}
	}
}

// Micro Benchmarks

// Route with Param (no write)
func BenchmarkXmux_Param1(b *testing.B) {
	router := loadXmuxSingle("GET", "/user/:name", httpHandlerC)

	r, _ := http.NewRequest("GET", "/user/gordon", nil)
	benchRequestC(b, router, context.Background(), r)
}
func BenchmarkGoji_Param1(b *testing.B) {
	router := loadGojiSingle("GET", "/user/:name", httpHandlerFunc)

	r, _ := http.NewRequest("GET", "/user/gordon", nil)
	benchRequest(b, router, r)
}
func BenchmarkHTTPRouter_Param1(b *testing.B) {
	router := loadHTTPRouterSingle("GET", "/user/:name", httpRouterHandle)

	r, _ := http.NewRequest("GET", "/user/gordon", nil)
	benchRequest(b, router, r)
}

// Route with 5 Params (no write)
const fiveColon = "/:a/:b/:c/:d/:e"
const fiveRoute = "/test/test/test/test/test"

func BenchmarkXmux_Param5(b *testing.B) {
	router := loadXmuxSingle("GET", fiveColon, httpHandlerC)

	r, _ := http.NewRequest("GET", fiveRoute, nil)
	benchRequestC(b, router, context.Background(), r)
}
func BenchmarkGoji_Param5(b *testing.B) {
	router := loadGojiSingle("GET", fiveColon, httpHandlerFunc)

	r, _ := http.NewRequest("GET", fiveRoute, nil)
	benchRequest(b, router, r)
}
func BenchmarkHTTPRouter_Param5(b *testing.B) {
	router := loadHTTPRouterSingle("GET", fiveColon, httpRouterHandle)

	r, _ := http.NewRequest("GET", fiveRoute, nil)
	benchRequest(b, router, r)
}

// Route with 20 Params (no write)
const twentyColon = "/:a/:b/:c/:d/:e/:f/:g/:h/:i/:j/:k/:l/:m/:n/:o/:p/:q/:r/:s/:t"
const twentyRoute = "/a/b/c/d/e/f/g/h/i/j/k/l/m/n/o/p/q/r/s/t"

func BenchmarkXmux_Param20(b *testing.B) {
	router := loadXmuxSingle("GET", twentyColon, httpHandlerC)

	r, _ := http.NewRequest("GET", twentyRoute, nil)
	benchRequestC(b, router, context.Background(), r)
}
func BenchmarkGoji_Param20(b *testing.B) {
	router := loadGojiSingle("GET", twentyColon, httpHandlerFunc)

	r, _ := http.NewRequest("GET", twentyRoute, nil)
	benchRequest(b, router, r)
}
func BenchmarkHTTPRouter_Param20(b *testing.B) {
	router := loadHTTPRouterSingle("GET", twentyColon, httpRouterHandle)

	r, _ := http.NewRequest("GET", twentyRoute, nil)
	benchRequest(b, router, r)
}

// Route with Param and write
func BenchmarkXmux_ParamWrite(b *testing.B) {
	router := loadXmuxSingle("GET", "/user/:name", xhandlerWrite)

	r, _ := http.NewRequest("GET", "/user/gordon", nil)
	benchRequestC(b, router, context.Background(), r)
}
func BenchmarkGoji_ParamWrite(b *testing.B) {
	router := loadGojiSingle("GET", "/user/:name", gojiFuncWrite)

	r, _ := http.NewRequest("GET", "/user/gordon", nil)
	benchRequest(b, router, r)
}
func BenchmarkHTTPRouter_ParamWrite(b *testing.B) {
	router := loadHTTPRouterSingle("GET", "/user/:name", httpRouterHandleWrite)

	r, _ := http.NewRequest("GET", "/user/gordon", nil)
	benchRequest(b, router, r)
}
