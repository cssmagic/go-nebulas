// Copyright (C) 2017 go-nebulas authors
//
// This file is part of the go-nebulas library.
//
// the go-nebulas library is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// the go-nebulas library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with the go-nebulas library.  If not, see <http://www.gnu.org/licenses/>.
//

package console

import (
	"fmt"

	"github.com/peterh/liner"
)

// Stdin holds the stdin line reader (also using stdout for printing prompts).
var Stdin = newTerminalPrompter()

type terminalPrompter struct {
	liner     *liner.State
	supported bool
	origMode  liner.ModeApplier
	rawMode   liner.ModeApplier
}

// newTerminalPrompter create a terminal prompter
func newTerminalPrompter() *terminalPrompter {
	p := new(terminalPrompter)
	// Get the original mode before calling NewLiner.
	origMode, _ := liner.TerminalMode()
	// Turn on liner.
	p.liner = liner.NewLiner()
	rawMode, err := liner.TerminalMode()
	if err != nil || !liner.TerminalSupported() {
		p.supported = false
	} else {
		p.supported = true
		p.origMode = origMode
		p.rawMode = rawMode
		// Switch back to normal mode while we're not prompting.
		origMode.ApplyMode()
	}
	p.liner.SetCtrlCAborts(true)
	p.liner.SetTabCompletionStyle(liner.TabPrints)
	p.liner.SetMultiLineMode(true)
	return p
}

// Prompt displays the prompt and requests some textual
// data to be entered, returning the input.
func (p *terminalPrompter) Prompt(prompt string) (string, error) {
	if p.supported {
		p.rawMode.ApplyMode()
		defer p.origMode.ApplyMode()
	} else {
		fmt.Print(prompt)
		defer fmt.Println()
	}
	return p.liner.Prompt(prompt)
}

// PromptPassphrase displays the prompt and requests some textual
// data to be entered, but one which must not be echoed out into the terminal.
// The method returns the input.
func (p *terminalPrompter) PromptPassphrase(prompt string) (passwd string, err error) {
	if p.supported {
		p.rawMode.ApplyMode()
		defer p.origMode.ApplyMode()
		return p.liner.PasswordPrompt(prompt)
	}

	fmt.Print(prompt)
	passwd, err = p.liner.Prompt("")
	fmt.Println()
	return passwd, err
}
