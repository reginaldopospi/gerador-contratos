package rules

import (
	"fmt"
	"strings"
	"time"
)

func (s *Service) buildFullContract(numero, tipo string, data map[string]any) string {
	title := s.tipoJuridicoContrato(tipo, getString(data, "preco_financiamento"))
	sellerRole := s.papelParteVendedoraOuCedente(tipo)
	buyerRole := s.papelParteCompradoraOuCessionaria(tipo)

	parts := []string{
		strings.ToUpper(valueOrFallback(title, "CONTRATO")),
		fmt.Sprintf("Numero do contrato: %s", valueOrFallback(strings.TrimSpace(numero), "(nao informado)")),
		"",
		"QUADRO RESUMO",
		"DAS PARTES",
		fmt.Sprintf("Adiante simplesmente designado como %s:", sellerRole),
		valueOrFallback(strings.Join(buildPartyQualifications(data, "vendedores"), "\n\n"), "(nao informado)"),
		fmt.Sprintf("Adiante simplesmente designado como %s:", buyerRole),
		valueOrFallback(strings.Join(buildPartyQualifications(data, "compradores"), "\n\n"), "(nao informado)"),
		"DA INTERMEDIADORA",
		intermediadoraPadraoText(),
		"DO OBJETO DO CONTRATO",
		s.buildObjetoCompleto(data),
		"DAS CLAUSULAS E CONDICOES",
		s.clausulaPreambulo(data),
	}

	parts = append(parts, s.buildClauseLines(tipo, data)...)

	return strings.Join(filterNonEmpty(applyIndexedClauses(parts, collectIndexedClauses(data))...), "\n\n")
}

func (s *Service) buildObjetoCompleto(data map[string]any) string {
	tipoImovel := valueOrFallback(strings.TrimSpace(getString(data, "imovel__tipo")), "imovel")
	endereco := valueOrFallback(strings.TrimSpace(getString(data, "imovel__end__texto")), "endereco nao informado")
	matricula := strings.TrimSpace(getString(data, "imovel__matricula"))
	cartorio := strings.TrimSpace(getString(data, "imovel__cartorio"))
	comarca := strings.TrimSpace(getString(data, "imovel__cidade_cartorio"))
	descricaoMatricula := strings.TrimSpace(getString(data, "imovel__descricao_matricula"))
	codigoContribuinte := strings.TrimSpace(getString(data, "imovel__contribuinte"))
	precoTotal := strings.TrimSpace(getString(data, "preco_total"))
	prazoEntrega := strings.TrimSpace(s.textoEntregaChaves(data))
	pagamentos := buildResumoPagamentoItems(data)

	// Mantem o quadro resumo com a mesma hierarquia do modelo juridico solicitado.
	lines := []string{
		"Adiante simplesmente designado como IMOVEL:",
		buildResumoObjetoPrincipal(tipoImovel, endereco, matricula, cartorio, comarca),
		"",
		"IMOVEL: " + valueOrFallback(descricaoMatricula, "(nao informado)"),
		"CODIGO DE CONTRIBUINTE: " + valueOrFallback(codigoContribuinte, "(nao informado)"),
		"",
		"DO VALOR DO IMOVEL:",
		ensureTrailingPeriod(valueOrFallback(precoTotal, "(nao informado)")),
		"",
		"DA FORMA DE PAGAMENTO DO PRECO:",
	}

	if len(pagamentos) == 0 {
		lines = append(lines, "(nao informado).")
	} else {
		for idx, item := range pagamentos {
			lines = append(lines, fmt.Sprintf("%s) %s", alphabeticalItemToken(idx), item))
		}
	}

	lines = append(lines,
		"",
		"DO PRAZO DE ENTREGA DAS CHAVES DO IMOVEL:",
		ensureTrailingPeriod(valueOrFallback(prazoEntrega, "(nao informado)")),
	)

	return strings.Join(lines, "\n")
}

// Monta a linha principal do item 01 do objeto com dados de matricula/cartorio quando disponiveis.
func buildResumoObjetoPrincipal(tipoImovel, endereco, matricula, cartorio, comarca string) string {
	base := fmt.Sprintf("01 (um) %s situado em %s", tipoImovel, endereco)
	detalhes := []string{}

	if matricula != "" {
		detalhes = append(detalhes, "Matricula sob o n.o "+matricula)
	}
	if cartorio != "" {
		textoCartorio := cartorio + " Cartorio de Registro de Imoveis"
		if comarca != "" {
			textoCartorio += " de " + comarca
		}
		detalhes = append(detalhes, textoCartorio)
	}

	if len(detalhes) == 0 {
		return base + "."
	}
	return base + ". " + strings.Join(detalhes, ", ") + "."
}

