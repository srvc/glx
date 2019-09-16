package prefixer

import (
	"bytes"
	"hash/fnv"
	"io"

	"github.com/fatih/color"
	"golang.org/x/text/transform"
)

func NewWriter(w io.Writer, prefix string, subprefix string) io.Writer {
	t := NewTransformer(buildPrefix(prefix, subprefix))
	return transform.NewWriter(w, t)
}

func buildPrefix(name, subname string) string {
	buf := new(bytes.Buffer)

	c := determineColor(name)
	color.New(c+(color.FgHiBlack-color.FgBlack)).Fprint(buf, name)

	if subname != "" {
		buf.WriteString(" ")
		color.New(c).Fprint(buf, subname)
	}

	return buf.String()
}

var (
	colorList = []color.Attribute{
		color.FgRed,
		color.FgGreen,
		color.FgYellow,
		color.FgBlue,
		color.FgMagenta,
		color.FgCyan,
	}
)

func determineColor(name string) color.Attribute {
	hash := fnv.New32()
	hash.Write([]byte(name))
	idx := hash.Sum32() % uint32(len(colorList))

	return colorList[idx]
}
