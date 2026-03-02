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

	// Clausulas indexadas priorizam conteudo juridico quando informado.
	if !strings.Contains(preview.FullText, "1.1.2 Texto customizado de teste.") {
		t.Fatalf("expected indexed custom clause content in full text")
	}
	if !strings.Contains(preview.FullText, "2.1.1 Aplica multa padrao.") {
		t.Fatalf("expected indexed selected clause content in full text")
	}
}

func TestBuildPartyQualification_IncludesConjugeForUniaoEstavel(t *testing.T) {
	data := map[string]any{
		"vend_1__tipo":              "Pessoa Fisica",
		"vend_1__nome":              "Reginaldo Pospi do Nascimento Junior",
		"vend_1__nacionalidade":     "brasileiro(a)",
		"vend_1__estado_civil":      "uniao estavel",
		"vend_1__regime_bens":       "comunhao parcial de bens",
		"vend_1__profissao":         "advogado",
		"vend_1__rg":                "44.497.309-6",
		"vend_1__conj_nome":         "Maria Silva",
		"vend_1__conj_nacionalidade": "brasileira",
		"vend_1__conj_profissao":    "arquiteta",
		"vend_1__conj_rg":           "11.222.333-4",
		"vend_1__conj_cpf":          "111.222.333-44",
		"vend_1__end__texto":        "Rua Miguel dos Santos, n.o 373, Jardim Casa Branca, Suzano/SP - CEP: 08663-040",
	}

	qualification := buildPartyQualification(data, "vend_1")
	if !strings.Contains(qualification, "e MARIA SILVA, brasileira, arquiteta, RG n. 11.222.333-4, CPF n. 111.222.333-44") {
		t.Fatalf("expected spouse data block in qualification: %s", qualification)
	}
	if !strings.Contains(qualification, "ambos conviventes em uniao estavel entre si sob o regime de comunhao parcial de bens") {
		t.Fatalf("expected spouse details for uniao estavel: %s", qualification)
	}
	if !strings.Contains(qualification, "e residentes na Rua Miguel dos Santos") {
		t.Fatalf("expected both residents wording: %s", qualification)
	}
}

func TestBuildPartyQualification_IncludesConjugeForCasado(t *testing.T) {
	data := map[string]any{
		"compr_1__tipo":          "Pessoa Fisica",
		"compr_1__nome":          "Joao Pereira",
		"compr_1__nacionalidade": "brasileiro(a)",
		"compr_1__estado_civil":  "casado(a)",
		"compr_1__regime_bens":   "separacao total de bens",
		"compr_1__profissao":     "engenheiro",
		"compr_1__conj_nome":     "Ana Costa",
		"compr_1__end__texto":    "Avenida Brasil, n.o 10, Sao Paulo/SP - CEP: 01000-000",
	}

	qualification := buildPartyQualification(data, "compr_1")
	if !strings.Contains(qualification, "e ANA COSTA") {
		t.Fatalf("expected spouse details for casado(a): %s", qualification)
	}
	if !strings.Contains(qualification, "ambos casados entre si sob o regime de separacao total de bens") {
		t.Fatalf("expected spouse details for casado(a): %s", qualification)
	}
	if !strings.Contains(qualification, "e residentes na Avenida Brasil") {
		t.Fatalf("expected both residents wording: %s", qualification)
	}
}

func TestBuildPartyQualification_IncludesRepresentanteForPJ(t *testing.T) {
	data := map[string]any{
		"pj_1__tipo":         "Pessoa Juridica",
		"pj_1__razao_social": "58.132.597 REGINALDO POSPI DO NASCIMENTO JUNIOR",
		"pj_1__cnpj":         "58.132.597/0001-08",
		"pj_1__end__texto":   "Rua Benedito Augusto do Nascimento, n.o 30, Jardim Pilar, Maua/SP - CEP: 09370-060",
		"pj_1__rep_nome":     "Reginaldo Pospi do Nascimento Junior",
		"pj_1__rep_cpf":      "123.456.789-10",
	}

	qualification := buildPartyQualification(data, "pj_1")
	if !strings.Contains(qualification, "CNPJ n. 58.132.597/0001-08") {
		t.Fatalf("expected CNPJ in PJ qualification: %s", qualification)
	}
	if !strings.Contains(qualification, "neste ato representada por REGINALDO POSPI DO NASCIMENTO JUNIOR") {
		t.Fatalf("expected representative name in PJ qualification: %s", qualification)
	}
	if !strings.Contains(qualification, "CPF n. 123.456.789-10") {
		t.Fatalf("expected representative CPF in PJ qualification: %s", qualification)
	}
	if !strings.Contains(qualification, "na forma de sua situacao cadastral de pessoa juridica da Receita Federal ou contrato social") {
		t.Fatalf("expected representative legal basis text in PJ qualification: %s", qualification)
	}
}

func TestBuildPreview_FullTextIncludesRepresentanteForPJ(t *testing.T) {
	svc := NewService()
	preview := svc.BuildPreview("0001", "Compromisso de Venda e Compra de Imovel", map[string]any{
		"compradores":        []any{"comprador_1"},
		"comprador_1__tipo":  "Pessoa Juridica",
		"comprador_1__razao_social": "58.132.597 REGINALDO POSPI DO NASCIMENTO JUNIOR",
		"comprador_1__cnpj":  "58.132.597/0001-08",
		"comprador_1__end__texto": "Rua Benedito Augusto do Nascimento, n.o 30, Jardim Pilar, Maua/SP - CEP: 09370-060",
		"comprador_1__rep_nome": "Reginaldo Pospi do Nascimento Junior",
		"comprador_1__rep_cpf":  "123.456.789-10",
	})

	if !strings.Contains(preview.FullText, "neste ato representada por REGINALDO POSPI DO NASCIMENTO JUNIOR") {
		t.Fatalf("expected representative in full contract preview: %s", preview.FullText)
	}
}
