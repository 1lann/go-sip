package sipnet

const (
	StatusTrying               = 100
	StatusRinging              = 180
	StatusCallIsBeingForwarded = 181
	StatusQueued               = 182
	StatusSessionProgress      = 183

	StatusOK = 200

	StatusMultipleChoices    = 300
	StatusMovedPermanently   = 301
	StatusMovedTemporarily   = 302
	StatusUseProxy           = 305
	StatusAlternativeService = 380

	StatusBadRequest                  = 400
	StatusUnauthorized                = 401
	StatusPaymentRequired             = 402
	StatusForbidden                   = 403
	StatusNotFound                    = 404
	StatusMethodNotAllowed            = 405
	StatusNotAcceptable               = 406
	StatusProxyAuthenticationRequired = 407
	StatusRequestTimeout              = 408
	StatusGone                        = 410
	StatusRequestEntityTooLarge       = 413
	StatusRequestURITooLong           = 414
	StatusUnsupportedMediaType        = 415
	StatusUnsupportedURIScheme        = 416
	StatusBadExtension                = 420
	StatusExtensionRequired           = 421
	StatusIntervalTooBrief            = 423
	StatusTemporarilyUnavailable      = 480
	StatusCallTransactionDoesNotExist = 481
	StatusLoopDetected                = 482
	StatusTooManyHops                 = 483
	StatusAddressIncomplete           = 484
	StatusAmbigious                   = 485
	StatusBusyHere                    = 486
	StatusRequestTerminated           = 487
	StatusNotAcceptableHere           = 488
	StatusRequestPending              = 491
	StatusUndecipherable              = 493

	StatusServerInternalError = 500
	StatusNotImplemented      = 501
	StatusBadGateway          = 502
	StatusServiceUnavailable  = 503
	StatusServerTimeout       = 504
	StatusVersionNotSupported = 505
	StatusMessageTooLarge     = 513

	StatusBusyEverywhere       = 600
	StatusDecline              = 603
	StatusDoesNotExistAnywhere = 604
	StatusUnacceptable         = 606
)

var statusTexts = map[int]string{
	StatusTrying:                      "Trying",
	StatusRinging:                     "Ringing",
	StatusCallIsBeingForwarded:        "Call Is Being Forwarded",
	StatusQueued:                      "Queued",
	StatusSessionProgress:             "Session Progress",
	StatusOK:                          "OK",
	StatusMultipleChoices:             "Multiple Choices",
	StatusMovedPermanently:            "Moved Permanently",
	StatusMovedTemporarily:            "Moved Temporarily",
	StatusUseProxy:                    "Use Proxy",
	StatusAlternativeService:          "Alternative Service",
	StatusBadRequest:                  "Bad Request",
	StatusUnauthorized:                "Unauthorized",
	StatusPaymentRequired:             "Payment Required",
	StatusForbidden:                   "Forbidden",
	StatusNotFound:                    "Not Found",
	StatusMethodNotAllowed:            "Method Not Allowed",
	StatusNotAcceptable:               "Not Acceptable",
	StatusProxyAuthenticationRequired: "Proxy Authentication Required",
	StatusRequestTimeout:              "Request Timeout",
	StatusGone:                        "Gone",
	StatusRequestEntityTooLarge:       "Request Entity Too Large",
	StatusRequestURITooLong:           "Request-URI Too Long",
	StatusUnsupportedMediaType:        "Unsupported Media Type",
	StatusUnsupportedURIScheme:        "Unsupported URI Scheme",
	StatusBadExtension:                "Bad Extension",
	StatusExtensionRequired:           "Extension Required",
	StatusIntervalTooBrief:            "Interval Too Brief",
	StatusTemporarilyUnavailable:      "Temporarily Unavailable",
	StatusCallTransactionDoesNotExist: "Call/Transaction Does Not Exist",
	StatusLoopDetected:                "Loop Detected",
	StatusTooManyHops:                 "Too Many Hops",
	StatusAddressIncomplete:           "Address Incomplete",
	StatusAmbigious:                   "Ambiguous",
	StatusBusyHere:                    "Busy Here",
	StatusRequestTerminated:           "Request Terminated",
	StatusNotAcceptableHere:           "Not Acceptable Here",
	StatusRequestPending:              "Request Pending",
	StatusUndecipherable:              "Undecipherable",
	StatusServerInternalError:         "Server Internal Error",
	StatusNotImplemented:              "Not Implemented",
	StatusBadGateway:                  "Bad Gateway",
	StatusServiceUnavailable:          "Service Unavailable",
	StatusServerTimeout:               "Server Timeout",
	StatusVersionNotSupported:         "Version Not Supported",
	StatusMessageTooLarge:             "Message Too Large",
	StatusBusyEverywhere:              "Busy Everywhere",
	StatusDecline:                     "Decline",
	StatusDoesNotExistAnywhere:        "Does Not Exist Anywhere",
	StatusUnacceptable:                "Not Acceptable",
}

// StatusText returns the human readable text representation of a status code.
func StatusText(code int) string {
	return statusTexts[code]
}
