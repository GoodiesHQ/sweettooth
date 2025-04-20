package api

const (
    PAGINATION_LIMIT_MIN = 1
    PAGINATION_LIMIT_MAX = 1000
    PAGINATION_SORT_ASC  = "ASC"
    PAGINATION_SORT_DESC = "DESC"
)

type Pagination struct {
    Limit  int `json:"limit"`
    Offset int `json:"offset"`
    Sort string `json:"sort"`
}

func PaginationLimit(value int) int {
    if value < PAGINATION_LIMIT_MIN {
        return PAGINATION_LIMIT_MIN
    } else if value > PAGINATION_LIMIT_MAX {
        return PAGINATION_LIMIT_MAX
    }
    return value
}

func DefaultPagination() *Pagination {
    return &Pagination{
        Limit:  100,
        Offset: 0,
        Sort:   PAGINATION_SORT_ASC,
    }
}