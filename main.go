package main

import (
	"fmt"
	"os"

	"golang.design/x/clipboard"
)

const defaultProfileName = "default"

func main() {

	// this checks the command line args for a profile name override
	var profileName string
	profileNameFromArgs := len(os.Args) > 1
	if profileNameFromArgs {
		profileName = os.Args[1]
	}

	// this loads the incoming credential from clipboard data
	err := clipboard.Init()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	b := clipboard.Read(clipboard.FmtText)

	// this parses the clipboard data into a credential profile entry
	pe := parseNewCreds(b, profileName)
	if pe == nil {
		fmt.Println("❌ invalid credentials data in the clipboard")
		os.Exit(1)
	}

	// this use the default profileName if one is not set
	if profileName == "" {
		profileName = defaultProfileName
	}
	// this sets the proflieName in the incoming credential profile if needed
	if profileNameFromArgs || pe.insertSectionHeader {
		pe.Name = profileName
	}

	// this loads the credentials file from disk
	b = loadCreds()
	// and parses the file data to a list of credential profile entries
	pes := parseCredsFile(b)
	// this injects the incoming (clipboard-sourced) credential profile into
	// profile list sourced from the credentials file
	pes = injectCreds(pes, pe)
	// this writes the credentiails file with the incoming credential injected
	err = writeCreds(pes)
	if err != nil {
		fmt.Println("❌ profile not updated:", pe.Name)
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("✅ profile updated:", pe.Name)
}
