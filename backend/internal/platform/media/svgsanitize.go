package media

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

// disallowedElements are dropped along with their entire subtree. <script> and
// event handlers execute JS; <foreignObject> can embed arbitrary HTML/JS;
// <style> can pull remote resources / legacy expressions; the media/plugin
// elements can load active content.
var disallowedElements = map[string]bool{
	"script":           true,
	"foreignobject":    true,
	"style":            true,
	"iframe":           true,
	"embed":            true,
	"object":           true,
	"audio":            true,
	"video":            true,
	"animate":          true, // SMIL animate* can carry attributeName=href tricks
	"animatetransform": true,
	"set":              true,
	"handler":          true,
}

// SanitizeSVG parses an SVG and re-serializes it with a strict allowlist: it
// drops <script>/<foreignObject>/<style> (and subtree), every on* event handler
// attribute, the `style` attribute, any href/xlink:href that is not a same-file
// fragment (#id) — i.e. no remote/javascript/data refs — and any DTD/PI/comment
// (guards XXE + entity vectors). It errors if the input is not XML or has no
// <svg> root, so the caller returns 422.
func SanitizeSVG(in []byte) ([]byte, error) {
	dec := xml.NewDecoder(bytes.NewReader(in))
	// Do not resolve external entities (XXE): with no Entity map and default
	// settings, Go's decoder errors on external/unknown entity references rather
	// than fetching them.
	dec.Strict = false
	dec.AutoClose = xml.HTMLAutoClose

	var out bytes.Buffer
	enc := xml.NewEncoder(&out)

	skipDepth := 0
	sawSVG := false

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("svg: parse: %w", err)
		}

		switch t := tok.(type) {
		case xml.StartElement:
			if skipDepth > 0 {
				skipDepth++
				continue
			}
			name := strings.ToLower(t.Name.Local)
			if disallowedElements[name] {
				skipDepth = 1
				continue
			}
			if name == "svg" {
				sawSVG = true
			}
			clean := xml.StartElement{Name: t.Name, Attr: filterAttrs(t.Attr)}
			if err := enc.EncodeToken(clean); err != nil {
				return nil, err
			}
		case xml.EndElement:
			if skipDepth > 0 {
				skipDepth--
				continue
			}
			if err := enc.EncodeToken(t); err != nil {
				return nil, err
			}
		case xml.CharData:
			if skipDepth > 0 {
				continue
			}
			if err := enc.EncodeToken(t); err != nil {
				return nil, err
			}
		case xml.Comment, xml.Directive, xml.ProcInst:
			// Drop comments, DTDs/DOCTYPE, and processing instructions.
		}
	}
	if err := enc.Flush(); err != nil {
		return nil, err
	}
	if !sawSVG {
		return nil, fmt.Errorf("svg: no <svg> root element")
	}
	return out.Bytes(), nil
}

// filterAttrs drops dangerous attributes: any on* event handler, the style
// attribute, and any href/xlink:href that is not a same-document fragment.
func filterAttrs(attrs []xml.Attr) []xml.Attr {
	out := make([]xml.Attr, 0, len(attrs))
	for _, a := range attrs {
		local := strings.ToLower(a.Name.Local)
		if strings.HasPrefix(local, "on") {
			continue // event handler
		}
		if local == "style" {
			continue
		}
		if local == "href" { // covers SVG2 href and xlink:href (Local == "href")
			if !strings.HasPrefix(strings.TrimSpace(a.Value), "#") {
				continue // drop remote/javascript/data refs; keep fragment refs
			}
		}
		// Defensively drop any attribute whose value smuggles a javascript: URL.
		if strings.Contains(strings.ToLower(strings.ReplaceAll(a.Value, " ", "")), "javascript:") {
			continue
		}
		out = append(out, a)
	}
	return out
}