// Lista os pagamentos informados no formulario para preencher os itens a), b), c) do quadro resumo.
func buildResumoPagamentoItems(data map[string]any) []string {
	fields := []struct {
		key   string
		label string
	}{
		{key: "preco_sinal", label: "sinal"},
		{key: "preco_entrada", label: "entrada"},
		{key: "preco_financiamento", label: "financiamento"},
		{key: "preco_fgts", label: "FGTS"},
		{key: "preco_recurso_proprio", label: "recurso proprio"},
		{key: "preco_carta_credito", label: "carta de credito"},
		{key: "preco_subsidio", label: "subsidio"},
		{key: "preco_parcelamento_total", label: "parcelamento"},
		{key: "preco_outros", label: "outros valores"},
	}

	items := make([]string, 0, len(fields))
	for _, field := range fields {
		value := strings.TrimSpace(getString(data, field.key))
		if value == "" {
			continue
		}
		items = append(items, ensureTrailingPeriod(fmt.Sprintf("%s referente a %s", value, field.label)))
	}
	return items
}

// Gera identificadores alfabeticos dos itens de pagamento mantendo fallback numerico para listas longas.
func alphabeticalItemToken(index int) string {
	if index >= 0 && index < 26 {
		return string(rune('a' + index))
	}
	return fmt.Sprintf("%d", index+1)
}

// Padroniza finalizacao de sentencas do quadro resumo para evitar pontuacao duplicada.
func ensureTrailingPeriod(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	lastChar := trimmed[len(trimmed)-1]
	if lastChar == '.' || lastChar == ':' || lastChar == ';' {
		return trimmed
	}
	return trimmed + "."
}

func (s *Service) clausulaPreambulo(data map[string]any) string {
	destino := "tabeliao de notas competente"
	if strings.TrimSpace(getString(data, "preco_financiamento")) != "" {
		destino = "instituicao financeira competente"
	}

	return fmt.Sprintf(
		"As partes qualificadas no quadro resumo pactuam entre si o presente compromisso de compra e venda do IMOVEL, o qual sera oportunamente aperfeicoado mediante instrumento celebrado perante %s, mediante as seguintes clausulas e condicoes, a saber:",
		destino,
	)
}

