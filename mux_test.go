package xmux

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/xhandler"
	"github.com/stretchr/testify/assert"

	"golang.org/x/net/context"
)

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

func TestParams(t *testing.T) {
	ps := ParamHolder{
		{"param1", "value1"},
		{"param2", "value2"},
		{"param3", "value3"},
	}
	for i := range ps {
		assert.Equal(t, ps[i].Value, ps.Get(ps[i].Name))
	}
	assert.Equal(t, "", ps.Get("noKey"), "Expected empty string for not found")
}

func TestParamsDup(t *testing.T) {
	ps := ParamHolder{
		{"param", "value1"},
		{"param", "value2"},
	}
	assert.Equal(t, "value1", ps.Get("param"))
}

func TestCtxParams(t *testing.T) {
	ps := ParamHolder{{"param1", "value1"}}
	ctx := newParamContext(context.TODO(), ps)
	assert.Equal(t, "value1", Params(ctx).Get("param1"))
	assert.Equal(t, emptyParams, Params(context.TODO()))
	assert.Equal(t, emptyParams, Params(nil))
}

func TestCtxParam(t *testing.T) {
	ps := ParamHolder{{"param1", "value1"}}
	ctx := newParamContext(context.TODO(), ps)
	assert.Equal(t, "value1", Param(ctx, "param1"))
	assert.Equal(t, "", Param(context.TODO(), "param1"))
	assert.Equal(t, "", Param(nil, "param1"))
}

func TestMux(t *testing.T) {
	mux := New()

	routed := false
	mux.HandleC("GET", "/user/:name", xhandler.HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		routed = true
		assert.Equal(t, ParamHolder{{"name", "gopher"}}, Params(ctx))
	}))

	w := new(mockResponseWriter)
	r, _ := http.NewRequest("GET", "/user/gopher", nil)
	mux.ServeHTTPC(context.Background(), w, r)
	assert.True(t, routed, "routing failed")
}

func TestMuxAdaptors(t *testing.T) {
	mux := New()

	var handle, handleFunc, handleFuncC bool
	mux.Handle("GET", "/handle", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handle = true
	}))
	mux.HandleFunc("GET", "/handleFunc", func(w http.ResponseWriter, r *http.Request) {
		handleFunc = true
	})
	mux.HandleFuncC("GET", "/handleFuncC", func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		handleFuncC = true
	})

	w := new(mockResponseWriter)
	r, _ := http.NewRequest("GET", "/handle", nil)
	mux.ServeHTTPC(context.Background(), w, r)
	assert.True(t, handle, "routing failed")

	w = new(mockResponseWriter)
	r, _ = http.NewRequest("GET", "/handleFunc", nil)
	mux.ServeHTTPC(context.Background(), w, r)
	assert.True(t, handleFunc, "routing failed")

	w = new(mockResponseWriter)
	r, _ = http.NewRequest("GET", "/handleFuncC", nil)
	mux.ServeHTTPC(context.Background(), w, r)
	assert.True(t, handleFuncC, "routing failed")
}

type handlerStruct struct {
	handeled *bool
}

func (h handlerStruct) ServeHTTPC(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	*h.handeled = true
}

func TestMuxAPI(t *testing.T) {
	var get, head, options, post, put, patch, delete bool

	mux := New()
	mux.GET("/GET", xhandler.HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		get = true
	}))
	mux.HEAD("/GET", xhandler.HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		head = true
	}))
	mux.OPTIONS("/GET", xhandler.HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		options = true
	}))
	mux.POST("/POST", xhandler.HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		post = true
	}))
	mux.PUT("/PUT", xhandler.HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		put = true
	}))
	mux.PATCH("/PATCH", xhandler.HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		patch = true
	}))
	mux.DELETE("/DELETE", xhandler.HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		delete = true
	}))

	w := new(mockResponseWriter)

	r, _ := http.NewRequest("GET", "/GET", nil)
	mux.ServeHTTPC(context.Background(), w, r)
	assert.NotNil(t, get, "routing GET failed")

	r, _ = http.NewRequest("HEAD", "/GET", nil)
	mux.ServeHTTPC(context.Background(), w, r)
	assert.NotNil(t, head, "routing HEAD failed")

	r, _ = http.NewRequest("OPTIONS", "/GET", nil)
	mux.ServeHTTPC(context.Background(), w, r)
	assert.NotNil(t, options, "routing OPTIONS failed")

	r, _ = http.NewRequest("POST", "/POST", nil)
	mux.ServeHTTPC(context.Background(), w, r)
	assert.NotNil(t, post, "routing POST failed")

	r, _ = http.NewRequest("PUT", "/PUT", nil)
	mux.ServeHTTPC(context.Background(), w, r)
	assert.NotNil(t, put, "routing PUT failed")

	r, _ = http.NewRequest("PATCH", "/PATCH", nil)
	mux.ServeHTTPC(context.Background(), w, r)
	assert.NotNil(t, patch, "routing PATCH failed")

	r, _ = http.NewRequest("DELETE", "/DELETE", nil)
	mux.ServeHTTPC(context.Background(), w, r)
	assert.NotNil(t, delete, "routing DELETE failed")
}

