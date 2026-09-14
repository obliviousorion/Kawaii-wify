package credentials

import (
	"github.com/zalando/go-keyring"
)

const ServiceName = "kawaii-wify"


func Set(username, password string) error {
	return keyring.Set(ServiceName, username, password)
}

func Get(username string) (string, error) {
	return keyring.Get(ServiceName, username)
}

func Delete(username string) error {
	return keyring.Delete(ServiceName, username)
}

