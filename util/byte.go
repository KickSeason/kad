package util

import "errors"

//Xor to calculate xor
func Xor(a []byte, b []byte) ([]byte, error) {
	if len(a) != len(b) || len(a) == 0 {
		return nil, errors.New("[Xor] invalid input")
	}
	r := make([]byte, len(a))
	for i, v := range a {
		r[i] = v ^ b[i]
	}
	return r, nil
}
