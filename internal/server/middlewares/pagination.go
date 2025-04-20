package middlewares

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/goodieshq/sweettooth/internal/server/requests"
	"github.com/goodieshq/sweettooth/internal/server/responses"
	"github.com/goodieshq/sweettooth/pkg/api"
)

func MiddlewarePaginate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// default pagination
		pagination := api.DefaultPagination()

		limitString := r.URL.Query().Get("limit")
		if limitString != "" {
			limit, err := strconv.Atoi(limitString)
			if err != nil {
				responses.ErrInvalidPagination(w, r, err)
				return
			}
			pagination.Limit = api.PaginationLimit(limit)
		}

		offsetString := r.URL.Query().Get("offset")
		if offsetString != "" {
			offset, err := strconv.Atoi(offsetString)
			if err != nil {
				responses.ErrInvalidPagination(w, r, err)
				return
			}
			pagination.Offset = offset
		}

		sort := r.URL.Query().Get("sort")
		if sort != "" {
			switch strings.ToUpper(sort) {
			case api.PAGINATION_SORT_ASC:
				pagination.Sort = api.PAGINATION_SORT_ASC
			case api.PAGINATION_SORT_DESC:
				pagination.Sort = api.PAGINATION_SORT_DESC
			default:
				responses.ErrInvalidPagination(w, r, nil)
				return
			}
		}

		// set the page and limit for the request within the context
		r = requests.WithRequestPagination(r, pagination)

		next.ServeHTTP(w, r)
	})

}