func (s *Service) buildClauseLines(tipo string, data map[string]any) []string {
	hasFinancing := strings.TrimSpace(getString(data, "preco_financiamento")) != ""
	hasFGTS := strings.TrimSpace(getString(data, "preco_fgts")) != ""
	hasSignal := strings.TrimSpace(getString(data, "preco_sinal")) != ""
	hasInstallments := strings.TrimSpace(getString(data, "preco_parcelamento_total")) != ""
	isAlienado := strings.EqualFold(strings.TrimSpace(getString(data, "imovel__alienado")), "SIM")
	isMatriculaAreaMaior := strings.Contains(strings.ToLower(strings.TrimSpace(getString(data, "imovel__tipo"))), "matricula em area maior")

	lines := []string{
		"1. DAS DECLARACOES INICIAIS",
		"1.1 A PARTE VENDEDORA declara que e proprietaria e legitima possuidora do IMOVEL com justo titulo, livre e desembaracado de onus ou gravame, judicial ou extrajudicial.",
		"1.2 Na hipotese de haver apontamento de acao, execucao judicial, protesto ou debitos relativos ao IMOVEL, a PARTE VENDEDORA devera prestar esclarecimentos e documentos suficientes para manter a seguranca juridica da transacao.",
	}
	if isMatriculaAreaMaior {
		lines = append(lines,
			"1.3 Em se tratando de unidade com matricula em area maior, a PARTE VENDEDORA providenciara abertura da matricula individual no prazo aproximado de ate 90 dias.",
			"1.4 Se, por caso fortuito ou forca maior, a regularizacao da matricula nao ocorrer no prazo previsto, o negocio podera ser resilido sem multa entre as partes, com devolucao de valores eventualmente recebidos no prazo contratual.",
		)
	}
	if isAlienado {
		if hasFinancing {
			lines = append(lines, "1.5 A PARTE COMPRADORA declara ciencia de que o IMOVEL esta alienado fiduciariamente e que a quitacao do saldo da PARTE VENDEDORA sera realizada no fluxo de financiamento da PARTE COMPRADORA.")
		} else {
			lines = append(lines, "1.5 A PARTE COMPRADORA declara ciencia de que o IMOVEL esta alienado fiduciariamente, e que a quitacao ocorrera conforme os valores previstos neste instrumento.")
		}
	}

	lines = append(lines,
		"2. DO PRECO E FORMA DE PAGAMENTO",
		"2.1 A PARTE VENDEDORA compromete-se a transferir a propriedade do IMOVEL mediante recebimento de preco certo, liquido e exigivel, conforme quadro resumo.",
	)
	if hasInstallments {
		lines = append(lines,
			"2.2 Havendo parcelamento, as parcelas serao pagas por transferencia bancaria nas datas de vencimento acordadas entre as partes.",
			"2.3 Em caso de mora nas parcelas, incidirao multa, juros e honorarios de cobranca conforme previsao legal e contratual.",
		)
	}
	if hasSignal {
		lines = append(lines,
			"2.4 Em caso de inadimplemento da PARTE VENDEDORA, os valores pagos a titulo de sinal poderao ser restituidos em dobro, nos termos dos artigos 417 a 419 do Codigo Civil.",
			"2.5 Em caso de inadimplemento da PARTE COMPRADORA, o sinal podera ser retido pela PARTE VENDEDORA nos termos legais.",
		)
	}

	lines = append(lines, "3. DA ESCRITURA DEFINITIVA")
	if hasFinancing {
		lines = append(lines,
			"3.1 O presente contrato sera aperfeicoado por instrumento perante instituicao financeira competente.",
			"3.2 As partes se obrigam a comparecer no ato de assinatura da escritura definitiva na data acordada.",
			"3.3 A inadimplencia da PARTE COMPRADORA em promover a lavratura da escritura no prazo pactuado isenta a PARTE VENDEDORA de apresentar novas certidoes.",
			"3.4 A PARTE COMPRADORA se obriga a protocolar o registro da escritura definitiva em ate 48 horas da sua posse do documento.",
		)
	} else {
		lines = append(lines,
			"3.1 O presente contrato sera aperfeicoado por instrumento perante tabeliao de notas competente.",
			"3.2 As partes se obrigam a comparecer no ato de assinatura da escritura definitiva na data acordada.",
			"3.3 A inadimplencia da PARTE COMPRADORA em promover a lavratura da escritura no prazo pactuado isenta a PARTE VENDEDORA de apresentar novas certidoes.",
		)
	}

	if hasFinancing || hasFGTS {
		lines = append(lines,
			"4. DO FINANCIAMENTO E/OU LIBERACAO DO FGTS",
			"4.1 A PARTE COMPRADORA declara ter recebido esclarecimentos sobre documentacao, exigencias operacionais, custos de escrituracao e encargos de ITBI, custas e emolumentos.",
			"4.2 A PARTE COMPRADORA declara ciencia das condicoes de concessao do credito e/ou liberacao de FGTS, assumindo os onus de eventuais alteracoes supervenientes das regras aplicaveis.",
			"4.3 A instituicao financeira podera, ao seu juizo, nao conceder os valores pretendidos se os requisitos juridicos ou economicos nao forem atendidos.",
		)
	}

	lines = append(lines,
		"5. DA ENTREGA DAS CHAVES E DAS CONTAS DE CONSUMO",
		"5.1 A PARTE VENDEDORA se obriga a entregar as chaves e comprovantes de quitacao das contas de consumo nos prazos ajustados.",
		"5.2 A PARTE VENDEDORA entregara o IMOVEL livre e desocupado, arcando com consumos e encargos ate a efetiva entrega.",
		"6. DAS TRANSFERENCIAS PERANTE PREFEITURA E CONCESSIONARIAS",
		"6.1 A PARTE COMPRADORA se obriga a transferir titularidade das contas de consumo no prazo contratual.",
		"6.2 A PARTE COMPRADORA se obriga a providenciar a transferencia de IPTU no prazo contratual e legal.",
		"7. DO PAGAMENTO DOS HONORARIOS DA INTERMEDIADORA",
	)

	quemPagaComissao := strings.TrimSpace(getString(data, "quem_paga_comissao"))
	if quemPagaComissao != "" {
		lines = append(lines, "7.1 O pagamento da comissao de intermediacao observara o pagador definido no quadro resumo: "+quemPagaComissao+".")
	} else {
		lines = append(lines, "7.1 O pagamento da comissao de intermediacao seguira o contrato de corretagem firmado entre as partes aplicaveis.")
	}
	lines = append(lines, "7.2 Em caso de inadimplemento da comissao, poderao incidir cobrancas e honorarios advocaticios conforme previsao contratual.")

	if !hasInstallments {
		lines = append(lines,
			"8. DO PRAZO DE VALIDADE DO INSTRUMENTO E SUA CONCLUSAO",
		)
		if hasFinancing || hasFGTS {
			lines = append(lines, "8.1 O prazo de validade do instrumento para conclusao sera de ate 120 dias, prorrogavel nos termos deste contrato.")
		} else {
			lines = append(lines, "8.1 O prazo de validade do instrumento para conclusao sera de ate 60 dias, prorrogavel nos termos deste contrato.")
		}
		lines = append(lines,
			"8.2 Sem conclusao no prazo e sem culpa das partes, podera haver resilicao sem multa entre parte vendedora e compradora.",
			"8.3 Havendo valores recebidos pela PARTE VENDEDORA, deverao ser restituidos no prazo contratual se houver extincao por resilicao.",
		)
	}

	lines = append(lines,
		"9. DA RESOLUCAO CONTRATUAL",
	)
	if hasSignal {
		lines = append(lines,
			"9.1 Em caso de lesao por inadimplemento culposo, a parte inocente podera pedir resolucao e indenizacao, considerando as arras como minimo indenizatorio.",
			"9.2 Se a desistencia for da PARTE VENDEDORA, havera devolucao integral dos valores pagos e demais encargos previstos.",
			"9.3 Se a desistencia for da PARTE COMPRADORA, aplicam-se os efeitos de perda do sinal e demais encargos previstos.",
		)
	} else {
		lines = append(lines, "9.1 Em caso de inadimplemento culposo, a parte inocente podera pedir resolucao ou cumprimento, com multa e perdas e danos.")
	}

	lines = append(lines,
		"10. DA IRRETRATABILIDADE",
		"10.1 O presente contrato e celebrado em carater irretratavel e irrevogavel, salvo hipoteses expressas de extincao previstas neste instrumento.",
		"11. DA EVICCAO DE DIREITO E VICIOS REDIBITORIOS",
		"11.1 A PARTE VENDEDORA responde por eviccao de direito e vicios redibitorios originados antes da transacao.",
		"11.2 Eventuais constricoes ou anulacoes que atinjam o IMOVEL asseguram a PARTE COMPRADORA o direito de recomposicao e perdas e danos.",
	)

	ficaraBens := strings.ToUpper(strings.TrimSpace(getString(data, "imovel__ficara_bens")))
	if ficaraBens == "SIM" {
		lines = append(lines,
			"12. DAS DECLARACOES DAS PARTES EM RELACAO AO IMOVEL",
			"12.1 A PARTE COMPRADORA declara que visitou o IMOVEL e o aceita no estado em que se encontra.",
			"12.2 A venda e feita na forma ad corpus, independentemente de medidas.",
			"12.3 Permanecerao vinculados ao IMOVEL os seguintes bens: "+valueOrFallback(strings.TrimSpace(getString(data, "imovel__bens")), "(nao informado)")+".",
		)
	} else {
		lines = append(lines,
			"12. DA DECLARACAO DA PARTE COMPRADORA EM RELACAO AS CONDICOES DO IMOVEL",
			"12.1 A PARTE COMPRADORA declara que visitou o IMOVEL e o aceita no estado em que se encontra.",
			"12.2 A venda e feita na forma ad corpus, independentemente de medidas.",
		)
	}

	lines = append(lines,
		"13. DO TERMINO DA PRESTACAO DE SERVICO DA INTERMEDIADORA",
		"13.1 A prestacao de servico da INTERMEDIADORA se aperfeicoa com a assinatura deste instrumento, sem prejuizo de orientacao operacional acessoria.",
		"13.2 A INTERMEDIADORA prestou esclarecimentos e suporte tecnico-juridico sobre a documentacao apresentada pelas partes.",
		"14. DAS DISPOSICOES GERAIS",
		"14.1 Eventual registro deste compromisso no cartorio competente correra por conta da PARTE COMPRADORA.",
		"14.2 Alteracoes deste instrumento exigem aditamento formal assinado pelas partes e testemunhas.",
		"14.3 Comunicacoes entre as partes devem observar os enderecos e contatos informados no contrato.",
		"15. ELEICAO DO FORO",
		"15.1 Fica eleito o foro da situacao do IMOVEL, com renuncia a qualquer outro, por mais privilegiado que seja.",
		"15.2 Por estarem justas e contratadas, as partes assinam o presente instrumento em 03 (tres) vias de igual teor e forma.",
		fmt.Sprintf("15.3 %s", lineLocalData(data, time.Now().UTC())),
	)

	lines = append(lines, signatureLines(data, s.papelParteVendedoraOuCedente(tipo), s.papelParteCompradoraOuCessionaria(tipo))...)
	return lines
}

