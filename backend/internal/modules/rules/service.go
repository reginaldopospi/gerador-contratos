package rules

import (
	"fmt"
	"strings"
)

type Preview struct {
	Title    string   `json:"title"`
	Sections []string `json:"sections"`
}

type Service struct {
	deliveryDefaults map[string]string
}

func NewService() *Service {
	return &Service{
		deliveryDefaults: map[string]string{
			"30 dias apos credito em conta": "Em ate 30 dias corridos apos o valor total do imovel ser creditado na conta da parte vendedora.",
			"30 dias apos assinatura no Banco": "Em ate 30 dias corridos apos assinatura da escritura definitiva perante instituicao financeira.",
			"30 dias apos assinatura do CCV": "Em ate 30 dias corridos apos assinatura da parte compradora no presente instrumento.",
			"No ato da assinatura no Banco": "No ato da assinatura da escritura definitiva perante instituicao financeira.",
			"No ato da assinatura do CCV": "No ato da assinatura da parte compradora no presente instrumento.",
			"24 horas do credito em conta": "Em ate 24 horas apos o valor total do imovel ser creditado na conta da parte vendedora.",
			"Escrever no contrato": "Texto livre definido manualmente no contrato.",
		},
	}
}

func (s *Service) BuildPreview(numero, tipo string, data map[string]any) Preview {
	title := s.tipoJuridicoContrato(tipo, getString(data, "preco_financiamento"))

	sellerRole := s.papelParteVendedoraOuCedente(tipo)
	buyerRole := s.papelParteCompradoraOuCessionaria(tipo)

	vendedores := collectPartyNames(data, "vendedores")
	compradores := collectPartyNames(data, "compradores")
	if len(vendedores) == 0 {
		vendedores = []string{"(nao informado)"}
	}
	if len(compradores) == 0 {
		compradores = []string{"(nao informado)"}
	}

	objeto := buildObjeto(data)
	pagamento := buildPagamento(data)
	entrega := s.textoEntregaChaves(data)
	comissao := buildComissao(data)

	sections := []string{
		fmt.Sprintf("Contrato %s - Numero %s", title, valueOrFallback(numero, "(sem numero)")),
		fmt.Sprintf("%s: %s", sellerRole, strings.Join(vendedores, "; ")),
		fmt.Sprintf("%s: %s", buyerRole, strings.Join(compradores, "; ")),
		"Objeto: " + objeto,
		"Pagamento: " + pagamento,
		"Entrega de chaves: " + entrega,
		"Comissao e intermediacao: " + comissao,
	}

	return Preview{Title: title, Sections: sections}
}

func (s *Service) tipoJuridicoContrato(tipo, financiamento string) string {
	tipo = strings.TrimSpace(tipo)
	lower := strings.ToLower(tipo)
	if strings.Contains(lower, "cessao") || strings.Contains(lower, "posse") {
		if tipo == "" {
			return "Cessao de Posse e Direitos sobre Imovel"
		}
		return tipo
	}
	if strings.TrimSpace(financiamento) != "" {
		return "Compromisso de Venda e Compra de Imovel com Financiamento"
	}
	return "Compromisso de Compra e Venda de Imovel"
}

func (s *Service) papelParteVendedoraOuCedente(tipo string) string {
	lower := strings.ToLower(strings.TrimSpace(tipo))
	if strings.Contains(lower, "cessao") || strings.Contains(lower, "posse") {
		return "PARTE CEDENTE"
	}
	return "PARTE VENDEDORA"
}

func (s *Service) papelParteCompradoraOuCessionaria(tipo string) string {
	lower := strings.ToLower(strings.TrimSpace(tipo))
	if strings.Contains(lower, "cessao") || strings.Contains(lower, "posse") {
		return "PARTE CESSIONARIA"
	}
	return "PARTE COMPRADORA"
}

