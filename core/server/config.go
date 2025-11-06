package server

// DefaultConfig is the default configuration for the [Server].
var DefaultConfig = Config{
	Port: 8080,
	Path: "/",
	CORS: CORS{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "Authorization", "Connect-Protocol-Version", "Connect-Timeout-Ms"},
	},
}

// Config represents the configuration for the [Server].
type Config struct {
	Port uint16
	Path string
	CORS CORS
}

// CORS represents the CORS configuration for the server.
type CORS struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
}
