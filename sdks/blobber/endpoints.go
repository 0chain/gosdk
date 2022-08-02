package blobber

import (
	"github.com/0chain/common/constants/endpoint"
	"github.com/0chain/common/constants/endpoint/v1_endpoint/blobber_endpoint"
)

var (
	// EndpointWriteMarkerLock api endpoint of WriteMarkerLock
	EndpointWriteMarkerLock = blobber_endpoint.WriteMarkerLock.FormattedPath(endpoint.LeadingAndTrailingSlash)

	// EndpointRootHashnode api endpoint of getting root hashnode of an allocation
	EndpointRootHashnode = blobber_endpoint.HashnodeRoot.FormattedPath(endpoint.LeadingAndTrailingSlash)
)
