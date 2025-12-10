module github.com/jefferyjob/go-redislock

go 1.21

retract (
	v1.6.0 // adapter package name is invalid
	v1.5.0 // adapter package name is invalid
	v1.0.0 // package name error
)

require (
	github.com/golang/mock v1.6.0
	github.com/google/uuid v1.6.0
)
