package hooks

import "embed"

//go:embed templates
var templateFS embed.FS

// HookNames lists all git hooks that autovendor manages.
var HookNames = []string{"post-merge", "post-checkout", "post-rewrite"}

// Template returns the shell script content for the given hook name.
func Template(name string) ([]byte, error) {
	return templateFS.ReadFile("templates/" + name + ".sh")
}
