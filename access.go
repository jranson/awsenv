package main

import (
	"fmt"
	"os"
	"path"
	"strings"
)

const credsDir = ".aws"
const credsFile = "credentials"

var credPath = os.Getenv("HOME") + "/" + credsDir + "/" + credsFile

// see https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html

// ProfileLines aliases []string so functions can be attached
type ProfileLines []string

// ProfileEntry defines a TOML entry in ~/.aws/credentials
type ProfileEntry struct {
	Name                string
	Lines               ProfileLines
	insertSectionHeader bool
}

// ProfileEntries aliases []*ProfileEntry so functions can be attached
type ProfileEntries []*ProfileEntry

// parseNewCreds returns an ProfileEntry ptr from the input, as copied from
// Option 1 (env) or Option 2 (credentials file) of an IAM Identity Center
// login page. nil is returned if the input can't be parsed.
func parseNewCreds(input []byte, nameOverride string) *ProfileEntry {

	var pe *ProfileEntry

	// pe will be nil unless the input was parsed
	setupAccessObj := func() {
		if pe == nil {
			pe = &ProfileEntry{}
		}
	}

	var foundKeyId bool

	for _, l := range strings.Split(string(input), "\n") {
		// this part handles export lines (Option 1) for MacOS / Linux
		if strings.HasPrefix(l, "export AWS_ACCESS_KEY_ID=") {
			setupAccessObj()
			pe.Lines = append(pe.Lines,
				"aws_access_key_id="+strings.Trim(l[25:], `"`))
			foundKeyId = true
		} else if strings.HasPrefix(l, "export AWS_SECRET_ACCESS_KEY=") {
			setupAccessObj()
			pe.Lines = append(pe.Lines,
				"aws_secret_access_key="+strings.Trim(l[29:], `"`))
		} else if strings.HasPrefix(l, "export AWS_SESSION_TOKEN=") {
			setupAccessObj()
			pe.Lines = append(pe.Lines,
				"aws_session_token="+strings.Trim(l[25:], `"`))

			// this part handles export lines (Option 1) for Windows
		} else if strings.HasPrefix(l, "SET AWS_ACCESS_KEY_ID=") {
			setupAccessObj()
			pe.Lines = append(pe.Lines,
				"aws_access_key_id="+strings.Trim(l[22:], `"`))
			foundKeyId = true
		} else if strings.HasPrefix(l, "SET AWS_SECRET_ACCESS_KEY=") {
			setupAccessObj()
			pe.Lines = append(pe.Lines,
				"aws_secret_access_key="+strings.Trim(l[26:], `"`))
		} else if strings.HasPrefix(l, "SET AWS_SESSION_TOKEN=") {
			setupAccessObj()
			pe.Lines = append(pe.Lines,
				"aws_session_token="+strings.Trim(l[22:], `"`))
			// this part handles export lines (Option 1) for PowerShell
		} else if strings.HasPrefix(l, "$Env:AWS_ACCESS_KEY_ID=") {
			setupAccessObj()
			pe.Lines = append(pe.Lines,
				"aws_access_key_id="+strings.Trim(l[23:], `"`))
			foundKeyId = true
		} else if strings.HasPrefix(l, "$Env:AWS_SECRET_ACCESS_KEY=") {
			setupAccessObj()
			pe.Lines = append(pe.Lines,
				"aws_secret_access_key="+strings.Trim(l[27:], `"`))
		} else if strings.HasPrefix(l, "$Env:AWS_SESSION_TOKEN=") {
			setupAccessObj()
			pe.Lines = append(pe.Lines,
				"aws_session_token="+strings.Trim(l[23:], `"`))
			// this part handles a newly-named TOML section from Option 2
			// (e.g., [123456789012_MySelectedRole])
		} else if strings.HasPrefix(l, "[") && strings.HasSuffix(l, "]") {
			setupAccessObj()
			pe.Name = l[1 : len(l)-1]
			if nameOverride != "" {
				l = strings.ReplaceAll(l, pe.Name, nameOverride)
				pe.Name = nameOverride
			}
			pe.Lines = append(pe.Lines, l)
			// this part adds any TOML-like configs to the Lines list as-is
			// from Option 2
		} else {
			setupAccessObj()
			foundKeyId = foundKeyId || strings.Contains(l, "aws_access_key_id")
			pe.Lines = append(pe.Lines, l)
		}
	}
	// return nil (invalid input) if no lines containing aws_access_key_id or
	// AWS_ACCESS_KEY_ID were present in the input
	if !foundKeyId {
		return nil
	}
	if pe != nil {
		// if Option 1 was copied (exports), a manual section header
		// will need to be written
		pe.insertSectionHeader = pe.Name == ""
	}
	return pe
}

// parseCredsFile is a basic toml parser for AWS Credentials file data
func parseCredsFile(input []byte) ProfileEntries {
	pes := make(ProfileEntries, 0)
	var pe *ProfileEntry
	// iterate through the input data line-by-line
	for _, l := range strings.Split(string(input), "\n") {
		if l == "" {
			continue
		}
		// this part handles a newly-named TOML section (e.g., [default])
		if strings.HasPrefix(l, "[") && strings.Contains(l, "]") {
			if pe != nil {
				// add the previously found credential
				pes = append(pes, pe)
			}

			pn := strings.Replace(strings.Replace(l, "[", "", 1), "]", "", 1)
			pe = &ProfileEntry{
				Name:  pn,
				Lines: make(ProfileLines, 0, 4),
			}
		}
		// lines coming prior to the first TOML section are currently ignored
		if pe == nil {
			continue
		}
		// add the line to the ProfileEntry verbatim
		pe.Lines = append(pe.Lines, l)
	}
	if pe != nil {
		// add the final credential
		pes = append(pes, pe)
	}
	return pes
}

// loadCreds reads the credentials file from the default path into a byte slice
func loadCreds() []byte {
	b, _ := os.ReadFile(credPath)
	return b
}

// injectCreds will inject the parsed clipboard credentials into the parsed
// credentials file, by either replacing an existing entry or adding a new one
func injectCreds(pes ProfileEntries, pe *ProfileEntry) ProfileEntries {
	if pe == nil || len(pe.Lines) == 0 {
		return pes
	}
	var found bool
	for i, pe2 := range pes {
		// this indicates the desired profile name exists and is overwritten
		if pe.Name == pe2.Name {
			found = true
			pes[i] = pe
			break
		}
	}
	if !found {
		// this indicates that the desired profile name does not exist and
		// will be appended to the bottom of the config entries list
		pes = append(pes, pe)
	}
	return pes
}

// writeCreds will write the ProfileEnries list out to the default creds file
func writeCreds(pes ProfileEntries) error {
	var sb strings.Builder
	for _, pe := range pes {
		if pe == nil || pe.Name == "" || len(pe.Lines) == 0 {
			continue
		}
		if pe.insertSectionHeader && pe.Name != "" {
			sb.WriteString(fmt.Sprintf("[%s]\n", pe.Name))
		}
		for _, l := range pe.Lines {
			if l == "" {
				continue
			}
			sb.WriteString(l + "\n")
		}
		sb.WriteString("\n")
	}
	err := os.MkdirAll(path.Dir(credPath), 0700)
	if err != nil {
		return err
	}
	return os.WriteFile(credPath, []byte(sb.String()), 0600)
}
