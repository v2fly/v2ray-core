package restfulapi

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"

	core "github.com/v2fly/v2ray-core/v5"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
)

var validate *validator.Validate

type StatsBound struct { // Better name?
	Uplink   int64 `json:"uplink"`
	Downlink int64 `json:"downlink"`
}

func (rs *restfulService) tagStats(w http.ResponseWriter, r *http.Request) {
	boundType := chi.URLParam(r, "bound_type")
	tag := chi.URLParam(r, "tag")

	if validate.Var(boundType, "required,oneof=inbounds outbounds") != nil ||
		validate.Var(tag, "required,min=1,max=255") != nil {
		render.Status(r, http.StatusUnprocessableEntity)
		render.JSON(w, r, render.M{})
		return
	}

	bound := boundType[:len(boundType)-1]
	upCounter := rs.stats.GetCounter(bound + ">>>" + tag + ">>>traffic>>>uplink")
	downCounter := rs.stats.GetCounter(bound + ">>>" + tag + ">>>traffic>>>downlink")
	if upCounter == nil || downCounter == nil {
		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, render.M{})
		return
	}

	render.JSON(w, r, &StatsBound{
		Uplink:   upCounter.Value(),
		Downlink: downCounter.Value(),
	})
}

func (rs *restfulService) version(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, render.M{"version": core.Version()})
}

func (rs *restfulService) TokenAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		text := strings.SplitN(header, " ", 2)

		hasInvalidHeader := text[0] != "Bearer"
		hasInvalidSecret := len(text) != 2 || text[1] != rs.config.AuthToken
		if hasInvalidHeader || hasInvalidSecret {
			render.Status(r, http.StatusUnauthorized)
			render.JSON(w, r, render.M{})
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (rs *restfulService) start() error {
	r := chi.NewRouter()
	r.Use(middleware.Heartbeat("/ping"))

	validate = validator.New()
	r.Route("/v1", func(r chi.Router) {
		r.Get("/{bound_type}/{tag}/stats", rs.tagStats)
	})
	r.Get("/version", rs.version)

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
