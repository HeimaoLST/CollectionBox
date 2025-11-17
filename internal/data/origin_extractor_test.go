package data

import (
	"context"
	"testing"
)

// helper to construct extractor with in-memory origin map
func newTestExtractor() *jsonOriginExtractor {
	return &jsonOriginExtractor{originMap: map[string]string{
		"bilibili.com": "Bilibili",
	}}
}

func TestExtractAll_ConcatenatedBilibiliURLs(t *testing.T) {
	extractor := newTestExtractor()
	text := "https://www.bilibili.com/video/BV15j1EBWEA7/?spm_id_from=333.1007.tianma.1-2-2.click&vd_source=x" +
		"https://www.bilibili.com/video/BV19qxNzXEWT/?spm_id_from=333.1007.tianma.1-1-1.click&vd_source=x"

	pairs, err := extractor.ExtractAll(context.Background(), text)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pairs) != 2 {
		t.Fatalf("expected 2 pairs, got %d: %+v", len(pairs), pairs)
	}
	for _, p := range pairs {
		if p.Origin != "Bilibili" {
			t.Fatalf("expected origin Bilibili, got %s", p.Origin)
		}
	}
}

func TestExtractAll_BareDomainURL(t *testing.T) {
	extractor := newTestExtractor()
	text := "收藏这个链接：www.bilibili.com/video/BV1xx411c7mD"

	pairs, err := extractor.ExtractAll(context.Background(), text)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pairs) != 1 {
		t.Fatalf("expected 1 pair, got %d: %+v", len(pairs), pairs)
	}
	if pairs[0].Origin != "Bilibili" {
		t.Fatalf("expected origin Bilibili, got %s", pairs[0].Origin)
	}
}
