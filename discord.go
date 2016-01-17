// Discordgo - Discord bindings for Go
// Available at https://github.com/bwmarrin/discordgo

// Copyright 2015-2016 Bruce Marriner <bruce@sqls.net>.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file contains high level helper functions and easy entry points for the
// entire discordgo package.  These functions are beling developed and are very
// experimental at this point.  They will most likley change so please use the
// low level functions if that's a problem.

// package discordgo provides Discord binding for Go
package discordgo

import "fmt"

// Discordgo Version, follows Symantic Versioning. (http://semver.org/)
const VERSION = "0.10.0-alpha"

// New creates a new Discord session and will automate some startup
// tasks if given enough information to do so.  Currently you can pass zero
// arguments and it will return an empty Discord session.
// There are 3 ways to call New:
//     With a single auth token - All requests will use the token blindly,
//         no verification of the token will be done and requests may fail.
//     With an email and password - Discord will sign in with the provided
//         credentials.
//     With an email, password and auth token - Discord will verify the auth
//         token, if it is invalid it will sign in with the provided
//         credentials. This is the Discord recommended way to sign in.
func New(args ...interface{}) (s *Session, err error) {

	// Create an empty Session interface.
	s = &Session{
		State:        NewState(),
		StateEnabled: true,
	}

	// If no arguments are passed return the empty Session interface.
	// Later I will add default values, if appropriate.
	if args == nil {
		return
	}

	// Variables used below when parsing func arguments
	var auth, pass string

	// Parse passed arguments
	for _, arg := range args {

		switch v := arg.(type) {

		case []string:
			if len(v) > 3 {
				err = fmt.Errorf("Too many string parameters provided.")
				return
			}

			// First string is either token or username
			if len(v) > 0 {
				auth = v[0]
			}

			// If second string exists, it must be a password.
			if len(v) > 1 {
				pass = v[1]
			}

			// If third string exists, it must be an auth token.
			if len(v) > 2 {
				s.Token = v[2]
			}

		case string:
			// First string must be either auth token or username.
			// Second string must be a password.
			// Only 2 input strings are supported.

			if auth == "" {
				auth = v
			} else if pass == "" {
				pass = v
			} else if s.Token == "" {
				s.Token = v
			} else {
				err = fmt.Errorf("Too many string parameters provided.")
				return
			}

			//		case Config:
			// TODO: Parse configuration

		default:
			err = fmt.Errorf("Unsupported parameter type provided.")
			return
		}
	}

	// If only one string was provided, assume it is an auth token.
	// Otherwise get auth token from Discord, if a token was specified
	// Discord will verify it for free, or log the user in if it is
	// invalid.
	if pass == "" {
		s.Token = auth
	} else {
		err = s.Login(auth, pass)
		if err != nil || s.Token == "" {
			err = fmt.Errorf("Unable to fetch discord authentication token. %v", err)
			return
		}
	}

	// TODO: Add code here to fetch authenticated user info like settings,
	// avatar, User ID, etc.  If fails, return error.

	// Open websocket connection
	err = s.Open()
	if err != nil {
		fmt.Println(err)
		return
	}

	// Do websocket handshake.
	err = s.Handshake()
	if err != nil {
		fmt.Println(err)
		return
	}

	// Listen for events.
	go s.Listen()

	return
}

// Close closes a Discord session
// TODO: Add support for Voice WS/UDP connections
func (s *Session) Close() {

	s.DataReady = false

	if s.heartbeatChan != nil {
		select {
		case <-s.heartbeatChan:
			break
		default:
			close(s.heartbeatChan)
		}
		s.heartbeatChan = nil
	}

	if s.listenChan != nil {
		select {
		case <-s.listenChan:
			break
		default:
			close(s.listenChan)
		}
		s.listenChan = nil
	}

	if s.wsConn != nil {
		s.wsConn.Close()
	}
}
