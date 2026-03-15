package middleware

func AllowedOrigins(origins []string) []string {
	var valid []string
	for _, o := range origins {
		if o != "" && o != "*" {
			valid = append(valid, o)
		}
	}
	return valid
}

func IsOriginAllowed(origin string, allowed []string) bool {
	for _, a := range allowed {
		if a == origin {
			return true
		}
	}
	return false
}