func buildPartyQualifications(data map[string]any, listKey string) []string {
	refs := asStringList(data, listKey)
	if len(refs) == 0 {
		names := collectPartyNames(data, listKey)
		if len(names) == 0 {
			return nil
		}
		out := make([]string, 0, len(names))
		for _, name := range names {
			out = append(out, strings.ToUpper(strings.TrimSpace(name)))
		}
		return out
	}

	out := []string{}
	for _, ref := range refs {
		out = append(out, buildPartyQualification(data, ref))
	}
	return filterNonEmpty(out...)
}

func buildPartyQualification(data map[string]any, prefix string) string {
	isPJ := strings.Contains(strings.ToLower(strings.TrimSpace(getString(data, prefix+"__tipo"))), "juridica")
	if isPJ {
		razao := valueOrFallback(strings.TrimSpace(getString(data, prefix+"__razao_social")), strings.TrimSpace(getString(data, prefix+"__nome")))
		cnpj := strings.TrimSpace(getString(data, prefix+"__cnpj"))
		endereco := strings.TrimSpace(getString(data, prefix+"__end__texto"))
		repNome := strings.ToUpper(strings.TrimSpace(getString(data, prefix+"__rep_nome")))
		repCpf := strings.TrimSpace(getString(data, prefix+"__rep_cpf"))

		parts := []string{
			strings.ToUpper(razao),
			ternaryStr(cnpj != "", "CNPJ n. "+cnpj, ""),
			ternaryStr(endereco != "", "com sede em "+endereco, ""),
		}
		if repNome != "" {
			// Segue o texto-base do codigo original para representar legalmente a PJ no contrato.
			representacao := "neste ato representada por " + repNome
			if repCpf != "" {
				representacao += ", CPF n. " + repCpf
			}
			representacao += ", na forma de sua situacao cadastral de pessoa juridica da Receita Federal ou contrato social"
			parts = append(parts, representacao)
		}

		return strings.Join(filterNonEmpty(parts...), ", ")
	}

	nome := valueOrFallback(strings.TrimSpace(getString(data, prefix+"__nome")), strings.TrimSpace(getString(data, prefix+"__razao_social")))
	nacionalidade := strings.TrimSpace(getString(data, prefix+"__nacionalidade"))
	estadoCivil := strings.TrimSpace(getString(data, prefix+"__estado_civil"))
	regimeBens := strings.TrimSpace(getString(data, prefix+"__regime_bens"))
	profissao := strings.TrimSpace(getString(data, prefix+"__profissao"))
	rg := strings.TrimSpace(getString(data, prefix+"__rg"))
	cpf := strings.TrimSpace(getString(data, prefix+"__cpf"))
	endereco := strings.TrimSpace(getString(data, prefix+"__end__texto"))
	conjugeTexto := buildConjugeQualificationText(data, prefix)

	if requiresConjugeQualification(estadoCivil) {
		// Mantem bloco do conjuge/companheiro(a) sempre presente para casado(a)/uniao estavel.
		if strings.TrimSpace(conjugeTexto) == "" {
			conjugeTexto = "(dados do conjuge/companheiro(a) nao informado)"
		}
		return strings.Join(filterNonEmpty(
			strings.ToUpper(nome),
			nacionalidade,
			profissao,
			ternaryStr(rg != "", "RG n. "+rg, ""),
			ternaryStr(cpf != "", "CPF n. "+cpf, ""),
			"e "+conjugeTexto,
			buildConjugalStateText(estadoCivil, regimeBens),
			ternaryStr(endereco != "", "e residentes na "+endereco, ""),
		), ", ")
	}

	return strings.Join(filterNonEmpty(
		strings.ToUpper(nome),
		nacionalidade,
		estadoCivil,
		profissao,
		ternaryStr(rg != "", "RG n. "+rg, ""),
		ternaryStr(cpf != "", "CPF n. "+cpf, ""),
		ternaryStr(endereco != "", "residente em "+endereco, ""),
	), ", ")
}

