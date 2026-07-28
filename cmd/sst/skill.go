package main

import (
	"bytes"
	"embed"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// skillName is the directory agents look for inside a harness skills dir.
const skillName = "review-writing-slop"

//go:embed all:skill
var skillFS embed.FS

var skillDoc = mustReadSkill("SKILL.md")

func mustReadSkill(name string) []byte {
	doc, err := skillFS.ReadFile("skill/" + name)
	if err != nil {
		panic(err)
	}
	return doc
}

const skillUsage = `sst skill - print or install the agent review skill

Usage:
  sst skill                  print SKILL.md to stdout
  sst skill --install DIR    write DIR/` + skillName + `/SKILL.md
  sst skill --install DIR --force   overwrite an existing different file

Install into your harness's skills directory, e.g. .claude/skills,
.factory/skills, or .agents/skills.
`

func runSkill(argv []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("skill", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	installDir := fs.String("install", "", "install into this skills directory")
	force := fs.Bool("force", false, "overwrite an existing different file")
	help := fs.Bool("help", false, "show help")
	fs.BoolVar(help, "h", false, "show help")
	if err := fs.Parse(argv); err != nil {
		fmt.Fprintln(stderr, "sst skill:", err)
		fmt.Fprintln(stderr, "try 'sst skill --help'")
		return exitUsage
	}
	if *help {
		fmt.Fprint(stdout, skillUsage)
		return exitOK
	}
	if fs.NArg() != 0 {
		fmt.Fprintln(stderr, "sst skill: unexpected positional arguments")
		fmt.Fprintln(stderr, "try 'sst skill --help'")
		return exitUsage
	}
	if *installDir == "" {
		if _, err := stdout.Write(skillDoc); err != nil {
			fmt.Fprintln(stderr, "sst skill:", err)
			return exitError
		}
		return exitOK
	}
	wrote, err := installSkill(*installDir, *force)
	if err != nil {
		fmt.Fprintln(stderr, "sst skill:", err)
		return exitError
	}
	target := filepath.Join(*installDir, skillName, "SKILL.md")
	if wrote {
		fmt.Fprintf(stdout, "installed %s\n", target)
	} else {
		fmt.Fprintf(stdout, "already up to date: %s\n", target)
	}
	return exitOK
}

// installSkill writes every embedded skill file under DIR/<skillName>/,
// preserving the embedded directory layout.
func installSkill(dir string, force bool) (wrote bool, err error) {
	root := filepath.Join(dir, skillName)
	err = fs.WalkDir(skillFS, "skill", func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel("skill", path)
		if err != nil {
			return err
		}
		target := filepath.Join(root, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o750)
		}
		content, err := skillFS.ReadFile(path)
		if err != nil {
			return err
		}
		if existing, readErr := os.ReadFile(target); readErr == nil {
			if bytes.Equal(existing, content) {
				return nil
			}
			if !force {
				return fmt.Errorf("%s exists with different content (use --force to overwrite)", target)
			}
		}
		if err := os.WriteFile(target, content, 0o600); err != nil {
			return fmt.Errorf("write %s: %w", target, err)
		}
		wrote = true
		return nil
	})
	return wrote, err
}
