// Copyright 2014 The imapsrv Authors.
// Copyright 2016 Duzy Chan <code@duzy.info>
//
// All rights reserved.
// 
// Use of this source code is governed by a BSD-style
// license that can be found in the imapsrv.LICENSE file.

package imap

import (
	"fmt"
)

// An IMAP command
type command interface {
	// Execute the command and return an imap response
	execute(s *session) *serverResponse
}

type serveNoopCommand struct {
	tag string
}

// Execute a NOOP
func (c *serveNoopCommand) execute(s *session) *serverResponse {
	return ok(c.tag, "NOOP Completed")
}

// A CAPABILITY command
type serveCapabilityCommand struct {
	tag string
}

// Execute a capability
func (c *serveCapabilityCommand) execute(s *session) *serverResponse {
	// The IMAP server is assumed to be running over SSL and so
	// STARTTLS is not supported and LOGIN is not disabled
	return ok(c.tag, "CAPABILITY completed").
		extra("CAPABILITY IMAP4rev1")
}

// A LOGIN command
type serveLoginCommand struct {
	tag      string
	userId   string
	password string
}

// Login command
func (c *serveLoginCommand) execute(sess *session) *serverResponse {

	// Has the user already logged in?
	if sess.st != notAuthenticated {
		message := "LOGIN already logged in"
		sess.log(message)
		return bad(c.tag, message)
	}

	// TODO: implement login
	if /*c.userId == "test@example.com"*/true {
		sess.st = authenticated
		return ok(c.tag, "LOGIN completed")
	}

	// Fail by default
	return no(c.tag, "LOGIN failure")
}

// A LOGOUT command
type serveLogoutCommand struct {
	tag string
}

// Logout command
func (c *serveLogoutCommand) execute(sess *session) *serverResponse {

	sess.st = notAuthenticated
	return ok(c.tag, "LOGOUT completed").
		extra("BYE IMAP4rev1 Server logging out").
		shouldClose()
}

// A SELECT command
type serveSelectCommand struct {
	tag     string
	mailbox string
}

// Select command
func (c *serveSelectCommand) execute(sess *session) *serverResponse {

	// Is the user authenticated?
	if sess.st != authenticated {
		message := "SELECT not authenticated"
		sess.log(message)
		return bad(c.tag, message)
	}

	// Select the mailbox
	exists, err := sess.selectMailbox(c.mailbox)

	if err != nil {
		return internalError(sess, c.tag, "SELECT", err)
	}

	if !exists {
		return no(c.tag, "SELECT No such mailbox")
	}

	// Build a response that includes mailbox information
	res := ok(c.tag, "SELECT completed")

	err = sess.addMailboxInfo(res)

	if err != nil {
		return internalError(sess, c.tag, "SELECT", err)
	}

	return res
}

type serveListCommand struct {
        tag string
}

func (c *serveListCommand) execute(s *session) *serverResponse {
	message := fmt.Sprintf("LIST %s", c.tag)
	s.log(message)
	return bad(c.tag, message)
}

type serveFetchCommand struct {
        tag string
        seq string
        dat string
}

func (c *serveFetchCommand) execute(s *session) *serverResponse {
	message := fmt.Sprintf("%s FETCH %s %s", c.tag, c.seq, c.dat)
	s.log(message)
	return bad(c.tag, message)
}

// An unknown/unsupported command
type serveUnknownCommand struct {
	tag string
	cmd string
}

// Report an error for an unknown command
func (c *serveUnknownCommand) execute(s *session) *serverResponse {
	message := fmt.Sprintf("BAD unknown '%s' command", c.cmd)
	s.log(message)
	return bad(c.tag, message)
}

//------ Helper functions ------------------------------------------------------

// Log an error and return an response
func internalError(sess *session, tag string, commandName string, err error) *serverResponse {
	message := commandName + " " + err.Error()
	sess.log(message)
	return no(tag, message).shouldClose()
}
