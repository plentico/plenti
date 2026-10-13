package build

import (
	"strings"
	"testing"
)

func TestWrappedStaticImportsAreMatchedIndividually(t *testing.T) {
	ssr := `import { generated } from "svelte/internal";
import {
	aVeryLongExportedConstantNameUsedOnlyToForceImportWrapping
} from "./one.svelte";
import {
	anotherLongExportedConstantNameUsedOnlyToForceImportWrapping
} from "./two.svelte";`

	imports := reStaticImport.FindAllString(ssr, -1)
	if len(imports) != 3 {
		t.Fatalf("got %d import matches, want 3: %#v", len(imports), imports)
	}

	wantNames := []string{
		"generated",
		"aVeryLongExportedConstantNameUsedOnlyToForceImportWrapping",
		"anotherLongExportedConstantNameUsedOnlyToForceImportWrapping",
	}
	for i, imported := range imports {
		nameMatch := reStaticImportName.FindStringSubmatch(imported)
		if len(nameMatch) < 2 {
			t.Fatalf("import %d has no captured binding list: %q", i, imported)
		}
		if got := strings.Join(strings.Fields(nameMatch[1]), " "); got != "{ "+wantNames[i]+" }" {
			t.Errorf("import %d binding list = %q, want %q", i, got, "{ "+wantNames[i]+" }")
		}
	}
}
