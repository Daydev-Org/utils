/*
 * Copyright (c) 2022-2025. Daydev, Inc. All Rights Reserved
 */

package jwt

import "github.com/golang-jwt/jwt/v5"

// GenerateToken accepts any jwt.Claims.
// secret — signature key.
// always encrypts SigningMethodHS256.
func GenerateToken(secret []byte, claims jwt.Claims) (string, error) {
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(secret)
}
