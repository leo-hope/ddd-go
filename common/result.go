package common

const (
	CodeSuccess     = 200
	CodeParamsError = 400
	CodeSysError    = 500
	CodeTimeout     = 504
	CodeConcurrency = 600
)

// BaseResult is the response envelope returned by every command handler.
type BaseResult struct {
	Code    int
	SubCode string
	Message string
}

func (r *BaseResult) IsSuccess() bool { return r.Code == CodeSuccess }

func SuccessBaseResult() BaseResult {
	return BaseResult{Code: CodeSuccess, Message: "ok"}
}

func FailureBaseResult(code int, msg string) BaseResult {
	return BaseResult{Code: code, Message: msg}
}

func FailureBaseResultFrom(ec ErrorCode) BaseResult {
	return BaseResult{Code: ec.Code(), Message: ec.Msg()}
}

// Result[T] wraps BaseResult with a typed data payload.
type Result[T any] struct {
	BaseResult
	Data T
}

func Success[T any](data T) Result[T] {
	return Result[T]{BaseResult: SuccessBaseResult(), Data: data}
}

func Failure[T any](code int, msg string) Result[T] {
	return Result[T]{BaseResult: FailureBaseResult(code, msg)}
}

func FailureFrom[T any](ec ErrorCode) Result[T] {
	return Result[T]{BaseResult: FailureBaseResultFrom(ec)}
}

// MultiResult[T] carries a slice of items.
type MultiResult[T any] struct {
	BaseResult
	Data []T
}

func (m *MultiResult[T]) IsEmpty() bool { return len(m.Data) == 0 }

func MultiSuccess[T any](data []T) MultiResult[T] {
	return MultiResult[T]{BaseResult: SuccessBaseResult(), Data: data}
}

func MultiFailure[T any](ec ErrorCode) MultiResult[T] {
	return MultiResult[T]{BaseResult: FailureBaseResultFrom(ec)}
}

// PagingResult[T] extends MultiResult with pagination metadata.
type PagingResult[T any] struct {
	MultiResult[T]
	Total   int
	Pages   int
	PageNum int
}

func PagingSuccess[T any](data []T, total, pages, pageNum int) PagingResult[T] {
	return PagingResult[T]{
		MultiResult: MultiSuccess[T](data),
		Total:       total,
		Pages:       pages,
		PageNum:     pageNum,
	}
}

func PagingFailure[T any](ec ErrorCode) PagingResult[T] {
	return PagingResult[T]{MultiResult: MultiResult[T]{BaseResult: FailureBaseResultFrom(ec)}}
}
