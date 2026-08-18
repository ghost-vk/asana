package commands

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/ghost-vk/asana/richtext"
)

// bodyFormat is how the user asked a body to be interpreted before it is sent.
type bodyFormat int

const (
	formatPlain    bodyFormat = iota // notes, verbatim
	formatHTML                       // html_notes, already in Asana's tag subset
	formatMarkdown                   // markdown, converted to that subset here
)

// resolveFormat reads the mutually exclusive --html / --md flags.
func resolveFormat(c *cli.Context) bodyFormat {
	html, md := c.Bool("html"), c.Bool("md")
	if html && md {
		log.Fatal("fatal: --html and --md are mutually exclusive")
	}
	switch {
	case html:
		return formatHTML
	case md:
		return formatMarkdown
	default:
		return formatPlain
	}
}

// RichTextFlags are shared by every command that writes a body.
func RichTextFlags(fileUsage string) []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{Name: "html", Usage: "treat the body as Asana rich-text HTML (html_notes)"},
		&cli.BoolFlag{Name: "md", Usage: "convert the body from markdown to Asana rich text"},
		&cli.StringFlag{Name: "file", Aliases: []string{"f"}, Usage: fileUsage},
	}
}

// readBody takes the body from --file, else from the inline argument. Documents
// run to tens of kilobytes, which is more than a comfortable command-line
// argument. `--file -` reads stdin; stdin is never consulted implicitly,
// because a command run from a script or an agent has a non-tty stdin that
// never reaches EOF, and the CLI would simply hang.
//
// inlineGiven says whether the caller actually saw an inline argument, which an
// empty string cannot express: `asana body <gid> ""` clears the body and must
// stay distinguishable from passing nothing at all.
func readBody(c *cli.Context, inline string, inlineGiven bool) (string, bool) {
	path := c.String("file")
	if path == "" {
		if inlineGiven {
			return inline, true
		}
		return "", false
	}
	if inlineGiven {
		log.Fatal("fatal: pass the body either as an argument or via --file, not both")
	}
	if path == "-" {
		content, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			log.Fatalf("fatal: cannot read stdin: %v", err)
		}
		return string(content), true
	}
	content, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("fatal: cannot read %s: %v", path, err)
	}
	return string(content), true
}

// prepareBody converts and validates a body, returning the payload plus whether
// it belongs in the rich-text field. Validation happens here because Asana
// answers every malformed document with the same opaque xml_parsing_error.
func prepareBody(text string, format bodyFormat) (string, bool) {
	switch format {
	case formatMarkdown:
		return richtext.FromMarkdown(text), true
	case formatHTML:
		if err := richtext.Validate(text); err != nil {
			log.Fatalf("fatal: %v", err)
		}
		return richtext.Compact(text), true
	default:
		return text, false
	}
}

// describeFormat names the Asana field a body was written to, for the
// confirmation line.
func describeFormat(html bool) string {
	if html {
		return "html_notes"
	}
	return "notes"
}
