package shared

type Result[T any] struct {
	isSuccess bool
	value     T
}

func Ok[T any](data T) Result[T] {
	return Result[T]{isSuccess: true, value: data}
}
func Fail[T any](reason T) Result[T] {
	return Result[T]{isSuccess: false, value: reason}
}

func (r *Result[T]) IsSuccess() bool {
	return r.isSuccess == true
}
func (r *Result[T]) Value() T {
	return r.value
}
