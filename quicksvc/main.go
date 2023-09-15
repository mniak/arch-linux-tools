package main

import (
	"embed"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

func main() {
	var flagName string
	cmd := cobra.Command{
		Use: "quicksvc [<flags>] -- <program> [<args>]",
		RunE: func(cmd *cobra.Command, args []string) error {
			var programArgs []string
			if cmd.ArgsLenAtDash() < 0 {
				return errors.New("dash (--) must be present on the argument list")
			}
			cmdArgs := args[:cmd.ArgsLenAtDash()]
			programArgs = args[cmd.ArgsLenAtDash():]

			if len(programArgs) == 0 {
				return errors.New("program name is required")
			}
			program := lo.Must(filepath.Abs(programArgs[0]))
			programArgs = programArgs[1:]

			log.Println("CMD args:", cmdArgs)
			log.Println("Program:", program)
			log.Println("Program args:", programArgs)

			createPkgbuild(flagName, program, programArgs...)
			return nil
		},
	}
	cmd.Flags().StringVar(&flagName, "name", "", "Specify the service name")
	cmd.Execute()
}

//go:embed templates/*.tmpl
var templatesFs embed.FS

func createPkgbuild(name, sourceProgramPath string, args ...string) {
	sourceProgramPath = try(filepath.Abs(sourceProgramPath)).
		withMessage("failed to get absolute path to program").
		must()

	packageName := filepath.Base(sourceProgramPath)
	if name == "" {
		packageName = fmt.Sprintf("quicksvc-%s", name)
	}
	programPath := filepath.Join("/usr", "bin", packageName)

	tempdir := try(os.MkdirTemp("", "quicksvc-*")).
		withMessage("failed to create temporary directory").
		must()
	defer os.RemoveAll(tempdir)

	svcPath := filepath.Join(tempdir, packageName+".service")
	svcFile := try(os.OpenFile(svcPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)).
		withMessage("failed to create service file").
		must()
	defer svcFile.Close()

	pkgbuildPath := filepath.Join(tempdir, "PKGBUILD")
	pkgbuildFile := try(os.OpenFile(pkgbuildPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)).
		withMessage("failed to create PKGBUILD file").
		must()

	defer pkgbuildFile.Close()

	tmpl := must(template.ParseFS(templatesFs, "templates/*.tmpl"))
	templateData := map[string]any{
		"ServiceDescription": fmt.Sprintf("%s service", packageName),
		"PackageName":        packageName,
		"SourceProgramPath":  sourceProgramPath,
		"ProgramPath":        programPath,
		"ProgramArgs":        strings.Join(args, " "),
	}
	must0(tmpl.ExecuteTemplate(svcFile, "service.tmpl", templateData))
	must0(tmpl.ExecuteTemplate(pkgbuildFile, "PKGBUILD.bin.tmpl", templateData))

	cmd := exec.Command("makepkg", "-i")
	cmd.Dir = tempdir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	must0(cmd.Run())
}
