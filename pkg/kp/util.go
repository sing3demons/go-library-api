package kp

import (
	"context"
	"crypto/rand"
	"errors"
	"math"
	"net/http"
	"strings"
	"time"
)

func removeBraces(str string) string {
	return strings.ReplaceAll(strings.ReplaceAll(str, "{", ""), "}", "")
}

type ContextKey string

func SetParam(path string, r *http.Request) *http.Request {
	subPath := strings.Split(path, "/")
	sss := strings.Split(r.URL.Path, "/")

	// Remove empty strings from the slice
	if len(subPath) == len(sss) {
		for i := 0; i < len(subPath); i++ {
			var k, v string
			if subPath[i] != "" {
				k = subPath[i]
			}
			if sss[i] != "" {
				v = sss[i]
			}

			if k != "" && v != "" && k != v {
				key := removeBraces(k)
				if key != "" {
					ctx := context.WithValue(r.Context(), ContextKey(key), v)
					r = r.WithContext(ctx)
				}
			}
		}
	}
	return r
}

func preHandle(final HandleFunc, middlewares ...Middleware) HandleFunc {
	if final == nil {
		panic("no final handler")
		// Or return a default handler.
	}
	// Execute the middleware in the same order and return the final func.
	// This is a confusing and tricky construct :)
	// We need to use the reverse order since we are chaining inwards.
	for i := len(middlewares) - 1; i >= 0; i-- {
		final = middlewares[i](final) // mw1(mw2(mw3(final)))
	}
	return final
}

func preMiddleware(app []Middleware, middlewares []Middleware) []Middleware {
	var m []Middleware
	if len(app) > 0 {
		m = append(m, app...)
	}

	if len(middlewares) > 0 {
		m = append(m, middlewares...)
	}
	return m
}

var defaultAlphabet = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

const (
	defaultSize = 22
)

func getMask(alphabetSize int) int {
	for i := 1; i <= 8; i++ {
		mask := (2 << uint(i)) - 1
		if mask >= alphabetSize-1 {
			return mask
		}
	}
	return 0
}

func Generate(alphabet string, size int) (string, error) {
	chars := []rune(alphabet)

	if len(alphabet) == 0 || len(alphabet) > 255 {
		return "", errors.New("alphabet must not be empty and contain no more than 255 chars")
	}
	if size <= 0 {
		size = defaultSize
	}

	mask := getMask(len(chars))
	// estimate how many random bytes we will need for the ID, we might actually need more but this is tradeoff
	// between average case and worst case
	ceilArg := 1.6 * float64(mask*size) / float64(len(alphabet))
	step := int(math.Ceil(ceilArg))

	id := make([]rune, size)
	bytes := make([]byte, step)
	for j := 0; ; {
		_, err := rand.Read(bytes)
		if err != nil {
			return "", err
		}
		for i := 0; i < step; i++ {
			currByte := bytes[i] & byte(mask)
			if currByte < byte(len(chars)) {
				id[j] = chars[currByte]
				j++
				if j == size {
					return string(id[:size]), nil
				}
			}
		}
	}
}

func GenerateXTid(nodeName string) string {
	now := time.Now()
	date := now.Format("20060102")
	var xTid string

	if nodeName == "" {
		xTid = "default" + "-" + date
	} else if len(nodeName) > 5 {
		xTid = nodeName[:5] + "-" + date
	} else {
		xTid = nodeName + "-" + date
	}
	remainingLength := 22 - len(xTid)
	id, err := Generate(string(defaultAlphabet), remainingLength)
	if err != nil {
		return xTid + now.Format("20060102150405")
	}
	return xTid + id
}
