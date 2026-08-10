package media

import (
	"strings"
	"testing"
)

func TestSanitizeSVGStripsActiveContent(t *testing.T) {
	in := `<?xml version="1.0"?>
<!DOCTYPE svg [<!ENTITY xxe "danger">]>
<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" onload="alert(1)" width="10" height="10">
  <script>alert('xss')</script>
  <style>@import url(https://evil.example/x.css);</style>
  <a xlink:href="https://evil.example" onclick="steal()"><rect width="10" height="10"/></a>
  <image href="https://evil.example/t.png"/>
  <use href="#ok"/>
  <foreignObject><body xmlns="http://www.w3.org/1999/xhtml"><script>bad()</script></body></foreignObject>
  <path d="M0 0h10v10H0z" fill="#f00"/>
</svg>`
	out, err := SanitizeSVG([]byte(in))
	if err != nil {
		t.Fatalf("SanitizeSVG: %v", err)
	}
	s := strings.ToLower(string(out))
	for _, bad := range []string{"<script", "onload", "onclick", "<foreignobject", "<style", "evil.example", "alert("} {
		if strings.Contains(s, bad) {
			t.Errorf("sanitized output still contains %q:\n%s", bad, out)
		}
	}
	// Benign content survives.
	if !strings.Contains(s, "<path") {
		t.Errorf("expected <path> to survive, got:\n%s", out)
	}
	// Fragment ref survives.
	if !strings.Contains(s, `href="#ok"`) {
		t.Errorf("expected fragment href to survive, got:\n%s", out)
	}
}

func TestSanitizeSVGRejectsNonSVG(t *testing.T) {
	if _, err := SanitizeSVG([]byte(`<html><body>hi</body></html>`)); err == nil {
		t.Fatal("expected error for non-SVG XML")
	}
}

func TestSanitizeSVGRejectsGarbage(t *testing.T) {
	if _, err := SanitizeSVG([]byte("\x00\x01 not xml at all <<<")); err == nil {
		t.Fatal("expected parse error for non-XML input")
	}
}