func (s *Service) textoEntregaChaves(data map[string]any) string {
	escolha := strings.TrimSpace(getString(data, "entrega_chaves"))
	if escolha == "" {
		return "(nao informado)"
	}
	if strings.EqualFold(escolha, "Escrever no contrato") {
		manual := strings.TrimSpace(getString(data, "entrega_chaves_texto"))
		if manual != "" {
			return manual
		}
	}

	customMap := getMap(data, "clausulas_entrega_chaves")
	if len(customMap) > 0 {
		if v, ok := customMap[escolha]; ok {
			if txt := strings.TrimSpace(fmt.Sprintf("%v", v)); txt != "" {
				return txt
			}
		}
	}

	if txt, ok := s.deliveryDefaults[escolha]; ok {
		return txt
	}
	return escolha
}

func collectPartyNames(data map[string]any, listKey string) []string {
	items := getSlice(data, listKey)
	out := make([]string, 0, len(items))
	for _, item := range items {
		prefix := strings.TrimSpace(fmt.Sprintf("%v", item))
		if prefix == "" {
			continue
		}
		nome := strings.TrimSpace(getString(data, prefix+"__nome"))
		if nome == "" {
			nome = strings.TrimSpace(getString(data, prefix+"__razao_social"))
		}
		if nome != "" {
			out = append(out, nome)
		}
	}
	return out
}

func buildObjeto(data map[string]any) string {
	tipo := valueOrFallback(getString(data, "imovel__tipo"), "imovel")
	endereco := valueOrFallback(getString(data, "imovel__end__texto"), "endereco nao informado")
	matricula := strings.TrimSpace(getString(data, "imovel__matricula"))
	cartorio := strings.TrimSpace(getString(data, "imovel__cartorio"))

	parts := []string{fmt.Sprintf("%s localizado em %s", tipo, endereco)}
	if matricula != "" {
		parts = append(parts, "matricula "+matricula)
	}
	if cartorio != "" {
		parts = append(parts, "cartorio "+cartorio)
	}
	return strings.Join(parts, ", ")
}

func buildPagamento(data map[string]any) string {
	fields := []struct {
		Label string
		Key   string
	}{
		{"Preco total", "preco_total"},
		{"Financiamento", "preco_financiamento"},
		{"FGTS", "preco_fgts"},
		{"Entrada", "preco_entrada"},
		{"Sinal", "preco_sinal"},
		{"Recurso proprio", "preco_recurso_proprio"},
		{"Carta de credito", "preco_carta_credito"},
		{"Subsidio", "preco_subsidio"},
		{"Parcelamento", "preco_parcelamento_total"},
		{"Outros", "preco_outros"},
	}

	items := make([]string, 0)
	for _, f := range fields {
		v := strings.TrimSpace(getString(data, f.Key))
		if v == "" {
			continue
		}
		items = append(items, fmt.Sprintf("%s: %s", f.Label, v))
	}
	if len(items) == 0 {
		return "(nao informado)"
	}
	return strings.Join(items, " | ")
}

func buildComissao(data map[string]any) string {
	quemPaga := valueOrFallback(getString(data, "quem_paga_comissao"), "(nao informado)")
	valor := valueOrFallback(getString(data, "valor_comissao"), "(nao informado)")
	momento := valueOrFallback(getString(data, "momento_pagto"), "(nao informado)")
	return fmt.Sprintf("Pagador: %s; Valor: %s; Momento: %s", quemPaga, valor, momento)
}

func getString(data map[string]any, key string) string {
	if data == nil {
		return ""
	}
	v, ok := data[key]
	if !ok || v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

func getSlice(data map[string]any, key string) []any {
	if data == nil {
		return nil
	}
	raw, ok := data[key]
	if !ok || raw == nil {
		return nil
	}

	switch v := raw.(type) {
	case []any:
		return v
	case []string:
		out := make([]any, len(v))
		for i := range v {
			out[i] = v[i]
		}
		return out
	default:
		return nil
	}
}

func getMap(data map[string]any, key string) map[string]any {
	if data == nil {
		return nil
	}
	raw, ok := data[key]
	if !ok || raw == nil {
		return nil
	}
	if typed, ok := raw.(map[string]any); ok {
		return typed
	}
	return nil
}

func valueOrFallback(v, fallback string) string {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	return v
}
