package utils

type IntervalTypes int

const (
	Weeks  IntervalTypes = iota // iota starts at 0
	Months                      // implicitly Weeks + 1
)
