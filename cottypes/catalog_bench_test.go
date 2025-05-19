package cottypes

import "testing"

func BenchmarkCatalogGetType(b *testing.B) {
	cat := GetCatalog()
	for i := 0; i < b.N; i++ {
		if _, err := cat.GetType("a-f-G-E-X-N"); err != nil {
			b.Fatalf("GetType error: %v", err)
		}
	}
}

func BenchmarkCatalogFindByFullName(b *testing.B) {
	cat := GetCatalog()
	for i := 0; i < b.N; i++ {
		res := cat.FindByFullName("Nbc Equipment")
		if len(res) == 0 {
			b.Fatal("no results")
		}
	}
}
