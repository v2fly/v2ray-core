package restful_api

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
	"github.com/v2fly/v2ray-core/v4/common/net"
	"github.com/v2fly/v2ray-core/v4/transport/internet"

	"net/http"
	"strings"
)

func JSONResponse(w http.ResponseWriter, data interface{}, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(data)
}

var validate *validator.Validate

type StatsUser struct {
	uuid  string `validate:"required_without=email,uuid4"`
	email string `validate:"required_without=uuid,email"`
}

type StatsUserResponse struct {
	Uplink   int64 `json:"uplink"`
	Downlink int64 `json:"downlink"`
}

func (rs *restfulService) statsUser(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	statsUser := &StatsUser{
		uuid:  query.Get("uuid"),
		email: query.Get("email"),
	}

	if err := validate.Struct(statsUser); err != nil {
		JSONResponse(w, http.StatusText(422), 422)
	}

	response := &StatsUserResponse{
		Uplink:   0,
		Downlink: 0,
	}

	JSONResponse(w, response, 200)
}

type Stats struct {
	tag string `validate:"required,alpha,min=1,max=255"`
}

type StatsBound struct { // Better name?
	Uplink   int64 `json:"uplink"`
	Downlink int64 `json:"downlink"`
}

type StatsResponse struct {
	Inbound  StatsBound `json:"inbound"`
	Outbound StatsBound `json:"outbound"`
}

func (rs *restfulService) statsRequest(w http.ResponseWriter, r *http.Request) {
	stats := &Stats{
		tag: r.URL.Query().Get("tag"),
	}
	if err := validate.Struct(stats); err != nil {
		JSONResponse(w, http.StatusText(422), 422)
	}

	response := StatsResponse{
		Inbound: StatsBound{
			Uplink:   1,
			Downlink: 1,
		},
		Outbound: StatsBound{
			Uplink:   1,
			Downlink: 1,
		}}

	JSONResponse(w, response, 200)
}

func (rs *restfulService) TokenAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		const prefix = "Bearer "
		if !strings.HasPrefix(auth, prefix) {
			JSONResponse(w, http.StatusText(403), 403)
			return
		}
		auth = strings.TrimPrefix(auth, prefix)
		if auth != rs.config.AuthToken {
			JSONResponse(w, http.StatusText(403), 403)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (rs *restfulService) start() error {
	r := chi.NewRouter()
	r.Use(rs.TokenAuthMiddleware)
	r.Use(middleware.Heartbeat("/ping"))

	r.Route("/v1", func(r chi.Router) {
		r.Get("/stats/user", rs.statsUser)
		r.Get("/stats", rs.statsRequest)
	})

	var listener net.Listener
	var err error
	address := net.ParseAddress(rs.config.ListenAddr)

	switch {
	case address.Family().IsIP():
		listener, err = internet.ListenSystem(rs.ctx, &net.TCPAddr{IP: address.IP(), Port: int(rs.config.ListenPort)}, nil)
	case strings.EqualFold(address.Domain(), "localhost"):
		listener, err = internet.ListenSystem(rs.ctx, &net.TCPAddr{IP: net.IP{127, 0, 0, 1}, Port: int(rs.config.ListenPort)}, nil)
	default:
		return newError("restful api cannot listen on the address: ", address)
	}
	if err != nil {
		return newError("restful api cannot listen on the port ", rs.config.ListenPort).Base(err)
	}

	go func() {
		err := http.Serve(listener, r)
		if err != nil {
			newError("unable to serve restful api").WriteToLog()
		}
	}()

	return nil
}
