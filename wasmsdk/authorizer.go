package main

import "github.com/0chain/gosdk/core/sys"

func RegisterAuthorizer(authorise sys.AuthorizeFunc) {
	sys.Authorize = authorise
}
