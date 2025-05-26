package cottypes

import (
	"context"
	"testing"
)

func BenchmarkCatalogGetType(b *testing.B) {
	cat := GetCatalog()
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		if _, err := cat.GetType(ctx, "a-f-G-E-X-N"); err != nil {
			b.Fatalf("GetType error: %v", err)
		}
	}
}

func BenchmarkCatalogFindByFullName(b *testing.B) {
	cat := GetCatalog()
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		res := cat.FindByFullName(ctx, "Nbc Equipment")
		if len(res) == 0 {
			b.Fatal("no results")
		}
	}
}

func BenchmarkCatalogFindByDescription(b *testing.B) {
	cat := GetCatalog()
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		res := cat.FindByDescription(ctx, "NBC EQUIPMENT")
		if len(res) == 0 {
			b.Fatal("no results")
		}
	}
}