// Regras de qualificacao exigem cÃ´njuge/companheiro(a) em casado(a) ou uniao estavel.
func requiresConjugeQualification(estadoCivil string) bool {
	normalized := strings.ToLower(strings.TrimSpace(estadoCivil))
	return normalized == "casado(a)" || normalized == "uniao estavel" || normalized == "uniÃ£o estÃ¡vel"
}

// Monta bloco textual da qualificacao do cÃ´njuge/companheiro(a).
func buildConjugeQualificationText(data map[string]any, prefix string) string {
	nome := strings.TrimSpace(getString(data, prefix+"__conj_nome"))
	nacionalidade := strings.TrimSpace(getString(data, prefix+"__conj_nacionalidade"))
	profissao := strings.TrimSpace(getString(data, prefix+"__conj_profissao"))
	rg := strings.TrimSpace(getString(data, prefix+"__conj_rg"))
	cpf := strings.TrimSpace(getString(data, prefix+"__conj_cpf"))

	if nome == "" && nacionalidade == "" && profissao == "" && rg == "" && cpf == "" {
		return ""
	}
	nomeExibicao := strings.ToUpper(valueOrFallback(nome, "(conjuge/companheiro(a) nao informado)"))
	return strings.Join(filterNonEmpty(
		nomeExibicao,
		nacionalidade,
		profissao,
		ternaryStr(rg != "", "RG n. "+rg, ""),
		ternaryStr(cpf != "", "CPF n. "+cpf, ""),
	), ", ")
}