func TestMuxRoot(t *testing.T) {
	mux := New()
	recv := catchPanic(func() {
		mux.GET("noSlashRoot", nil)
	})
	assert.NotNil(t, recv, "registering path not beginning with '/' did not panic")
}

func TestMuxChaining(t *testing.T) {
	mux1 := New()
	mux2 := New()
	mux1.NotFound = mux2

	fooHit := false
	mux1.POST("/foo", xhandler.HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, req *http.Request) {
		fooHit = true
		w.WriteHeader(http.StatusOK)
	}))

	barHit := false
	mux2.POST("/bar", xhandler.HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, req *http.Request) {
		barHit = true
		w.WriteHeader(http.StatusOK)
	}))

	r, _ := http.NewRequest("POST", "/foo", nil)
	w := httptest.NewRecorder()
	mux1.ServeHTTPC(context.Background(), w, r)
	if !(w.Code == http.StatusOK && fooHit) {
		t.Errorf("Regular routing failed with router chaining.")
		t.FailNow()
	}

	r, _ = http.NewRequest("POST", "/bar", nil)
	w = httptest.NewRecorder()
	mux1.ServeHTTPC(context.Background(), w, r)
	if !(w.Code == http.StatusOK && barHit) {
		t.Errorf("Chained routing failed with router chaining.")
		t.FailNow()
	}

	r, _ = http.NewRequest("POST", "/qax", nil)
	w = httptest.NewRecorder()
	mux1.ServeHTTPC(context.Background(), w, r)
	if !(w.Code == http.StatusNotFound) {
		t.Errorf("NotFound behavior failed with router chaining.")
		t.FailNow()
	}
}

func TestMuxNotAllowed(t *testing.T) {
	handlerFunc := xhandler.HandlerFuncC(func(_ context.Context, _ http.ResponseWriter, _ *http.Request) {})

	mux := New()
	mux.POST("/path", handlerFunc)
	mux.PUT("/path", handlerFunc)

	// Test not allowed
	r, _ := http.NewRequest("GET", "/path", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTPC(context.Background(), w, r)
	assert.Equal(t, w.Code, http.StatusMethodNotAllowed, "NotAllowed handling failed")
	assert.Equal(t, "POST, PUT", w.HeaderMap.Get("Allow"))

	w = httptest.NewRecorder()
	responseText := "custom method"
	mux.MethodNotAllowed = xhandler.HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		w.Write([]byte(responseText))
	})
	mux.ServeHTTPC(context.Background(), w, r)
	assert.Equal(t, responseText, w.Body.String())
	assert.Equal(t, w.Code, http.StatusTeapot)
}

