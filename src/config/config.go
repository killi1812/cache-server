package config

/*
[cache-server]
hostname = cache-server
cache-dir = binary-caches
server-port = 12345
database = dbfile.db
deploy-port = 54321
key = secret
*/

var ConfigPath string

type Config struct {
	Hostname   string
	CacheDir   string
	ServerPort string
	Database   string
	DeployPort string
	Key        string
}