// Gera o trecho conjugal no formato juridico esperado para casado(a)/uniao estavel.
func buildConjugalStateText(estadoCivil, regimeBens string) string {
	normalized := strings.ToLower(strings.TrimSpace(estadoCivil))
	base := "ambos casados entre si"
	if normalized == "uniao estavel" || normalized == "união estável" {
		base = "ambos conviventes em uniao estavel entre si"
	}
	regime := strings.TrimSpace(regimeBens)
	if regime == "" {
		return base
	}
	return base + " sob o regime de " + regime
}

func signatureLines(data map[string]any, sellerRole, buyerRole string) []string {
	sellers := collectSignatureNames(data, "vendedores")
	buyers := collectSignatureNames(data, "compradores")
	lines := []string{
		sellerRole + ":",
		buildSignatureBlock(sellers),
		buyerRole + ":",
		buildSignatureBlock(buyers),
		"TESTEMUNHAS:",
		"________________________________\nNome:\nCPF:",
		"________________________________\nNome:\nCPF:",
	}
	return lines
}

func buildSignatureBlock(names []string) string {
	if len(names) == 0 {
		return "________________________________\n(nao informado)"
	}

	chunks := []string{}
	for _, name := range names {
		chunks = append(chunks, "________________________________\n"+strings.ToUpper(name))
	}
	return strings.Join(chunks, "\n\n")
}

func collectSignatureNames(data map[string]any, listKey string) []string {
	names := collectPartyNames(data, listKey)
	out := make([]string, 0, len(names))
	for _, name := range names {
		clean := strings.TrimSpace(name)
		if clean != "" {
			out = append(out, clean)
		}
	}
	return uniqueStrings(out)
}

type indexedClause struct {
	Index   string
	Title   string
	Content string
}

