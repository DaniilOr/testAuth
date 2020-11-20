package authenticator

import (
	"bytes"
	"context"
	"github.com/go-chi/chi"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAuthenticatorHTTPMux(t *testing.T) {
	mux := http.NewServeMux()
	authenticatorMd := Authenticator(func(ctx context.Context) (*string, error) {
		log.Println("first")
		id := "192.0.2.1"
		return &id, nil
	}, func(ctx context.Context, id *string) (interface{}, error) {
		return "USERAUTH", nil
	})
	mux.Handle(
		"/get",
		authenticatorMd(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			log.Println("second")
			profile, err := Authentication(request.Context())
			if err != nil {
				t.Fatal(err)
			}
			data := profile.(string)

			_, err = writer.Write([]byte(data))
			if err != nil {
				t.Fatal(err)
			}
		})),
	)

	type args struct {
		method string
		path   string
	}

	tests := []struct {
		name string
		args args
		want []byte
	}{
		{name: "GET", args: args{method: "GET", path: "/get"}, want: []byte("USERAUTH")},
		// TODO: write for other methods
	}

	for _, tt := range tests {
		request := httptest.NewRequest(tt.args.method, tt.args.path, nil)
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, request)
		got := response.Body.Bytes()
		if !bytes.Equal(tt.want, got) {
			t.Errorf("got %s, want %s", got, tt.want)
		}
	}
}

func TestAuthenticatorChi(t *testing.T) {
	router := chi.NewRouter()
	authenticatorMd := Authenticator(func(ctx context.Context) (*string, error) {
		id := "192.0.2.1"
		return &id, nil
	}, func(ctx context.Context, id *string) (interface{}, error) {
		return "USERAUTH", nil
	})
	router.With(authenticatorMd).Get(
		"/get",
		func(writer http.ResponseWriter, request *http.Request) {
			profile, err := Authentication(request.Context())
			if err != nil {
				t.Fatal(err)
			}
			data := profile.(string)

			_, err = writer.Write([]byte(data))
			if err != nil {
				t.Fatal(err)
			}
		},
	)

	type args struct {
		method string
		path   string
	}

	tests := []struct {
		name string
		args args
		want []byte
	}{
		{name: "GET", args: args{method: "GET", path: "/get"}, want: []byte("USERAUTH")},
		// TODO: write for other methods
	}

	for _, tt := range tests {
		request := httptest.NewRequest(tt.args.method, tt.args.path, nil)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		got := response.Body.Bytes()
		if !bytes.Equal(tt.want, got) {
			t.Errorf("got %s, want %s", got, tt.want)
		}
	}
}

func TestAuthenticatorHTTPMuxError(t *testing.T) {
	mux := http.NewServeMux()
	authenticatorMd := Authenticator(func(ctx context.Context) (*string, error) {
		id := "192.0.2.1"
		return &id, nil
	}, func(ctx context.Context, id *string) (interface{}, error) {
		if strings.Compare(*id, "0.0.0.0")==0{
			return "USERAUTH", nil
		}
		return "", ErrNoAuthentication
	})
	mux.Handle(
		"/get",
		authenticatorMd(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			log.Println("second")
			profile, err := Authentication(request.Context())
			if err != nil {
				t.Fatal(err)
			}
			data := profile.(string)

			_, err = writer.Write([]byte(data))
			if err != nil {
				t.Fatal(err)
			}
		})),
	)

	type args struct {
		method string
		path   string
	}

	tests := []struct {
		name string
		args args
		want int
	}{
		{name: "GET", args: args{method: "GET", path: "/get"}, want: http.StatusUnauthorized},
		// TODO: write for other methods
	}

	for _, tt := range tests {
		request := httptest.NewRequest(tt.args.method, tt.args.path, nil)
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, request)
		got := response.Code
		log.Println(got)
		if tt.want != got {
			t.Errorf("got %d, want %d", got, tt.want)
		}
	}
}