func TestMuxNotAllowedSkipOptions(t *testing.T) {
	handlerFunc := xhandler.HandlerFuncC(func(_ context.Context, _ http.ResponseWriter, _ *http.Request) {})

	mux := New()
	mux.OPTIONS("/path", handlerFunc)

	// Test not allowed
	r, _ := http.NewRequest("GET", "/path", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTPC(context.Background(), w, r)
	assert.Equal(t, w.Code, http.StatusNotFound, "Not allowed response when only OPTIONS is allowed")
}

func TestMuxNotFound(t *testing.T) {
	handlerFunc := xhandler.HandlerFuncC(func(_ context.Context, _ http.ResponseWriter, _ *http.Request) {})

	mux := New()
	mux.GET("/path", handlerFunc)
	mux.GET("/dir/", handlerFunc)
	mux.GET("/", handlerFunc)

	testRoutes := []struct {
		route  string
		code   int
		header string
	}{
		{"/path/", 301, "map[Location:[/path]]"},   // TSR -/
		{"/dir", 301, "map[Location:[/dir/]]"},     // TSR +/
		{"", 301, "map[Location:[/]]"},             // TSR +/
		{"/PATH", 301, "map[Location:[/path]]"},    // Fixed Case
		{"/DIR/", 301, "map[Location:[/dir/]]"},    // Fixed Case
		{"/PATH/", 301, "map[Location:[/path]]"},   // Fixed Case -/
		{"/DIR", 301, "map[Location:[/dir/]]"},     // Fixed Case +/
		{"/../path", 301, "map[Location:[/path]]"}, // CleanPath
		{"/nope", 404, ""},                         // NotFound
	}
	for _, tr := range testRoutes {
		r, _ := http.NewRequest("GET", tr.route, nil)
		w := httptest.NewRecorder()
		mux.ServeHTTPC(context.Background(), w, r)
		if !(w.Code == tr.code && (w.Code == 404 || fmt.Sprint(w.Header()) == tr.header)) {
			t.Errorf("NotFound handling route %s failed: Code=%d, Header=%v", tr.route, w.Code, w.Header())
		}
	}

	// Test custom not found handler
	var notFound bool
	mux.NotFound = xhandler.HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		notFound = true
	})
	r, _ := http.NewRequest("GET", "/nope", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTPC(context.Background(), w, r)
	if !(w.Code == 404 && notFound == true) {
		t.Errorf("Custom NotFound handler failed: Code=%d, Header=%v", w.Code, w.Header())
	}

	// Test other method than GET (want 307 instead of 301)
	mux.PATCH("/path", handlerFunc)
	r, _ = http.NewRequest("PATCH", "/path/", nil)
	w = httptest.NewRecorder()
	mux.ServeHTTPC(context.Background(), w, r)
	if !(w.Code == 307 && fmt.Sprint(w.Header()) == "map[Location:[/path]]") {
		t.Errorf("Custom NotFound handler failed: Code=%d, Header=%v", w.Code, w.Header())
	}

	// Test special case where no node for the prefix "/" exists
	mux = New()
	mux.GET("/a", handlerFunc)
	r, _ = http.NewRequest("GET", "/", nil)
	w = httptest.NewRecorder()
	mux.ServeHTTPC(context.Background(), w, r)
	assert.Equal(t, 404, w.Code)
}

func TestMuxPanicHandler(t *testing.T) {
	mux := New()
	panicHandled := false

	mux.PanicHandler = func(ctx context.Context, w http.ResponseWriter, r *http.Request, p interface{}) {
		panicHandled = true
	}

	mux.HandleC("PUT", "/user/:name", xhandler.HandlerFuncC(func(_ context.Context, _ http.ResponseWriter, _ *http.Request) {
		panic("oops!")
	}))

	w := new(mockResponseWriter)
	req, _ := http.NewRequest("PUT", "/user/gopher", nil)

	defer func() {
		if rcv := recover(); rcv != nil {
			t.Fatal("handling panic failed")
		}
	}()

	mux.ServeHTTPC(context.Background(), w, req)

	assert.True(t, panicHandled, "simulating failed")
}

func TestMuxLookup(t *testing.T) {
	routed := false
	wantHandler := xhandler.HandlerFuncC(func(_ context.Context, _ http.ResponseWriter, _ *http.Request) {
		routed = true
	})

	mux := New()

	// try empty router first
	handler, _, tsr := mux.Lookup("GET", "/nope")
	assert.Nil(t, handler, "Got handle for unregistered pattern: %v", handler)
	assert.False(t, tsr, "Got wrong TSR recommendation!")

	// insert route and try again
	mux.GET("/user/:name", wantHandler)

	handler, params, tsr := mux.Lookup("GET", "/user/gopher")
	if assert.NotNil(t, handler) {
		handler.ServeHTTPC(nil, nil, nil)
		assert.True(t, routed, "Routing failed!")
	}

	assert.Equal(t, newParams("name", "gopher"), params)

	handler, _, tsr = mux.Lookup("GET", "/user/gopher/")
	assert.Nil(t, handler, "Got handle for unregistered pattern: %v", handler)
	assert.True(t, tsr, "Got no TSR recommendation!")

	handler, _, tsr = mux.Lookup("GET", "/nope")
	assert.Nil(t, handler, "Got handle for unregistered pattern: %v", handler)
	assert.False(t, tsr, "Got wrong TSR recommendation!")
}
