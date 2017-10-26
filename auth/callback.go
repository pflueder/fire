package auth

import (
	"errors"

	"github.com/gonfire/fire"
	"github.com/gonfire/oauth2"
)

// Callback returns a callback that can be used to protect resources by
// requiring an access token with the provided scope to be granted.
//
// Note: It requires that the request has already been authorized using the
// Authorizer middleware from a Manager.
func Callback(scope string) fire.Callback {
	return func(ctx *fire.Context) error {
		// parse scope
		s := oauth2.ParseScope(scope)

		// get access token
		accessToken, ok := ctx.HTTPRequest.Context().Value(AccessTokenContextKey).(Token)
		if !ok || accessToken == nil {
			return fire.Fatal(errors.New("missing access token"))
		}

		// validate scope
		if !oauth2.Scope(accessToken.GetTokenData().Scope).Includes(s) {
			return errors.New("unauthorized")
		}

		return nil
	}
}
