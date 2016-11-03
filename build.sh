#!/bin/bash -ex
go build s3-versioned-bytes.go
go build lookup-sir.go
go build tail-log-group.go
go build sum-unused-ebs.go
go build s3-versioning.go