func collectIndexedClauses(data map[string]any) []indexedClause {
	out := make([]indexedClause, 0)
	known := map[string]struct{}{}

	for _, item := range getMapArray(data, "clausulas_selecionadas_vinculos") {
		clauseKey := strings.TrimSpace(getString(item, "clause_key"))
		if clauseKey == "" {
			continue
		}

		title := strings.TrimSpace(getString(item, "title"))
		if title == "" {
			title = strings.TrimSpace(getString(item, "titulo"))
		}
		if title == "" {
			title = clauseKey
		}

		content := strings.TrimSpace(getString(item, "content"))
		if content == "" {
			content = strings.TrimSpace(getString(item, "conteudo"))
		}
		index := sanitizeClauseIndex(getString(item, "indice"))
		if index == "" {
			index = sanitizeClauseIndex(getString(item, "index"))
		}

		key := clauseKey + "@" + index
		if _, ok := known[key]; ok {
			continue
		}
		known[key] = struct{}{}
		known[clauseKey+"@"] = struct{}{}

		out = append(out, indexedClause{
			Index:   index,
			Title:   title,
			Content: content,
		})
	}

	selected := append(asStringList(data, "clausulas_selecionadas"), asStringList(data, "clause_keys")...)
	selected = append(selected, asStringList(data, "clausulas_keys")...)
	for _, key := range selected {
		clauseKey := strings.TrimSpace(key)
		if clauseKey == "" {
			continue
		}

		composite := clauseKey + "@"
		if _, ok := known[composite]; ok {
			continue
		}
		known[composite] = struct{}{}

		out = append(out, indexedClause{
			Index:   "",
			Title:   clauseKey,
			Content: "",
		})
	}

	for _, item := range getMapArray(data, "clausulas_customizadas") {
		title := strings.TrimSpace(getString(item, "titulo"))
		if title == "" {
			title = strings.TrimSpace(getString(item, "title"))
		}
		content := strings.TrimSpace(getString(item, "conteudo"))
		if content == "" {
			content = strings.TrimSpace(getString(item, "content"))
		}
		index := sanitizeClauseIndex(getString(item, "indice"))
		if index == "" {
			index = sanitizeClauseIndex(getString(item, "index"))
		}

		if title == "" && content == "" {
			continue
		}

		out = append(out, indexedClause{
			Index:   index,
			Title:   title,
			Content: content,
		})
	}

	return out
}

func applyIndexedClauses(baseLines []string, clauses []indexedClause) []string {
	if len(clauses) == 0 {
		return baseLines
	}

	indexedByParent := make(map[string][]indexedClause)
	unindexed := make([]indexedClause, 0)

	for _, clause := range clauses {
		index := sanitizeClauseIndex(clause.Index)
		if index == "" {
			unindexed = append(unindexed, clause)
			continue
		}

		parent := parentIndex(index)
		if parent == "" {
			unindexed = append(unindexed, clause)
			continue
		}

		clause.Index = index
		indexedByParent[parent] = append(indexedByParent[parent], clause)
	}

	for parent := range indexedByParent {
		items := indexedByParent[parent]
		sortIndexedClauses(items)
		indexedByParent[parent] = items
	}

	out := make([]string, 0, len(baseLines)+len(clauses)+4)
	for _, line := range baseLines {
		out = append(out, line)

		token := lineNumberToken(line)
		if token == "" {
			continue
		}

		items, ok := indexedByParent[token]
		if !ok {
			continue
		}

		for _, clause := range items {
			out = append(out, renderIndexedClause(clause))
		}
		delete(indexedByParent, token)
	}

	for _, leftovers := range indexedByParent {
		unindexed = append(unindexed, leftovers...)
	}

	if len(unindexed) > 0 {
		out = append(out, "CLAUSULAS ADICIONAIS")
		for i, clause := range unindexed {
			out = append(out, fmt.Sprintf("A.%d %s", i+1, renderUnindexedClause(clause)))
		}
	}

	return out
}

func renderIndexedClause(clause indexedClause) string {
	title := strings.TrimSpace(clause.Title)
	content := strings.TrimSpace(clause.Content)
	if content != "" {
		return fmt.Sprintf("%s %s", clause.Index, content)
	}
	if title == "" {
		title = "CLAUSULA ADICIONAL"
	}
	return fmt.Sprintf("%s %s", clause.Index, title)
}

func renderUnindexedClause(clause indexedClause) string {
	title := strings.TrimSpace(clause.Title)
	content := strings.TrimSpace(clause.Content)
	if content != "" {
		return content
	}
	if title == "" {
		title = "CLAUSULA ADICIONAL"
	}
	return title
}

func parentIndex(index string) string {
	clean := sanitizeClauseIndex(index)
	if clean == "" {
		return ""
	}
	parts := strings.Split(clean, ".")
	if len(parts) <= 1 {
		return clean
	}
	return strings.Join(parts[:len(parts)-1], ".")
}

func sanitizeClauseIndex(value string) string {
	clean := strings.TrimSpace(value)
	clean = strings.TrimSuffix(clean, ".")
	clean = strings.ReplaceAll(clean, " ", "")
	if clean == "" {
		return ""
	}

	parts := strings.Split(clean, ".")
	valid := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if !isDigits(part) {
			return ""
		}
		valid = append(valid, part)
	}
	if len(valid) == 0 {
		return ""
	}
	return strings.Join(valid, ".")
}

