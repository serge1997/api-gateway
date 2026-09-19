package router

import (
	"context"
	"fmt"
	"maps"
	"net/http"
	"slices"
	"time"

	"github.com/serge1997/apigateway/internal/api"
	apigateway "github.com/serge1997/apigateway/internal/apiGateway"
	"github.com/serge1997/apigateway/internal/database"
	"github.com/serge1997/apigateway/internal/reporitory"
	"github.com/serge1997/apigateway/shared"
	httpresponse "github.com/serge1997/apigateway/shared/httpResponse"
)

var routes = []Route{
	{
		path:    "/",
		methods: []string{http.MethodGet, http.MethodDelete, http.MethodPut, http.MethodPatch, http.MethodPost},
		hanlder: api.ServicesHandler,
	},
	{
		path:    "/health",
		methods: []string{http.MethodGet},
		hanlder: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("200 OK - healthy\n"))
		},
	},
	{
		path:    "/api/services",
		methods: []string{http.MethodGet},
		hanlder: func(w http.ResponseWriter, r *http.Request) {
			shared.AllowOrigin(w)
			serviceCollection := slices.Collect(maps.Values(apigateway.Services()))
			response := shared.HttpResponse{Data: serviceCollection, Message: "todos os serviços", Status: 200}.Json()
			fmt.Fprint(w, response)
		},
	},
	{
		path:    "/api/summaries",
		methods: []string{http.MethodGet},
		hanlder: func(w http.ResponseWriter, r *http.Request) {
			shared.AllowOrigin(w)
			ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
			defer cancel()
			repo := reporitory.New(database.Db())
			latencies, err := repo.AvgLatency(ctx)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			errRates, err := repo.ErrRate(ctx)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			reqInLastMinute, err := repo.ReqInMinute(ctx, 1)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			reqInLastTenMinutes, err := repo.ReqInMinute(ctx, 10)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			reqInCountByMinutes, err := repo.ReqInMinuteGroupedByMinute(ctx, 10)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			fmt.Fprint(w, httpresponse.SuccessResponse(map[string]interface{}{
				"latencies":                       latencies,
				"errRates":                        errRates,
				"reqInLastMin":                    reqInLastMinute,
				"reqInLastTenMin":                 reqInLastTenMinutes,
				"reqLastTenMinutesCountByMinutes": reqInCountByMinutes,
			}, http.StatusOK, "").Json())
			return
		},
	},
	{
		path:    "/api/metrics-of-service",
		methods: []string{http.MethodGet},
		hanlder: func(w http.ResponseWriter, r *http.Request) {
			shared.AllowOrigin(w)
			ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
			defer cancel()
			serviceName := r.URL.Query().Get("service")
			repo := reporitory.New(database.Db())
			httpStatusMetrics, err := repo.HttpStatusMetricsOfService(ctx, serviceName)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			httpStatusMethodsMetrics, err := repo.HttpMethodsMetricsOfService(ctx, serviceName)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			requestLast24Hours, err := repo.ReqInMinuteGroupedByMinuteOfService(ctx, 60, serviceName)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			reqLogs, err := repo.AllInLastHourOfService(ctx, serviceName)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			durationMetrics, err := repo.DurationMetricsOfService(ctx, serviceName)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			fmt.Fprint(w, httpresponse.SuccessResponse(map[string]interface{}{
				"statusMetrics":       httpStatusMetrics,
				"methodsMetrics":      httpStatusMethodsMetrics,
				"requestLastRequests": requestLast24Hours,
				"logs":                reqLogs,
				"durationMetrics":     durationMetrics,
			}, http.StatusOK, "").Json())
			return
		},
	},
}
