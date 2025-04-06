package handler

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"net/http"
	"strings"
)

func AddCustomHeadersAndContextObjects(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		//  only appears on AWS Lambda.
		requestContextRaw, ok := ctx.Value("RequestContext").(events.APIGatewayProxyRequestContext)
		if ok {
			for key, value := range requestContextRaw.Authorizer {
				// make all context objects lowercase because there is no consistency.
				lowerCaseString := strings.ToLower(key)
				ctx = context.WithValue(ctx, lowerCaseString, value) // nolint
			}
		}

		// populate context from headers, but don't override context keys if they are already set. this will prevent tampering.
		for key, value := range r.Header {
			if _, ok = ctx.Value(key).(string); !ok {
				if len(value) == 1 {
					contextValue := value[0]

					// make all context objects lowercase because there is no consistency.
					lowerCaseString := strings.ToLower(key)
					ctx = context.WithValue(ctx, lowerCaseString, contextValue) // nolint
				}
			}
		}

		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

func VerifyMandatoryContextObjects(skipRoutes ...string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			for _, route := range skipRoutes {
				if r.URL.Path == route {
					next.ServeHTTP(w, r)
					return
				}
			}
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}