func lineNumberToken(line string) string {
	fields := strings.Fields(strings.TrimSpace(line))
	if len(fields) == 0 {
		return ""
	}
	token := strings.TrimSpace(fields[0])
	token = strings.TrimSuffix(token, ".")
	return sanitizeClauseIndex(token)
}

func sortIndexedClauses(items []indexedClause) {
	if len(items) < 2 {
		return
	}

	for i := 0; i < len(items)-1; i += 1 {
		for j := i + 1; j < len(items); j += 1 {
			if compareClauseIndex(items[i].Index, items[j].Index) <= 0 {
				continue
			}
			items[i], items[j] = items[j], items[i]
		}
	}
}

func compareClauseIndex(a, b string) int {
	aa := strings.Split(a, ".")
	bb := strings.Split(b, ".")
	size := len(aa)
	if len(bb) > size {
		size = len(bb)
	}

	for i := 0; i < size; i += 1 {
		av := -1
		bv := -1
		if i < len(aa) {
			av = parseIntSafe(aa[i])
		}
		if i < len(bb) {
			bv = parseIntSafe(bb[i])
		}
		if av < bv {
			return -1
		}
		if av > bv {
			return 1
		}
	}
	return 0
}

func parseIntSafe(v string) int {
	value := 0
	for _, ch := range v {
		if ch < '0' || ch > '9' {
			return value
		}
		value = value*10 + int(ch-'0')
	}
	return value
}

func isDigits(v string) bool {
	if v == "" {
		return false
	}
	for _, ch := range v {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}

func getMapArray(data map[string]any, key string) []map[string]any {
	raw := getSlice(data, key)
	if len(raw) == 0 {
		return nil
	}

	out := make([]map[string]any, 0, len(raw))
	for _, item := range raw {
		mapped, ok := item.(map[string]any)
		if ok {
			out = append(out, mapped)
		}
	}
	return out
}

func asStringList(data map[string]any, key string) []string {
	raw := getSlice(data, key)
	out := make([]string, 0, len(raw))
	for _, item := range raw {
		clean := strings.TrimSpace(fmt.Sprintf("%v", item))
		if clean != "" {
			out = append(out, clean)
		}
	}
	return uniqueStrings(out)
}

func lineLocalData(data map[string]any, now time.Time) string {
	cidade := strings.TrimSpace(getString(data, "imovel__end__cidade"))
	uf := strings.TrimSpace(getString(data, "imovel__end__uf"))
	dataTxt := dataPorExtenso(now)
	if cidade != "" && uf != "" {
		return fmt.Sprintf("%s/%s, %s.", cidade, uf, dataTxt)
	}
	if cidade != "" {
		return fmt.Sprintf("%s, %s.", cidade, dataTxt)
	}
	return dataTxt + "."
}

func dataPorExtenso(now time.Time) string {
	meses := []string{"janeiro", "fevereiro", "marco", "abril", "maio", "junho", "julho", "agosto", "setembro", "outubro", "novembro", "dezembro"}
	month := int(now.Month())
	if month < 1 || month > 12 {
		month = 1
	}
	return fmt.Sprintf("%02d de %s de %d", now.Day(), meses[month-1], now.Year())
}

func intermediaryAddressFallback() string {
	return "Rua Roberto, n. 14, Jardim Santa Mena, Guarulhos/SP - CEP: 07096-070"
}

func intermediadoraPadraoText() string {
	return "IMOBILIARIA MONTE SIAO LTDA, pessoa juridica de direito privado, CNPJ n. 30.177.724/0001-76, CRECI n. 33.150-J, com sede na " + intermediaryAddressFallback() + ", representada por JOSIVAN MOURA DA SILVA, brasileiro, corretor de imoveis, RG n. 55.786.890-7 SSP, CPF n. 343.173.968-74."
}

func filterNonEmpty(items ...string) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		clean := strings.TrimSpace(item)
		if clean != "" {
			out = append(out, clean)
		}
	}
	return out
}

func uniqueStrings(items []string) []string {
	seen := make(map[string]struct{}, len(items))
	out := make([]string, 0, len(items))
	for _, item := range items {
		clean := strings.TrimSpace(item)
		if clean == "" {
			continue
		}
		if _, ok := seen[clean]; ok {
			continue
		}
		seen[clean] = struct{}{}
		out = append(out, clean)
	}
	return out
}

func ternaryStr(cond bool, whenTrue, whenFalse string) string {
	if cond {
		return whenTrue
	}
	return whenFalse
}
