package api

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/felixge/httpsnoop"
	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/api/respond"
	"github.com/kozlov-ma/sesc-backend/auth"
	"github.com/kozlov-ma/sesc-backend/company"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
)

type ctxKey string

var (
	userCtxKey = ctxKey("user")
)

// CurrentUser retrieves the user from the request context if it exists
func CurrentUser(ctx context.Context) (company.User, bool) {
	uv := ctx.Value(userCtxKey)
	u, ok := uv.(company.User)

	return u, ok
}

func (a *API) CurrentUserMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		rec := event.Get(ctx)

		rec.Sub("user").Set("authorized", false)

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			a.writeJSON(ctx, w, respond.WithError(ctx, company.ErrNeedAuthorization))
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			a.writeJSON(
				ctx,
				w,
				respond.WithError(ctx, fmt.Errorf("%w: wrong Authorization header format", auth.ErrInvalidToken)),
			)
			return
		}

		token := authHeader[7:]
		u, err := a.iam.ImWatermelon(ctx, token)
		if err != nil {
			rec.Add(events.Error, err)
			next.ServeHTTP(w, r)
			return
		}

		rec.Sub("user").Set(
			"authorized", true,
			"id", u.ID,
			"role", u.Role,
		)

		ctx = context.WithValue(ctx, userCtxKey, u)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (a *API) EventMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx, rec := event.NewRecord(ctx, "http_request")

		defer func() {
			if r := recover(); r != nil {
				rec.Set("panic", r)
				rec.Set("panic_message", fmt.Sprintf("%v", r))
				a.eventSink.RecordHTTPRequest(ctx, rec)
				panic(r)
			}
		}()

		rec.Set(events.TraceID, uuid.Must(uuid.NewV7()))

		httprec := rec.Sub("http")

		rec.Set(
			"time", time.Now(),
		)

		httprec.Sub("request").Set(
			"method", r.Method,
			"path", r.URL.Path,
			"proto", r.Proto,
			"authorization_header_present", r.Header.Get("Authorization") != "",
			"content_length", r.ContentLength,
			"host", r.Host,
			"form_values", formValues(r.Form),
			"remote_addr", r.RemoteAddr,
			"header", event.Group(
				"content_type", r.Header.Get("Content-Type"),
			),
		)

		m := httpsnoop.CaptureMetrics(next, w, r.WithContext(ctx))

		rec.Set(
			"processing_time", m.Duration,
		)

		httprec.Sub("response").Set(
			"code", m.Code,
			"bytes_written", m.Written,
			"header", event.Group(
				"content_type", w.Header().Get("Content-Type"),
				"access_control_allow_origin", w.Header().Get("Access-Control-Allow-Origin"),
				"access_control_allow_methods", w.Header().Get("Access-Control-Allow-Methods"),
				"access_control_allow_headers", w.Header().Get("Access-Control-Allow-Headers"),
			),
		)

		if m.Code >= http.StatusInternalServerError {
			rec.Set(events.Critical, true)
		}

		a.eventSink.RecordHTTPRequest(ctx, rec)
	})
}

func formValues(vals url.Values) *event.Record {
	const recordValuesPerFormValue = 2
	values := make([]any, 0, len(vals)*recordValuesPerFormValue)
	for key, val := range vals {
		values = append(values, key, strings.Join(val, ","))
	}
	return event.Group(values...)
}
