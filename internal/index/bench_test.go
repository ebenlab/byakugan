package index

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// buildTree writes n HTML pages (~2.5 KB each) spread across projects.
func buildTree(tb testing.TB, root string, n int) {
	tb.Helper()
	body := strings.Repeat("<p>Ledger entries balance to zero across the journal partitions. </p>\n", 30)
	for i := 0; i < n; i++ {
		proj := fmt.Sprintf("project-%02d", i%40)
		dir := filepath.Join(root, proj)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			tb.Fatal(err)
		}
		page := fmt.Sprintf(`<html><head><title>Doc %d</title></head>
<body><h1>Doc %d</h1><h2>Design</h2>%s</body></html>`, i, i, body)
		if err := os.WriteFile(filepath.Join(dir, fmt.Sprintf("doc-%04d.html", i)), []byte(page), 0o644); err != nil {
			tb.Fatal(err)
		}
	}
}

func BenchmarkRebuild(b *testing.B) {
	for _, n := range []int{500, 2000, 5000} {
		b.Run(fmt.Sprintf("pages=%d", n), func(b *testing.B) {
			root := b.TempDir()
			buildTree(b, root, n)
			ix := New(root)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if err := ix.Rebuild(); err != nil {
					b.Fatal(err)
				}
			}
			b.StopTimer()
			payload, _ := json.Marshal(ix.Current())
			b.ReportMetric(float64(len(payload))/1024/1024, "MB-payload")
		})
	}
}
