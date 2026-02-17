package rules

import (
	"strings"
	"testing"
)

func TestBuildPreview_TitleWithFinancing(t *testing.T) {
	svc := NewService()
	preview := svc.BuildPreview("123", "Compromisso de Venda e Compra de Imovel", map[string]any{
		"preco_financiamento": "R$ 100.000,00",
	})

	if preview.Title != "Compromisso de Venda e Compra de Imovel com Financiamento" {
		t.Fatalf("unexpected title: %s", preview.Title)
	}
}

func TestBuildPreview_CessaoRoles(t *testing.T) {
	svc := NewService()
	preview := svc.BuildPreview("10", "Cessao de Posse e Direitos sobre Imovel", map[string]any{})

	if len(preview.Sections) < 3 {
		t.Fatalf("expected sections >= 3")
	}
	if len(preview.Sections[1]) < len("PARTE CEDENTE") || preview.Sections[1][:len("PARTE CEDENTE")] != "PARTE CEDENTE" {
		t.Fatalf("unexpected seller role section: %s", preview.Sections[1])
	}
	if len(preview.Sections[2]) < len("PARTE CESSIONARIA") || preview.Sections[2][:len("PARTE CESSIONARIA")] != "PARTE CESSIONARIA" {
		t.Fatalf("unexpected buyer role section: %s", preview.Sections[2])
	}
}

func TestBuildPreview_DeliveryManualText(t *testing.T) {
	svc := NewService()
	preview := svc.BuildPreview("101", "Compromisso de Compra e Venda de Imovel", map[string]any{
		"entrega_chaves":       "Escrever no contrato",
		"entrega_chaves_texto": "Entrega em ate 5 dias uteis apos quitacao.",
	})

	found := false
	for _, section := range preview.Sections {
		if section == "Entrega de chaves: Entrega em ate 5 dias uteis apos quitacao." {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected manual delivery text in sections: %#v", preview.Sections)
	}
}

func TestBuildPreview_FullTextWithIndexedCustomClause(t *testing.T) {
	svc := NewService()
	preview := svc.BuildPreview("999", "Compromisso de Venda e Compra de Imovel", map[string]any{
		"clausulas_customizadas": []any{
			map[string]any{
				"titulo":   "CLAUSULA X",
				"conteudo": "Texto customizado de teste.",
				"indice":   "1.1.2",
			},
		},
		"clausulas_selecionadas_vinculos": []any{
			map[string]any{
				"clause_key": "multa_atraso",
				"title":      "Multa por atraso",
				"content":    "Aplica multa padrao.",
				"indice":     "2.1.1",
			},
		},
	})

	if !strings.Contains(preview.FullText, "1.1.2 CLAUSULA X") {
		t.Fatalf("expected indexed custom clause in full text")
	}
	if !strings.Contains(preview.FullText, "2.1.1 Aplica multa padrao.") {
		t.Fatalf("expected indexed selected clause content in full text")
	}
}
