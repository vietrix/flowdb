package auth

import "flowdb/backend/util"

func RandomState() (string, error) {
	return util.RandomToken(32)
}
