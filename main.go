package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Goal: to be used by Dockerfile frontends in best-effort manner to install a root ca cert (/cacert.pem) into the OS store for
// various linux distros. This is just one step of a cert install process, such as pointing tech stacks/tools to them via env vars.
// Must work on "distroless" base images lacking a shell. Logs should be written to /tmp/install-certs.log if a debug flag is set,
// otherwise logs should go to stdout to avoid increasing image size. (?)
// Must always return 0 to avoid causing docker build to fail if certs could not be installed.
const RootCaFile = "/cacert.pem"

var OsInfoFiles = []string{"/etc/os-release", "/usr/lib/os-release"}

func main() {
	if fileExists(OsInfoFiles[0]) && !isStringEmpty(readFile(OsInfoFiles[0])) {
		distro := IdentifyDistro(OsInfoFiles[0])
		InstallCertsForDistro(distro)
	} else if fileExists(OsInfoFiles[1]) && !isStringEmpty(readFile(OsInfoFiles[1])) {
		distro := IdentifyDistro(OsInfoFiles[1])
		InstallCertsForDistro(distro)
	} else {
		fmt.Println("Could not identify distro, so did not install certs.")
	}
}

func InstallCertsForDistro(distro Distro) {
	switch distro {
	case Alpine:
		fmt.Println("Installing certs for Alpine")
		appendFile("/etc/ssl/certs/ca-certificates.crt", readFile(RootCaFile))
	case Debian:
		InstallCertsOnDebian(RootCaFile)
	case Fedora:
		InstallCertsOnFedora(RootCaFile)
	}
}

func InstallCertsOnFedora(rootCaFile string) {
	fmt.Println("Installing certs for Fedora")

	copyFile(rootCaFile, "/etc/pki/ca-trust/source/anchors")

	if fileExists("/usr/bin/update-ca-trust") {
		result := executeCommand("/usr/bin/update-ca-trust")
		fmt.Printf("Installed certs using update-ca-trust: %v\n", result)
	} else if fileExists("/usr/bin/trust") {
		result := executeCommand("/usr/bin/trust", "anchor", rootCaFile)
		fmt.Printf("Installed certs using trust: %v\n", result)
	} else if fileExists("/usr/bin/p11-kit") {
		result := executeCommand("/usr/bin/p11-kit", "extract", "--comment", "--format=pem-bundle", "--filter=certificates", "--overwrite", "--purpose", "server-auth", "/etc/pki/ca-trust/extracted/pem/tls-ca-bundle.pem")
		fmt.Printf("Installed certs using p11-kit: %v\n", result)
	} else {
		fmt.Println("Couldn't attempt Fedora install approach 3 because update-ca-trust, trust, and p11-kit were missing.")
		os.Exit(0)
	}
}

func InstallCertsOnDebian(rootCaFile string) {
	fmt.Println("Installing certs for Debian")
	copyFile(rootCaFile, "/usr/local/share/ca-certificates/"+rootCaFile)

	if !fileExists("/usr/sbin/update-ca-certificates") {
		if !fileExists("/usr/bin/apt") {
			fmt.Println("update-ca-certificates missing, and can't install it because apt is missing. Not installing certs.")
			os.Exit(0)
		}

		// Install ca-certificates using apt
		executeCommand("/usr/bin/apt", "update")
		executeCommand("/usr/bin/apt", "install", "-y", "ca-certificates")
	}

	// Execute update-ca-certificates
	executeCommand("/usr/sbin/update-ca-certificates")
}

func IdentifyDistro(releaseFile string) Distro {
	contents := strings.ToLower(readFile(releaseFile))
	switch {
	case strings.Contains(contents, "debian"):
		return Debian
	case strings.Contains(contents, "alpine"):
		return Alpine
	case strings.Contains(contents, "fedora"):
		return Fedora
	default:
		return Unknown
	}
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func isStringEmpty(s string) bool {
	return len(s) == 0
}

func readFile(filename string) string {
	content, err := os.ReadFile(filename)
	if err != nil {
		return ""
	}
	return string(content)
}

func appendFile(filename, content string) {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer f.Close()
	_, err = f.WriteString(content)
	if err != nil {
		fmt.Println("Error appending to file:", err)
	}
}

func copyFile(src, dst string) {
	data, err := os.ReadFile(src)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	err = os.WriteFile(dst, data, 0644)
	if err != nil {
		fmt.Println("Error copying file:", err)
	}
}

func executeCommand(command string, args ...string) int {
	cmd := exec.Command(command, args...)
	cmd.Dir, _ = os.Getwd()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error executing command:", err)
	}
	return cmd.ProcessState.ExitCode()
}

type Distro int

const (
	Unknown Distro = iota
	Alpine
	Debian
	Fedora
)
