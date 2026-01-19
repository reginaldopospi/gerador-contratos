import streamlit as st

# ✅ TEM QUE SER O PRIMEIRO COMANDO STREAMLIT DO APP
st.set_page_config(page_title="Gerador de Contratos", page_icon="📄", layout="wide")

import re
import requests
from datetime import date

from supabase import create_client, Client
from typing import Optional

# ============================================================
# STATE HELPERS (BASE DO APP) - get / set_ / get_list
# ============================================================

def _ensure_dados():
    if "dados" not in st.session_state:
        st.session_state.dados = {}

def set_(key: str, value):
    _ensure_dados()
    st.session_state.dados[key] = value

def get_list(key: str) -> list:
    _ensure_dados()
    v = st.session_state.dados.get(key)
    if isinstance(v, list):
        return v
    if v is None:
        st.session_state.dados[key] = []
        return st.session_state.dados[key]
    # se tiver algum valor errado salvo, normaliza para lista vazia
    st.session_state.dados[key] = []
    return st.session_state.dados[key]

# ============================================================
# AUTH (LOGIN VIA STREAMLIT SECRETS) - APENAS 1 MÉTODO
# ============================================================

def auth_users() -> dict:
    """
    Retorna o dicionário de usuários/senhas definido em st.secrets.

    Formato esperado em Secrets (TOML):
    [auth]
    users = { reginaldo="senha", imobiliaria1="senha" }
    """
    try:
        users = st.secrets.get("auth", {}).get("users", {})
        return dict(users) if users else {}
    except Exception:
        return {}

def is_logged_in() -> bool:
    return bool(st.session_state.get("auth_ok", False))

def do_logout():
    st.session_state["auth_ok"] = False
    st.session_state["auth_user"] = ""
    st.rerun()

def render_login():
    st.title("🔐 Acesso restrito")
    st.caption("Digite seu usuário e senha para acessar o sistema.")

    users = auth_users()
    if not users:
        st.error("⚠️ Nenhum usuário configurado. Configure em Settings → Secrets no Streamlit Cloud.")
        st.stop()

    col1, col2 = st.columns(2)
    with col1:
        user = st.text_input("Usuário", key="login_user")
    with col2:
        pwd = st.text_input("Senha", type="password", key="login_pwd")

    if st.button("Entrar", key="btn_login"):
        user = (user or "").strip()
        pwd = (pwd or "").strip()

        if user in users and pwd == str(users[user]):
            st.session_state["auth_ok"] = True
            st.session_state["auth_user"] = user
            st.rerun()
        else:
            st.error("Usuário ou senha inválidos.")

# Inicializa sessão
if "auth_ok" not in st.session_state:
    st.session_state["auth_ok"] = False
if "auth_user" not in st.session_state:
    st.session_state["auth_user"] = ""

# Gate do app
if not is_logged_in():
    render_login()
    st.stop()

# ============================================================
# AUTH (Streamlit Secrets) - Usuário/Senha por imobiliária
# Secrets (Streamlit Cloud):
# [auth.users]
# monte_siao = "..."
# imobiliaria_x = "..."
# admin = "..."
# ============================================================

def auth_users() -> dict:
    """
    Lê os usuários/senhas do Streamlit Secrets.
    Retorna dict {usuario: senha}.
    """
    try:
        return dict(st.secrets.get("auth", {}).get("users", {}))
    except Exception:
        return {}

def validar_login(usuario: str, senha: str) -> bool:
    """
    Valida usuário e senha contra st.secrets['auth']['users'].
    """
    usuario = (usuario or "").strip()
    senha = (senha or "").strip()
    users = auth_users()
    return bool(usuario) and (users.get(usuario) == senha)


# ============================================================
# STATE
# ============================================================
if "step_index" not in st.session_state:
    st.session_state.step_index = 0

if "dados" not in st.session_state:
    st.session_state.dados = {}

def get(k, default=""):
    if "dados" not in st.session_state:
        st.session_state.dados = {}
    return st.session_state.dados.get(k, default)

def set_(k, v):
    if "dados" not in st.session_state:
        st.session_state.dados = {}
    st.session_state.dados[k] = v

def get_list(k):
    if "dados" not in st.session_state:
        st.session_state.dados = {}
    v = st.session_state.dados.get(k, [])
    if not isinstance(v, list):
        v = []
        st.session_state.dados[k] = v
    return v

def set_list(k, v):
    if "dados" not in st.session_state:
        st.session_state.dados = {}
    st.session_state.dados[k] = v


# ============================================================
# FLAGS DE TELAS OCULTAS
# ============================================================
if "admin_corretores_liberado" not in st.session_state:
    st.session_state.admin_corretores_liberado = False

# ✅ NOVO: flag para liberar acesso ao Admin de Cláusulas
if "admin_liberado" not in st.session_state:
    st.session_state.admin_liberado = False

# ============================================================
# FLAGS DE TELAS OCULTAS
# ============================================================

if "voltar_step_preco_chaves" not in st.session_state:
    st.session_state.voltar_step_preco_chaves = None

# ============================================================
# CONTROLE DA TELA "CADASTRO DE CORRETOR" (oculta no menu)
# ============================================================
if "cadastro_corretor_ativado" not in st.session_state.dados:
    st.session_state.dados["cadastro_corretor_ativado"] = False

if "cadastro_corretor_destino" not in st.session_state.dados:
    st.session_state.dados["cadastro_corretor_destino"] = ""

if "cadastro_corretor_prefix" not in st.session_state.dados:
    st.session_state.dados["cadastro_corretor_prefix"] = ""

# ============================================================
# SUPABASE (PERSISTÊNCIA) - CORRETORES (UNIFICADO)
# ============================================================

@st.cache_resource(show_spinner=False)
def _supabase() -> Optional["Client"]:
    """
    Cria cliente Supabase usando Secrets (aceita 2 formatos):

    Formato A (recomendado):
      supabase_url = "..."
      supabase_service_role_key = "..."

    Formato B (alternativo):
      [supabase]
      url = "..."
      service_role_key = "..."
    """
    try:
        url = (st.secrets.get("supabase_url") or "").strip()
        key = (st.secrets.get("supabase_service_role_key") or "").strip()

        if not url or not key:
            url = (st.secrets.get("supabase", {}).get("url") or "").strip()
            key = (st.secrets.get("supabase", {}).get("service_role_key") or "").strip()

        if not url or not key:
            return None

        return create_client(url, key)
    except Exception:
        return None


def _tenant_imobiliaria() -> str:
    """
    Isola os dados por imobiliária/usuário logado.
    """
    u = (st.session_state.get("auth_user", "") or "").strip()
    return u if u else "geral"


def _cache_key_corretores() -> str:
    return f"_corretores_loaded__{_tenant_imobiliaria()}"


def _carregar_corretores_supabase():
    """
    Carrega do Supabase para st.session_state.dados['corretores_cadastrados'].

    Espera tabela: corretores
    Colunas: id (uuid), imobiliaria (text), nome, cpf, banco, agencia, conta, pix
    """
    sb = _supabase()
    if sb is None:
        st.session_state.dados["corretores_cadastrados"] = st.session_state.dados.get("corretores_cadastrados", [])
        return

    tenant = _tenant_imobiliaria()

    try:
        res = (
            sb.table("corretores")
              .select("id, imobiliaria, nome, cpf, banco, agencia, conta, pix")
              .eq("imobiliaria", tenant)
              .order("nome")
              .execute()
        )

        data = res.data or []
        st.session_state.dados["corretores_cadastrados"] = [
            {
                "id": str(row.get("id") or ""),
                "nome": row.get("nome") or "",
                "cpf": row.get("cpf") or "",
                "banco": row.get("banco") or "",
                "agencia": row.get("agencia") or "",
                "conta": row.get("conta") or "",
                "pix": row.get("pix") or "",
            }
            for row in data
        ]

    except Exception as e:
        # Não derruba o app inteiro; mantém lista vazia e mostra erro
        st.session_state.dados["corretores_cadastrados"] = st.session_state.dados.get("corretores_cadastrados", [])
        st.error("Erro ao consultar corretores no Supabase. Abra 'Manage app' → Logs para ver detalhes.")


def ensure_corretores_carregados(forcar: bool = False):
    """
    Garante que os corretores foram carregados 1x por sessão e por usuário.
    Se forcar=True, recarrega mesmo que já tenha carregado.
    """
    ck = _cache_key_corretores()

    if forcar:
        st.session_state[ck] = False

    if st.session_state.get(ck, False):
        return

    _carregar_corretores_supabase()
    st.session_state[ck] = True


def listar_corretores_nomes():
    ensure_corretores_carregados()
    base = st.session_state.dados.get("corretores_cadastrados", [])
    return [c.get("nome", "") for c in base if (c.get("nome", "") or "").strip()]


def buscar_corretor_por_nome(nome: str):
    ensure_corretores_carregados()
    nome = (nome or "").strip()
    base = st.session_state.dados.get("corretores_cadastrados", [])
    for c in base:
        if (c.get("nome") or "").strip() == nome:
            return c
    return None


def salvar_corretor_supabase(nome, cpf, banco, agencia, conta, pix, corretor_id=None) -> str:
    """
    Insere/atualiza corretor no Supabase e retorna id.
    """
    sb = _supabase()
    if sb is None:
        return str(corretor_id or "")

    tenant = _tenant_imobiliaria()

    payload = {
        "imobiliaria": tenant,
        "nome": (nome or "").strip(),
        "cpf": (cpf or "").strip(),
        "banco": (banco or "").strip(),
        "agencia": (agencia or "").strip(),
        "conta": (conta or "").strip(),
        "pix": (pix or "").strip(),
    }

    if corretor_id:
        payload["id"] = corretor_id

    res = sb.table("corretores").upsert(payload).execute()

    if res.data and isinstance(res.data, list) and len(res.data) > 0:
        return str(res.data[0].get("id") or corretor_id or "")

    return str(corretor_id or "")

import json
from datetime import datetime, timezone

def _now_iso():
    return datetime.now(timezone.utc).isoformat()

def sb_get_max_versao(tenant: str, numero_contrato: str) -> int:
    sb = _supabase()
    if sb is None:
        return 0

    try:
        res = (
            sb.table("contratos")
              .select("versao")
              .eq("imobiliaria", tenant)
              .eq("numero_contrato", numero_contrato)
              .order("versao", desc=True)
              .limit(1)
              .execute()
        )
        data = res.data or []
        if not data:
            return 0
        return int(data[0].get("versao") or 0)
    except Exception:
        return 0

def sb_salvar_contrato_nova_versao():
    """
    Salva o contrato inteiro (st.session_state.dados) no Supabase em public.contratos,
    criando sempre uma NOVA versão (versao = max+1).
    """
    sb = _supabase()
    if sb is None:
        raise RuntimeError("Supabase não configurado (ver Secrets).")

    tenant = _tenant_imobiliaria()

    numero = (get("contrato__numero", "") or "").strip()
    if not numero:
        raise RuntimeError("Número do contrato está vazio. Preencha em 'Início'.")

    max_v = sb_get_max_versao(tenant, numero)
    nova_versao = max_v + 1
    label = f"versao_{nova_versao}"

    payload = {
        "imobiliaria": tenant,
        "numero_contrato": numero,
        "versao": nova_versao,
        "numero_versao_label": label,
        "dados": st.session_state.dados,  # jsonb
        "updated_at": _now_iso(),
    }

    # created_at só na criação (se seu banco já seta default, pode até remover)
    if nova_versao == 1:
        payload["created_at"] = _now_iso()

    res = sb.table("contratos").insert(payload).execute()
    return {"versao": nova_versao, "label": label, "data": (res.data or [])}


def excluir_corretor_supabase(corretor_id: str) -> bool:
    sb = _supabase()
    if sb is None or not corretor_id:
        return False

    tenant = _tenant_imobiliaria()
    sb.table("corretores").delete().eq("id", corretor_id).eq("imobiliaria", tenant).execute()
    return True


def adicionar_corretor_completo(nome, cpf, banco, agencia, conta, pix):
    """
    Cadastra corretor e garante que a lista recarregue para aparecer imediatamente.
    """
    nome = (nome or "").strip()
    if not nome:
        return ""

    # carrega base atual
    ensure_corretores_carregados()

    # evita duplicidade por nome (na mesma imobiliária)
    base = st.session_state.dados.get("corretores_cadastrados", [])
    for c in base:
        if (c.get("nome", "").strip() == nome):
            return c.get("id", "") or ""

    new_id = salvar_corretor_supabase(nome, cpf, banco, agencia, conta, pix, corretor_id=None)

    # ✅ FORÇA RECARGA para aparecer na lista imediatamente
    ensure_corretores_carregados(forcar=True)

    return new_id

# ============================================================
# HELPERS - DIGITOS / MÁSCARAS
# ============================================================

def so_digitos(s: str) -> str:
    return re.sub(r"\D", "", s or "")

def mask_cpf(v: str) -> str:
    d = so_digitos(v)[:11]
    if len(d) <= 3:
        return d
    if len(d) <= 6:
        return f"{d[:3]}.{d[3:]}"
    if len(d) <= 9:
        return f"{d[:3]}.{d[3:6]}.{d[6:]}"
    return f"{d[:3]}.{d[3:6]}.{d[6:9]}-{d[9:]}"

def cpf_callback_key(key: str):
    st.session_state[key] = mask_cpf(st.session_state.get(key, ""))
    set_(key, st.session_state[key])

def mask_cnpj(v: str) -> str:
    d = so_digitos(v)[:14]
    if len(d) <= 2:
        return d
    if len(d) <= 5:
        return f"{d[:2]}.{d[2:]}"
    if len(d) <= 8:
        return f"{d[:2]}.{d[2:5]}.{d[5:]}"
    if len(d) <= 12:
        return f"{d[:2]}.{d[2:5]}.{d[5:8]}/{d[8:]}"
    return f"{d[:2]}.{d[2:5]}.{d[5:8]}/{d[8:12]}-{d[12:]}"


def mask_cep(v: str) -> str:
    d = so_digitos(v)[:8]
    if len(d) <= 5:
        return d
    return f"{d[:5]}-{d[5:]}"


def parse_money_br(s: str) -> float:
    if not s:
        return 0.0
    t = s.strip().replace("R$", "").strip()
    t = t.replace(".", "").replace(" ", "")
    t = t.replace(",", ".")
    try:
        return float(t)
    except:
        return 0.0


def mask_money_br(s: str) -> str:
    if not s:
        return ""

    # remove qualquer coisa que não seja número, vírgula ou ponto
    t = s.strip().replace("R$", "").strip()

    # se digitou apenas letras ou vazio
    if not so_digitos(t):
        return ""

    v = parse_money_br(t)

    out = f"{v:,.2f}"
    out = out.replace(",", "X").replace(".", ",").replace("X", ".")

    return f"R$ {out}"


def money_br(v: float) -> str:
    out = f"{v:,.2f}"
    out = out.replace(",", "X").replace(".", ",").replace("X", ".")
    return f"R$ {out}"


def mask_ordinal_cartorio(s: str) -> str:
    d = so_digitos(s)
    if not d:
        return ""
    return f"{int(d)}º"

def abrir_cadastro_corretor(destino: str, prefix: str):
    """
    destino: 'venda' ou 'captacao'
    prefix: corv01, corc01 etc
    """

    # ✅ Ativa tela oculta
    set_("cadastro_corretor_ativado", True)

    # define para onde voltar
    set_("cadastro_corretor_destino", destino)
    set_("cadastro_corretor_prefix", prefix)

    # limpa campos
    set_("novo_corretor_nome", "")
    set_("novo_corretor_cpf", "")
    set_("novo_corretor_banco", "")
    set_("novo_corretor_agencia", "")
    set_("novo_corretor_conta", "")
    set_("novo_corretor_pix", "")

    go_to_step("cadastro_corretor")
    st.rerun()


def voltar_para_preco_chaves():
    # ✅ Desativa tela oculta
    set_("cadastro_corretor_ativado", False)

    # limpa dados de destino
    set_("cadastro_corretor_destino", "")
    set_("cadastro_corretor_prefix", "")

    go_to_step("preco_chaves")
    st.rerun()

from datetime import date

MESES_PT = [
    "janeiro", "fevereiro", "março", "abril", "maio", "junho",
    "julho", "agosto", "setembro", "outubro", "novembro", "dezembro"
]

def data_por_extenso(dt: date) -> str:
    """
    Retorna a data no formato: 04 de janeiro de 2026
    """
    return f"{dt.day:02d} de {MESES_PT[dt.month - 1]} de {dt.year}"

def linha_local_data() -> str:
    """
    Monta linha do tipo:
    Guarulhos/SP, 04 de janeiro de 2026.

    Pega cidade/UF do endereço do imóvel:
    - imovel__end__cidade
    - imovel__end__uf

    Se não existir cidade/UF, retorna só a data.
    """
    cidade = get("imovel__end__cidade", "").strip()
    uf = get("imovel__end__uf", "").strip()

    hoje = date.today()
    dt_txt = data_por_extenso(hoje)

    if cidade and uf:
        return f"{cidade}/{uf}, {dt_txt}."
    elif cidade:
        return f"{cidade}, {dt_txt}."
    else:
        return f"{dt_txt}."


def linha_direita(texto: str):
    st.markdown(
        f"<div style='text-align:right; font-size:15px; margin: 18px 0;'>{texto}</div>",
        unsafe_allow_html=True
    )

# ============================================================
# VIA CEP - BUSCA
# ============================================================
def buscar_endereco_por_cep(cep: str):
    cep_limpo = so_digitos(cep)
    if len(cep_limpo) != 8:
        return None
    try:
        r = requests.get(f"https://viacep.com.br/ws/{cep_limpo}/json/", timeout=6)
        r.raise_for_status()
        data = r.json()
        if data.get("erro"):
            return None
        return data
    except:
        return None


def format_endereco_completo(logradouro, numero, complemento, bairro, cidade, uf, cep):
    partes = []
    if logradouro:
        partes.append(logradouro)
    if numero:
        partes.append(f"n.º {numero}")
    if complemento:
        partes.append(complemento)
    if bairro:
        partes.append(bairro)
    if cidade and uf:
        partes.append(f"{cidade}/{uf}")
    elif cidade:
        partes.append(cidade)
    elif uf:
        partes.append(uf)

    texto = ", ".join([p for p in partes if p])
    if cep:
        texto += f" - CEP: {cep}"
    return texto.strip()


# ============================================================
# RECEITAWS - BUSCA CNPJ (TERCEIRO)
# ============================================================
def buscar_empresa_por_cnpj(cnpj: str):
    cnpj_limpo = so_digitos(cnpj)
    if len(cnpj_limpo) != 14:
        return None
    try:
        r = requests.get(f"https://receitaws.com.br/v1/cnpj/{cnpj_limpo}", timeout=12)
        r.raise_for_status()
        data = r.json()
        if data.get("status") == "ERROR":
            return None
        return data
    except:
        return None


# ============================================================
# WIZARD STEPS (dinâmico)
# ============================================================
WIZARD_STEPS_BASE = [
    {"id": "inicio", "title": "Início"},
    {"id": "imovel", "title": "Imóvel"},
    {"id": "vendedores", "title": "Parte Vendedora"},
    {"id": "compradores", "title": "Parte Compradora"},
    {"id": "preco_chaves", "title": "Preço e Chaves"},
    {"id": "parcelamento", "title": "Parcelamento (Detalhado)"},
    {"id": "permutas_dacao", "title": "Permutas / Dação (Detalhado)"},

    # ✅ Tela visível - será pré-visualização do contrato final
    {"id": "clausulas", "title": "Prévia de Contrato"},

    # ✅ TELAS OCULTAS
    {"id": "cadastro_corretor", "title": "Cadastro de Corretor", "hidden": True},
    {"id": "senha_admin", "title": "Senha Admin", "hidden": True},
    {"id": "admin_corretores", "title": "Admin Corretores", "hidden": True},

    # ✅ NOVA: ADMIN CLÁUSULAS (OCULTO)
    {"id": "admin_clausulas", "title": "Admin de Cláusulas", "hidden": True},
]

def steps():
    out = []
    for s in WIZARD_STEPS_BASE:

        # ✅ OCULTA A TELA CADASTRO CORRETOR NO MENU
        if s["id"] == "cadastro_corretor" and not get("cadastro_corretor_ativado", False):
            continue

        if s["id"] == "parcelamento" and not get("parcelamento_ativado", False):
            continue

        if s["id"] == "permutas_dacao" and not get("permutas_dacao_ativado", False):
            continue

        out.append(s)

    return out

def step():
    return steps()[st.session_state.step_index]

def go_next():
    if st.session_state.step_index < len(steps()) - 1:
        st.session_state.step_index += 1

def go_prev():
    if st.session_state.step_index > 0:
        st.session_state.step_index -= 1

def go_to_step(step_id: str):
    ids = [s["id"] for s in steps()]
    if step_id in ids:
        st.session_state.step_index = ids.index(step_id)

# ============================================================
# NAVEGAÇÃO PARA TELAS OCULTAS (ADMIN CORRETORES)
# ============================================================

def abrir_admin_corretores():
    st.session_state.step_index = steps().index(next(s for s in steps() if s["id"] == "admin_corretores"))
    st.rerun()

def abrir_admin_corretores_com_senha(step_voltar=None):
    # guarda de onde veio (para voltar depois)
    st.session_state.voltar_step_preco_chaves = step_voltar
    st.session_state.step_index = steps().index(next(s for s in steps() if s["id"] == "senha_admin"))
    st.rerun()

def abrir_admin_clausulas_com_senha(step_voltar=None):
    st.session_state.voltar_step_preco_chaves = step_voltar
    set_("destino_admin", "admin_clausulas")  # ✅ diz que o destino é admin_clausulas
    go_to_step("senha_admin")
    st.rerun()

def abrir_admin_clausulas():
    st.session_state.step_index = steps().index(next(s for s in steps() if s["id"] == "admin_clausulas"))
    st.rerun()

def voltar_da_admin_para_origem():
    # volta para a tela anterior (normalmente Preço e Chaves)
    if st.session_state.voltar_step_preco_chaves is not None:
        st.session_state.step_index = st.session_state.voltar_step_preco_chaves
    else:
        st.session_state.step_index = 0  # volta pro início se não tiver origem
    st.rerun()

# COMPONENTE: ENDEREÇO REUTILIZÁVEL (CEP automático)
# ============================================================
def endereco_callback(prefix: str):
    cep_key = f"{prefix}__cep"
    cep = mask_cep(st.session_state.get(cep_key, ""))
    st.session_state[cep_key] = cep
    set_(cep_key, cep)

    if len(so_digitos(cep)) == 8:
        data = buscar_endereco_por_cep(cep)
        if data:
            st.session_state[f"{prefix}__logradouro"] = data.get("logradouro", "")
            st.session_state[f"{prefix}__bairro"] = data.get("bairro", "")
            st.session_state[f"{prefix}__cidade"] = data.get("localidade", "")
            st.session_state[f"{prefix}__uf"] = data.get("uf", "")

            set_(f"{prefix}__logradouro", data.get("logradouro", ""))
            set_(f"{prefix}__bairro", data.get("bairro", ""))
            set_(f"{prefix}__cidade", data.get("localidade", ""))
            set_(f"{prefix}__uf", data.get("uf", ""))


def render_endereco(prefix: str, titulo: str):
    st.markdown(f"### 📍 {titulo}")

    # ============================
    # ✅ Inicialização correta
    # ============================
    keys = {
        "cep": f"{prefix}__cep",
        "logradouro": f"{prefix}__logradouro",
        "numero": f"{prefix}__numero",
        "complemento": f"{prefix}__complemento",
        "bairro": f"{prefix}__bairro",
        "cidade": f"{prefix}__cidade",
        "uf": f"{prefix}__uf",
        "texto": f"{prefix}__texto",
    }

    for k in keys.values():
        if k not in st.session_state:
            st.session_state[k] = get(k, "")

    # ============================
    # ✅ CEP com callback
    # ============================
    st.text_input(
        "CEP",
        key=keys["cep"],
        on_change=lambda: endereco_callback(prefix),
        placeholder="Ex.: 08663-040"
    )
    set_(keys["cep"], st.session_state[keys["cep"]])

    # ============================
    # ✅ Inputs SEM value= (Streamlit usa session_state)
    # ============================
    st.text_input("Logradouro", key=keys["logradouro"])
    st.text_input("Número", key=keys["numero"])
    st.text_input("Complemento", key=keys["complemento"])
    st.text_input("Bairro", key=keys["bairro"])
    st.text_input("Cidade", key=keys["cidade"])
    st.text_input("UF", key=keys["uf"])

    # ============================
    # ✅ Salvar em dados
    # ============================
    for campo in ["logradouro", "numero", "complemento", "bairro", "cidade", "uf"]:
        set_(keys[campo], st.session_state[keys[campo]])

    # ============================
    # ✅ Gerar endereço completo
    # ============================
    endereco = format_endereco_completo(
        st.session_state[keys["logradouro"]],
        st.session_state[keys["numero"]],
        st.session_state[keys["complemento"]],
        st.session_state[keys["bairro"]],
        st.session_state[keys["cidade"]],
        st.session_state[keys["uf"]],
        st.session_state[keys["cep"]],
    )

    # ✅ salva e mostra
    st.session_state[keys["texto"]] = endereco
    set_(keys["texto"], endereco)

    st.text_area(
        "Endereço completo (gerado)",
        value=endereco,
        height=90,
        disabled=True
    )


# ============================================================
# COMPONENTE: PF
# ============================================================
NACIONALIDADES = [
    "brasileiro", "brasileira",
    "portuguesa", "português",
    "italiana", "italiano",
    "espanhola", "espanhol",
    "argentina", "argentino",
    "americana", "americano",
    "alemã", "alemão",
    "francesa", "francês",
    "japonesa", "japonês",
    "chinesa", "chinês",
    "outra (escrever)"
]

def render_nacionalidade(prefix: str):
    nat_key = f"{prefix}__nacionalidade"
    if nat_key not in st.session_state:
        st.session_state[nat_key] = get(nat_key, "brasileiro")

    escolha = st.selectbox("Nacionalidade", NACIONALIDADES, key=nat_key, index=NACIONALIDADES.index(st.session_state[nat_key]) if st.session_state[nat_key] in NACIONALIDADES else 0)

    if escolha == "outra (escrever)":
        txt = st.text_input("Escreva a nacionalidade", value=get(f"{prefix}__nacionalidade_outra", ""), key=f"{prefix}__nacionalidade_outra")
        set_(nat_key, txt)
        set_(f"{prefix}__nacionalidade_outra", txt)
        return txt
    else:
        set_(nat_key, escolha)
        return escolha


def cpf_callback(prefix: str):
    k = f"{prefix}__cpf"
    st.session_state[k] = mask_cpf(st.session_state.get(k, ""))
    set_(k, st.session_state[k])


def render_pf(prefix: str, permitir_conjuge=True, titulo="PESSOA FÍSICA"):
    st.subheader(titulo)

    nome = st.text_input("Nome completo", value=get(f"{prefix}__nome", ""), key=f"{prefix}__nome")
    set_(f"{prefix}__nome", nome)

    nacionalidade = render_nacionalidade(prefix)
    set_(f"{prefix}__nacionalidade", nacionalidade)

    rg = st.text_input("RG nº", value=get(f"{prefix}__rg", ""), key=f"{prefix}__rg")
    set_(f"{prefix}__rg", rg)

    if f"{prefix}__cpf" not in st.session_state:
        st.session_state[f"{prefix}__cpf"] = get(f"{prefix}__cpf", "")

    st.text_input("CPF n.º", key=f"{prefix}__cpf", on_change=lambda: cpf_callback(prefix), placeholder="000.000.000-00")

    profissao = st.text_input("Profissão", value=get(f"{prefix}__profissao", ""), key=f"{prefix}__profissao")
    set_(f"{prefix}__profissao", profissao)

    estado_civil = st.selectbox(
        "Estado civil",
        ["solteiro(a)", "casado(a)", "união estável", "divorciado(a)", "viúvo(a)"],
        index=["solteiro(a)", "casado(a)", "união estável", "divorciado(a)", "viúvo(a)"].index(get(f"{prefix}__estado_civil", "solteiro(a)")),
        key=f"{prefix}__estado_civil"
    )
    set_(f"{prefix}__estado_civil", estado_civil)

    # ✅ Regime de bens (somente se casado(a) ou união estável)
    regime_key = f"{prefix}__regime_bens"
    if regime_key not in st.session_state:
        st.session_state[regime_key] = get(regime_key, "")

    if estado_civil in ("casado(a)", "união estável"):
        regime = st.selectbox(
            "Regime de bens",
            [
                "comunhão parcial de bens",
                "comunhão universal de bens",
                "separação total de bens",
                "participação final nos aquestos",
                "outro (escrever)"
            ],
            key=regime_key
        )

        if regime == "outro (escrever)":
            outro = st.text_input(
                "Escreva o regime de bens",
                value=get(f"{prefix}__regime_bens_outro", ""),
                key=f"{prefix}__regime_bens_outro"
            )
            set_(regime_key, outro)
            set_(f"{prefix}__regime_bens_outro", outro)
        else:
            set_(regime_key, regime)
    else:
        set_(regime_key, "")
        set_(f"{prefix}__regime_bens_outro", "")

    render_endereco(f"{prefix}__end", "Endereço")

    if permitir_conjuge and estado_civil in ("casado(a)", "união estável"):
        st.markdown("### 👥 Cônjuge / Companheiro(a)")

        rotulo = "Nome do cônjuge" if estado_civil == "casado(a)" else "Nome do companheiro(a)"

        nome_c = st.text_input(rotulo, value=get(f"{prefix}__conj_nome", ""), key=f"{prefix}__conj_nome")
        set_(f"{prefix}__conj_nome", nome_c)

        # ✅ Nacionalidade do cônjuge/companheiro(a)
        st.markdown("**Nacionalidade**")
        nat_conj_key = f"{prefix}__conj_nacionalidade"
        if nat_conj_key not in st.session_state:
            st.session_state[nat_conj_key] = get(nat_conj_key, "brasileiro")

        nat_conj = st.selectbox(" ", NACIONALIDADES, key=nat_conj_key)
        if nat_conj == "outra (escrever)":
            txt = st.text_input(
                "Escreva a nacionalidade do cônjuge/companheiro(a)",
                value=get(f"{prefix}__conj_nacionalidade_outra", ""),
                key=f"{prefix}__conj_nacionalidade_outra"
            )
            set_(nat_conj_key, txt)
            set_(f"{prefix}__conj_nacionalidade_outra", txt)
        else:
            set_(nat_conj_key, nat_conj)

        # ✅ Profissão do cônjuge/companheiro(a)
        prof_c = st.text_input(
            "Profissão do cônjuge/companheiro(a)",
            value=get(f"{prefix}__conj_profissao", ""),
            key=f"{prefix}__conj_profissao"
        )
        set_(f"{prefix}__conj_profissao", prof_c)

        # ✅ RG do cônjuge/companheiro(a)
        rg_c = st.text_input(
            "RG do cônjuge/companheiro(a)",
            value=get(f"{prefix}__conj_rg", ""),
            key=f"{prefix}__conj_rg"
        )
        set_(f"{prefix}__conj_rg", rg_c)

        # ✅ CPF do cônjuge/companheiro(a)
        if f"{prefix}__conj_cpf" not in st.session_state:
            st.session_state[f"{prefix}__conj_cpf"] = get(f"{prefix}__conj_cpf", "")

        st.text_input(
            "CPF n.º do cônjuge/companheiro(a)",
            key=f"{prefix}__conj_cpf",
            on_change=lambda: cpf_callback_key(f"{prefix}__conj_cpf"),
            placeholder="000.000.000-00"
        )
        set_(f"{prefix}__conj_cpf", st.session_state.get(f"{prefix}__conj_cpf", ""))

    # ============================================================
    # ✅ VALIDAÇÃO OBRIGATÓRIA DO CÔNJUGE / COMPANHEIRO(A)
    # ============================================================
    obrigatorio_conjuge = estado_civil in ("casado(a)", "união estável")

    if obrigatorio_conjuge:
        if not get(f"{prefix}__conj_nome", "").strip():
            st.error("⚠️ Para estado civil CASADO(A) ou UNIÃO ESTÁVEL, o preenchimento do cônjuge/companheiro(a) é obrigatório.")
            set_(f"{prefix}__bloqueio_avancar", True)
        else:
            set_(f"{prefix}__bloqueio_avancar", False)
    else:
        set_(f"{prefix}__bloqueio_avancar", False)


# ============================================================
# COMPONENTE: PJ (CNPJ primeiro + busca Receita)
# ============================================================
def cnpj_callback(prefix: str):
    k = f"{prefix}__cnpj"
    st.session_state[k] = mask_cnpj(st.session_state.get(k, ""))
    set_(k, st.session_state[k])

    dados = buscar_empresa_por_cnpj(st.session_state[k])
    if not dados:
        return

    razao = dados.get("nome", "")
    set_(f"{prefix}__razao_social", razao)
    st.session_state[f"{prefix}__razao_social"] = razao

    # endereço da receita
    cep = mask_cep(dados.get("cep", ""))
    set_(f"{prefix}__end__cep", cep)
    st.session_state[f"{prefix}__end__cep"] = cep

    # dispara busca do cep para preencher logradouro/bairro/cidade/uf
    endereco_callback(f"{prefix}__end")

    # número e complemento
    numero = dados.get("numero", "")
    comp = dados.get("complemento", "")
    set_(f"{prefix}__end__numero", numero)
    set_(f"{prefix}__end__complemento", comp)
    st.session_state[f"{prefix}__end__numero"] = numero
    st.session_state[f"{prefix}__end__complemento"] = comp


def render_pj(prefix: str, titulo="PESSOA JURÍDICA"):
    st.subheader(titulo)

    if f"{prefix}__cnpj" not in st.session_state:
        st.session_state[f"{prefix}__cnpj"] = get(f"{prefix}__cnpj", "")

    st.text_input("CNPJ nº (preencher primeiro)", key=f"{prefix}__cnpj", on_change=lambda: cnpj_callback(prefix), placeholder="00.000.000/0000-00")

    razao = st.text_input("Razão social (vinda da Receita)", value=get(f"{prefix}__razao_social", ""), key=f"{prefix}__razao_social", disabled=True)
    set_(f"{prefix}__razao_social", razao)

    render_endereco(f"{prefix}__end", "Endereço da empresa")

    st.divider()
    st.markdown("### 👤 Representante legal (quem assina)")

    # Representante: só Nome + CPF
    rep_nome = st.text_input("Nome do representante", value=get(f"{prefix}__rep_nome", ""), key=f"{prefix}__rep_nome")
    set_(f"{prefix}__rep_nome", rep_nome)

    if f"{prefix}__rep_cpf" not in st.session_state:
        st.session_state[f"{prefix}__rep_cpf"] = get(f"{prefix}__rep_cpf", "")

    st.text_input(
        "CPF do representante",
        key=f"{prefix}__rep_cpf",
        on_change=lambda: cpf_callback_key(f"{prefix}__rep_cpf"),
        placeholder="000.000.000-00"
)

# ============================================================
# FORMULÁRIO DE PARTE (PF/PJ)
# ============================================================
def render_parte(prefix: str, titulo: str):
    st.header(titulo)

    tipo_key = f"{prefix}__tipo"
    if tipo_key not in st.session_state:
        st.session_state[tipo_key] = get(tipo_key, "Pessoa Física")

    tipo = st.radio("Esta parte é:", ["Pessoa Física", "Pessoa Jurídica"], horizontal=True, key=tipo_key)
    set_(tipo_key, tipo)

    st.divider()

    if tipo == "Pessoa Física":
        render_pf(prefix, permitir_conjuge=True)
    else:
        render_pj(prefix)


# ============================================================
# DINÂMICOS: vendedores / compradores
# ============================================================
def ensure_min_one_party(list_key: str, base_prefix: str):
    lst = get_list(list_key)
    if len(lst) == 0:
        lst.append(f"{base_prefix}01")
        set_list(list_key, lst)


def add_party(list_key: str, base_prefix: str):
    lst = get_list(list_key)
    nxt = len(lst) + 1
    lst.append(f"{base_prefix}{nxt:02d}")
    set_list(list_key, lst)


def remove_last_party(list_key: str):
    lst = get_list(list_key)
    if len(lst) > 1:
        lst.pop()
        set_list(list_key, lst)


# ============================================================
# CORRETORES / CAPTADORES
# ============================================================

def ensure_agents():
    # garante pelo menos 1 corretor em cada lista
    if "corretores_venda" not in st.session_state.dados:
        set_list("corretores_venda", ["corv01"])
    if "corretores_captacao" not in st.session_state.dados:
        set_list("corretores_captacao", ["corc01"])


def mask_percent(s: str) -> str:
    d = so_digitos(s)
    if not d:
        return ""
    return f"{int(d)}%"


def percent_callback_key(key: str):
    st.session_state[key] = mask_percent(st.session_state.get(key, ""))
    set_(key, st.session_state[key])


def render_agente(prefix: str, titulo: str, pct_default: str):

    nomes = listar_corretores_nomes()
    opcoes = ["(selecionar)"] + nomes

    escolha = st.selectbox(
        titulo,
        opcoes,
        key=f"{prefix}__select",
        index=0
    )

    if escolha != "(selecionar)":
        set_(f"{prefix}__nome", escolha)
        st.session_state[f"{prefix}__nome"] = escolha

        # salva dados completos em session_state (se precisar)
        corretor = buscar_corretor_por_nome(escolha)
        if corretor:
            set_(f"{prefix}__cpf", corretor.get("cpf", ""))
            set_(f"{prefix}__banco", corretor.get("banco", ""))
            set_(f"{prefix}__agencia", corretor.get("agencia", ""))
            set_(f"{prefix}__conta", corretor.get("conta", ""))
            set_(f"{prefix}__pix", corretor.get("pix", ""))

    # ✅ botão abre tela oculta de cadastro
    if st.button("➕ Cadastrar novo corretor", key=f"{prefix}__novo"):
        destino = "venda" if prefix.startswith("corv") else "captacao"
        abrir_cadastro_corretor(destino, prefix)

    # ✅ % com máscara automática
    if f"{prefix}__pct" not in st.session_state:
        st.session_state[f"{prefix}__pct"] = get(f"{prefix}__pct", pct_default)

    st.text_input(
        "% da comissão",
        key=f"{prefix}__pct",
        on_change=lambda: percent_callback_key(f"{prefix}__pct"),
        placeholder=pct_default
    )
    set_(f"{prefix}__pct", st.session_state.get(f"{prefix}__pct", ""))



# ============================================================
# RESUMO
# ============================================================
def resumo_endereco(prefix: str):
    return get(f"{prefix}__texto", "")


def resumo_parte(prefix: str):
    tipo = get(f"{prefix}__tipo", "Pessoa Física")
    out = []

    if tipo == "Pessoa Física":
        out.append(f"{get(f'{prefix}__nome','')}")
        out.append(f"Nacionalidade: {get(f'{prefix}__nacionalidade','')}")
        out.append(f"CPF: {get(f'{prefix}__cpf','')}")
        if get(f"{prefix}__rg"):
            out.append(f"RG: {get(f'{prefix}__rg')}")
        if get(f"{prefix}__profissao"):
            out.append(f"Profissão: {get(f'{prefix}__profissao')}")
        out.append(f"Estado civil: {get(f'{prefix}__estado_civil','')}")
        out.append(f"Endereço: {resumo_endereco(f'{prefix}__end')}")

        if get(f"{prefix}__estado_civil") in ("casado(a)", "união estável"):
            out.append(f"Cônjuge: {get(f'{prefix}__conj_nome','')} CPF: {get(f'{prefix}__conj_cpf','')}")

    else:
        out.append(f"{get(f'{prefix}__razao_social','')}")
        out.append(f"CNPJ: {get(f'{prefix}__cnpj','')}")
        out.append(f"Endereço: {resumo_endereco(f'{prefix}__end')}")
        out.append(f"Representante: {get(f'{prefix}__rep_nome','')} CPF: {get(f'{prefix}__rep_cpf','')}")

    return "\n".join([x for x in out if x.strip()])


def resumo_completo():
    linhas = []

    linhas.append("=== CONTRATO ===")
    linhas.append(f"Nº: {get('contrato__numero','')}")
    linhas.append(f"Tipo: {get('contrato__tipo','')}")
    linhas.append(f"E-mail solicitante: {get('contrato__email_solicitante','')}")
    linhas.append("")

    linhas.append("=== IMÓVEL ===")
    linhas.append(f"Tipo: {get('imovel__tipo','')}")
    linhas.append(f"Matrícula: {get('imovel__matricula','')}")
    linhas.append(f"Cartório: {get('imovel__cartorio','')}")
    linhas.append(f"Cidade do cartório: {get('imovel__cidade_cartorio','')}")
    linhas.append(f"Contribuinte: {get('imovel__contribuinte','')}")
    linhas.append(f"Endereço: {get('imovel__end__texto','')}")
    if get("imovel__descricao_matricula"):
        linhas.append("Descrição: " + get("imovel__descricao_matricula"))
    linhas.append("")

    linhas.append("=== VENDEDORES ===")
    for i, pfx in enumerate(get_list("vendedores"), start=1):
        linhas.append(f"\n--- VENDEDOR {i} ---")
        linhas.append(resumo_parte(pfx))

    linhas.append("\n=== COMPRADORES ===")
    for i, pfx in enumerate(get_list("compradores"), start=1):
        linhas.append(f"\n--- COMPRADOR {i} ---")
        linhas.append(resumo_parte(pfx))

    linhas.append("\n=== PREÇO / CHAVES / COMISSÃO ===")
    linhas.append(f"Preço total: {get('preco_total','')}")
    linhas.append(f"Entrega de chaves: {get('entrega_chaves','')}")
    linhas.append(f"Quem paga comissão: {get('quem_paga_comissao','')}")
    linhas.append(f"Valor comissão: {get('valor_comissao','')}")
    linhas.append(f"Momento pgto: {get('momento_pagto','')}")
    linhas.append("")

    linhas.append("Corretores de venda:")
    for pfx in get_list("corretores_venda"):
        linhas.append(f"- {get(pfx+'__nome','')} ({get(pfx+'__pct','')}%)")

    linhas.append("\nCorretores de captação:")
    for pfx in get_list("corretores_captacao"):
        linhas.append(f"- {get(pfx+'__nome','')} ({get(pfx+'__pct','')}%)")

    if get("parcelamento_ativado", False):
        linhas.append("\n=== PARCELAMENTO ===")
        linhas.append(get("parcelamento_descricao", ""))

    if get("permutas_dacao_ativado", False):
        linhas.append("\n=== PERMUTAS / DAÇÃO ===")
        linhas.append(get("dacao_descricao", ""))
        if get("dacao_imovel", "NÃO") == "SIM":
            linhas.append(f"Endereço do imóvel da dação: {get('dacao_imovel__end__texto','')}")

    return "\n".join(linhas)

# ============================================================
# CLÁUSULAS: ENTREGA DE CHAVES (GERADOR + EDITOR)
# ============================================================

def clausulas_padrao_entrega_chaves() -> dict:
    """
    Retorna o dicionário PADRÃO (original) de textos para cada opção de entrega de chaves.
    """
    return {
        "30 dias após crédito em conta": (
            "Em até 30 (trinta) dias corridos após o valor total do IMÓVEL seja disponibilizado "
            "ou creditado na conta corrente da PARTE VENDEDORA ou na conta de quem esta indicar expressamente."
        ),
        "30 dias após assinatura no Banco": (
            "Em até 30 (trinta) dias corridos após assinatura da escritura definitiva perante "
            "instituição financeira competente."
        ),
        "30 dias após assinatura do CCV": (
            "Em até 30 (trinta) dias corridos após assinatura da PARTE COMPRADORA do presente instrumento."
        ),
        "No ato da assinatura no Banco": (
            "No ato da assinatura da escritura definitiva perante instituição financeira competente."
        ),
        "No ato da assinatura do CCV": (
            "No ato da assinatura da PARTE COMPRADORA do presente instrumento."
        ),
        "24 horas do crédito em conta": (
            "Em até 24 (vinte e quatro) horas após o valor total do IMÓVEL seja disponibilizado "
            "ou creditado na conta corrente da PARTE VENDEDORA ou na conta de quem esta indicar expressamente."
        ),
        "Escrever no contrato": (
            "⚠️ Texto a ser redigido manualmente no contrato final (campo específico)."
        ),
    }


def ensure_clausulas_entrega_chaves():
    """
    Garante que o dicionário de cláusulas de entrega de chaves exista no st.session_state.dados.
    """
    if "clausulas_entrega_chaves" not in st.session_state.dados:
        set_("clausulas_entrega_chaves", clausulas_padrao_entrega_chaves())


def obter_clausula_entrega_chaves() -> str:
    """
    Retorna o texto final da cláusula de entrega de chaves com base na escolha do usuário.
    """
    ensure_clausulas_entrega_chaves()

    escolha = get("entrega_chaves", "").strip()
    if not escolha:
        return ""

    if escolha == "Escrever no contrato":
        return get("entrega_chaves_texto", "").strip()

    mapa = get("clausulas_entrega_chaves", {})
    return mapa.get(escolha, "")
    
# ============================================================
# TAGS PARA CONTRATO (injeção em Word/HTML/Texto)
# ============================================================

def tag_dias_entrega_chaves() -> str:
    """
    Retorna o texto que substitui a tag <DIAS_ENTREGA_DE_CHAVES> no contrato.

    - Se a entrega for "Escrever no contrato", retorna o texto digitado no campo.
    - Caso contrário, retorna o texto padrão (ou editado no admin) conforme o selectbox.
    """
    return obter_clausula_entrega_chaves().strip()

# ============================================================
# HELPERS DE FORMATAÇÃO (centralizado / justificado)
# ============================================================

def texto_centralizado(texto: str, tamanho_px: int = 18, negrito: bool = True):
    """
    Exibe texto centralizado no Streamlit, em caixa alta.
    Use para títulos principais do contrato.
    """
    if not texto:
        return
    fw = "700" if negrito else "400"
    st.markdown(
        f"<div style='text-align:center; font-size:{tamanho_px}px; font-weight:{fw}; text-transform:uppercase;'>{texto}</div>",
        unsafe_allow_html=True
    )

def texto_justificado(texto: str, tamanho_px: int = 15):
    """
    Exibe texto justificado (alinhamento total).
    Use para cláusulas e textos corridos.
    """
    if not texto:
        return
    st.markdown(
        f"<div style='text-align:justify; font-size:{tamanho_px}px; line-height:1.6;'>{texto}</div>",
        unsafe_allow_html=True
    )

# ============================================================
# HELPERS DE FORMATAÇÃO (BOX COM BORDA)
# ============================================================

def box_texto_justificado(texto: str, tamanho_px: int = 15):
    """
    Exibe um bloco com borda externa, fundo leve e texto justificado.
    Ideal para QUALIFICAÇÕES DAS PARTES no contrato.
    """
    if not texto:
        return

    html = f"""
    <div style="
        border: 1px solid rgba(120,120,120,0.6);
        padding: 14px 16px;
        border-radius: 6px;
        background: rgba(255,255,255,0.02);
        text-align: justify;
        font-size: {tamanho_px}px;
        line-height: 1.65;
        ">
        {texto}
    </div>
    """
    st.markdown(html, unsafe_allow_html=True)

# ============================================================
# REGRAS DO CONTRATO (derivações por tipo)
# ============================================================

def papel_parte_vendedora_ou_cedente() -> str:
    """
    Decide automaticamente qual termo usar:
    - "PARTE VENDEDORA" se for Compromisso de Compra e Venda
    - "PARTE CEDENTE" se for Cessão de Posse e Direitos
    """
    tipo = get("contrato__tipo", "").strip().lower()

    # ✅ ajuste seguro para variações de escrita
    if "cessão" in tipo or "posse" in tipo:
        return "PARTE CEDENTE"

    # padrão: compromisso compra e venda
    return "PARTE VENDEDORA"

def tipo_juridico_contrato() -> str:
    """
    Define automaticamente o título jurídico do contrato.
    - Compra e venda com financiamento -> "Compromisso de Venda e Compra de Imóvel com Financiamento"
    - Compra e venda sem financiamento -> "Compromisso de Compra e Venda de Imóvel"
    - Cessão de posse -> mantém o texto original (não existe financiamento)
    """

    tipo_raw = get("contrato__tipo", "").strip()
    tipo_lower = tipo_raw.lower()

    financiamento = get("preco_financiamento", "").strip()

    # Cessão não muda e não tem financiamento
    if "cessão" in tipo_lower or "posse" in tipo_lower:
        return tipo_raw

    # Compra e venda
    if financiamento:
        return "Compromisso de Venda e Compra de Imóvel com Financiamento"
    return "Compromisso de Compra e Venda de Imóvel"


def frase_adiante_designado() -> str:
    """
    Monta a frase variável conforme o tipo do contrato.
    Exemplo:
    'Adiante simplesmente designado como PARTE VENDEDORA'
    """
    papel = papel_parte_vendedora_ou_cedente()
    return f"Adiante simplesmente designado como {papel}:"

def papel_parte_compradora_ou_cessionaria() -> str:
    """
    Decide automaticamente qual termo usar:
    - "PARTE COMPRADORA" se for Compromisso de Venda e Compra
    - "PARTE CESSIONÁRIA" se for Cessão de Posse e Direitos
    """
    tipo = get("contrato__tipo", "").strip().lower()

    if "cessão" in tipo or "posse" in tipo:
        return "PARTE CESSIONÁRIA"

    return "PARTE COMPRADORA"


def frase_adiante_designado_compradora() -> str:
    """
    Monta a frase variável conforme o tipo do contrato.
    Exemplo:
    'Adiante simplesmente designado como PARTE COMPRADORA:'
    """
    papel = papel_parte_compradora_ou_cessionaria()
    return f"Adiante simplesmente designado como {papel}:"

# ============================================================
# QUALIFICAÇÃO DAS PARTES (VENDEDOR / CEDENTE)
# ============================================================

def eh_feminino_pela_nacionalidade(nacionalidade: str) -> bool:
    """
    Determina o gênero presumido pelo termo da nacionalidade:
    - 'brasileira' -> feminino
    - 'brasileiro' -> masculino
    Se não for possível inferir, retorna False (masculino por padrão).
    """
    nat = (nacionalidade or "").strip().lower()
    return nat.endswith("a")  # brasileira, portuguesa, italiana...


def ajustar_estado_civil_genero(estado_civil: str, nacionalidade: str) -> str:
    """
    Ajusta automaticamente solteiro/divorciado/viúvo conforme gênero inferido da nacionalidade.
    Somente aplica quando:
    - estado_civil estiver no formato com (a)
    - e a nacionalidade for claramente masculina/feminina (termina com 'o' ou 'a')
    """
    ec = (estado_civil or "").strip().lower()
    nat = (nacionalidade or "").strip().lower()

    # Se não tiver o padrão (a), não mexe
    if "(a)" not in ec:
        return ec

    feminino = eh_feminino_pela_nacionalidade(nat)

    mapa = {
        "solteiro(a)": ("solteiro", "solteira"),
        "divorciado(a)": ("divorciado", "divorciada"),
        "viúvo(a)": ("viúvo", "viúva"),
        "casado(a)": ("casado", "casada"),
    }

    if ec in mapa:
        masc, fem = mapa[ec]
        return fem if feminino else masc

    return ec

def qualificar_pf(prefix: str) -> str:
    """
    Qualificação PF no padrão jurídico solicitado.

    Regras:
    - Se SOLTEIRO(A), DIVORCIADO(A) ou VIÚVO(A):
        -> estado civil aparece na qualificação individual
        -> ordem: NOME, NACIONALIDADE, ESTADO CIVIL, PROFISSÃO, RG, CPF, ENDEREÇO.
    - Se CASADO(A) ou UNIÃO ESTÁVEL:
        -> qualificação conjunta quando houver cônjuge/companheiro(a)
        -> inclui "ambos casados entre si" / "conviventes em união estável entre si"
        -> regime de bens aparece apenas uma vez
        -> endereço aparece apenas uma vez no final
    - Endereço sempre aparece ao final.
    - ✅ Corrige automaticamente solteiro/divorciado/viúvo/casado conforme gênero inferido da nacionalidade.
    """

    # ============================
    # Dados da pessoa principal
    # ============================
    nome = get(f"{prefix}__nome", "").strip().upper()
    nacionalidade = get(f"{prefix}__nacionalidade", "").strip()
    profissao = get(f"{prefix}__profissao", "").strip()
    rg = get(f"{prefix}__rg", "").strip()
    cpf = get(f"{prefix}__cpf", "").strip()
    estado_civil_raw = get(f"{prefix}__estado_civil", "").strip()
    regime_bens = get(f"{prefix}__regime_bens", "").strip()
    endereco = get(f"{prefix}__end__texto", "").strip()

    # ✅ estado civil ajustado por gênero (solteira/divorciada/viúva etc.)
    estado_civil_ajustado = ajustar_estado_civil_genero(estado_civil_raw, nacionalidade)

    # ============================
    # Dados do cônjuge/companheiro(a)
    # ============================
    conj_nome = get(f"{prefix}__conj_nome", "").strip().upper()
    conj_nacionalidade = get(f"{prefix}__conj_nacionalidade", "").strip()
    conj_profissao = get(f"{prefix}__conj_profissao", "").strip()
    conj_rg = get(f"{prefix}__conj_rg", "").strip()
    conj_cpf = get(f"{prefix}__conj_cpf", "").strip()

    # ============================
    # Função auxiliar de qualificação individual
    # ============================
    def qual_individual(nome, nacionalidade, estado_civil, profissao, rg, cpf):
        if not nome:
            return ""

        detalhes = []

        if nacionalidade:
            detalhes.append(nacionalidade)

        # ✅ estado civil vem logo após nacionalidade quando for informado
        if estado_civil:
            detalhes.append(estado_civil)

        if profissao:
            detalhes.append(profissao)

        if rg:
            detalhes.append(f"RG n.º {rg}")

        if cpf:
            detalhes.append(f"CPF n.º {cpf}")

        return f"{nome}, " + ", ".join(detalhes) if detalhes else nome

    # ============================
    # 1) SEM cônjuge/companheiro(a)
    # ============================
    if not conj_nome:
        # ✅ inclui estado civil no corpo individual
        texto = qual_individual(
            nome, nacionalidade, estado_civil_ajustado, profissao, rg, cpf
        )

        # ✅ regime de bens apenas se CASADO(A) ou UNIÃO ESTÁVEL
        if estado_civil_raw in ("casado(a)", "união estável") and regime_bens:
            texto += f", sob o regime de {regime_bens}"

        # ✅ endereço sempre no final
        if endereco:
            texto += f", com residência e domicílio em {endereco}."
        else:
            texto += "."

        return texto

    # ============================
    # 2) COM cônjuge/companheiro(a)
    # ============================
    # ✅ quando há cônjuge, não repete estado civil individualmente
    p1 = qual_individual(nome, nacionalidade, "", profissao, rg, cpf)
    p2 = qual_individual(conj_nome, conj_nacionalidade, "", conj_profissao, conj_rg, conj_cpf)

    # ✅ frase padrão do casal conforme estado civil
    if estado_civil_raw == "união estável":
        uniao_txt = "conviventes em união estável entre si"
    else:
        uniao_txt = "ambos casados entre si"

    # ✅ regime de bens aparece apenas uma vez para o casal
    regime_txt = f", sob o regime de {regime_bens}" if regime_bens else ""

    # ✅ endereço aparece apenas uma vez para o casal
    if endereco:
        return f"{p1}, e {p2}, {uniao_txt}{regime_txt} e com residência e domicílio em {endereco}."

    return f"{p1}, e {p2}, {uniao_txt}{regime_txt}."



def qualificar_pj(prefix: str) -> str:
    """
    Monta a qualificação completa de uma Pessoa Jurídica, para uso no contrato.
    """
    razao = get(f"{prefix}__razao_social", "").strip().upper()
    cnpj = get(f"{prefix}__cnpj", "").strip()
    endereco = get(f"{prefix}__end__texto", "").strip()

    rep_nome = get(f"{prefix}__rep_nome", "").strip().upper()
    rep_cpf = get(f"{prefix}__rep_cpf", "").strip()

    partes = []
    if razao:
        partes.append(razao)

    detalhes = []
    if cnpj:
        detalhes.append(f"CNPJ n.º {cnpj}")
    if endereco:
        detalhes.append(f"com sede em {endereco}")

    if rep_nome:
        rep = f"neste ato representada por {rep_nome}"
        if rep_cpf:
            rep += f", CPF n.º {rep_cpf}"
        rep += ", na forma de dua situação cadastral de pessoa jurídica da Receita Federal ou contrato social"
        detalhes.append(rep)

    if detalhes:
        partes.append(", " + ", ".join(detalhes) + ".")

    return "".join(partes).strip()


def qualificar_parte(prefix: str) -> str:
    """
    Decide automaticamente se a parte é PF ou PJ e chama a função correta.
    """
    tipo = get(f"{prefix}__tipo", "Pessoa Física").strip()

    if tipo == "Pessoa Jurídica":
        return qualificar_pj(prefix)

    return qualificar_pf(prefix)


def bloco_qualificacao_vendedores() -> str:
    """
    Gera o texto completo da qualificação da PARTE VENDEDORA / CEDENTE,
    considerando 1 ou mais pessoas na lista "vendedores".

    Retorna HTML formatado com <br><br> para separar pessoas.
    """
    vendedores = get_list("vendedores")
    if not vendedores:
        return ""

    textos = []
    for pfx in vendedores:
        t = qualificar_parte(pfx)
        if t:
            textos.append(t)

    # separa cada pessoa com uma linha em branco (como no seu modelo)
    return "<br><br>".join(textos)

def papel_parte_compradora_ou_cessionaria() -> str:
    """
    Decide automaticamente qual termo usar:
    - "PARTE COMPRADORA" se for Compromisso de Venda e Compra
    - "PARTE CESSIONÁRIA" se for Cessão de Posse e Direitos
    """
    tipo = get("contrato__tipo", "").strip().lower()

    if "cessão" in tipo or "posse" in tipo:
        return "PARTE CESSIONÁRIA"

    return "PARTE COMPRADORA"


def frase_adiante_designado_comprador() -> str:
    """
    Monta a frase variável conforme o tipo do contrato para comprador/cessionária.
    """
    papel = papel_parte_compradora_ou_cessionaria()
    return f"Adiante simplesmente designado como {papel}:"


def bloco_qualificacao_compradores() -> str:
    """
    Gera o texto completo da qualificação da PARTE COMPRADORA / CESSIONÁRIA,
    considerando 1 ou mais pessoas na lista "compradores".
    Retorna HTML com <br><br> para separar pessoas.
    """
    compradores = get_list("compradores")
    if not compradores:
        return ""

    textos = []
    for pfx in compradores:
        t = qualificar_parte(pfx)
        if t:
            textos.append(t)

    return "<br><br>".join(textos)

def bloco_qualificacao_compradores() -> str:
    """
    Gera o texto completo da qualificação da PARTE COMPRADORA / CESSIONÁRIA,
    considerando 1 ou mais pessoas na lista "compradores".

    Retorna HTML formatado com <br><br> para separar pessoas.
    """
    compradores = get_list("compradores")
    if not compradores:
        return ""

    textos = []
    for pfx in compradores:
        t = qualificar_parte(pfx)
        if t:
            textos.append(t)

    return "<br><br>".join(textos)

def bloco_intermediadora() -> str:
    """
    Retorna o texto FIXO da INTERMEDIADORA para o contrato.
    Mais adiante, poderá virar dinâmico (lista de imobiliárias).
    """
    return (
        "IMOBILIÁRIA MONTE SIÃO LTDA, pessoa jurídica de direito privado, "
        "CNPJ n.º 30.177.724/0001-76, CRECI n.º 33.150-J, com sede na Rua Roberto, n.º 14, "
        "Jardim Santa Mena, Guarulhos/SP - CEP: 07096-070, representada por "
        "JOSIVAN MOURA DA SILVA, brasileiro, corretor de imóveis, RG n.º 55.786.890-7 SSP, "
        "CPF n.º 343.173.968-74."
    )

def pagamento_juridico() -> str:
    """
    Monta automaticamente o texto jurídico (itens a-i) da forma de pagamento,
    com base nos valores preenchidos no wizard (preco_sinal, preco_entrada, etc.).
    """

    sinal = get("preco_sinal", "").strip()
    entrada = get("preco_entrada", "").strip()
    financiamento = get("preco_financiamento", "").strip()
    fgts = get("preco_fgts", "").strip()
    subsidio = get("preco_subsidio", "").strip()
    recurso_proprio = get("preco_recurso_proprio", "").strip()
    carta_credito = get("preco_carta_credito", "").strip()
    parcelamento_total = get("preco_parcelamento_total", "").strip()
    outros = get("preco_outros", "").strip()
    outros_desc = get("preco_outros_descricao", "").strip()

    # ✅ Se houver financiamento, o texto muda (instituição financeira)
    ha_financiamento = bool(financiamento)

    # Tag variável: se tem financiamento, "instituição financeira competente", senão "tabelião de notas competente"
    destino_escritura = "instituição financeira competente" if ha_financiamento else "tabelião de notas competente"

    itens = []

    # a) SINAL
    if sinal:
        itens.append(
            f"a) {sinal}, em moeda corrente nacional, como sinal e princípio de pagamento, "
            f"que, com ciência e anuência da PARTE VENDEDORA, serão pagos diretamente à INTERMEDIADORA "
            f"na assinatura deste instrumento em sua conta bancária ou a conta de quem indicar;"
        )

    # b) ENTRADA
    if entrada:
        itens.append(
            f"b) {entrada}, em moeda corrente nacional, a serem pagos à PARTE VENDEDORA em sua conta bancária "
            f"ou na conta de quem indicar no dia da assinatura da escritura perante {destino_escritura};"
        )

    # c) FINANCIAMENTO
    if financiamento:
        itens.append(
            f"c) {financiamento}, através de financiamento bancário, a serem pagos à PARTE VENDEDORA;"
        )

    # d) FGTS
    if fgts:
        itens.append(
            f"d) {fgts}, através de valores vinculados à conta do Fundo de Garantia do Tempo de Serviço - FGTS, "
            f"a serem pagos à PARTE VENDEDORA;"
        )

    # e) SUBSÍDIO
    if subsidio:
        itens.append(
            f"e) {subsidio}, mediante subsídio governamental a serem pagos à PARTE VENDEDORA;"
        )

    # f) RECURSO PRÓPRIO
    if recurso_proprio:
        itens.append(
            f"f) {recurso_proprio}, em moeda corrente nacional, a serem transferidos à PARTE VENDEDORA em sua conta bancária "
            f"ou a conta de quem indicar no dia da assinatura da escritura perante instituição financeira competente;"
        )

    # g) CARTA DE CRÉDITO
    if carta_credito:
        itens.append(
            f"g) {carta_credito}, por intermédio de carta de crédito contemplada de titularidade da PARTE COMPRADORA;"
        )

    # h) PARCELAMENTO
    if parcelamento_total:
        itens.append(
            f"h) {parcelamento_total} em parcelas, sob os seguintes pagamentos:"
        )

        # ✅ se você tiver tela detalhada, encaixa o texto aqui
        if get("parcelamento_ativado", False) and get("parcelamento_descricao", "").strip():
            itens.append(f"<br><br>{get('parcelamento_descricao', '').strip()}")

    # i) OUTROS
    if outros:
        txt = f"i) {outros}, OUTROS"
        if outros_desc:
            txt += f": {outros_desc}"
        txt += ";"
        itens.append(txt)

    return "<br><br>".join(itens).strip()

def bloco_objeto() -> dict:
    """
    Retorna:
    - objeto: itens que devem ficar dentro do box "DO OBJETO DO CONTRATO"
    - secoes: itens que devem aparecer em boxes separados abaixo
    """

    # ============================
    # Dados do imóvel
    # ============================
    tipo_imovel = get("imovel__tipo", "").strip()  # ✅ NOVO (já existe no seu wizard)
    endereco_imovel = get("imovel__end__texto", "").strip()
    matricula = get("imovel__matricula", "").strip()
    cartorio = get("imovel__cartorio", "").strip()
    comarca = get("imovel__cidade_cartorio", "").strip()
    descricao_matricula = get("imovel__descricao_matricula", "").strip()
    contribuinte = get("imovel__contribuinte", "").strip()

    preco_total = get("preco_total", "").strip()

    # ============================
    # Helpers simples (gênero + preposição)
    # ============================
    def sufixo_situado(tipo: str) -> str:
        t = (tipo or "").lower()
        # feminino mais comum no seu conjunto
        if t.startswith("casa"):
            return "a"  # situada
        return "o"      # situado

    def preposicao_endereco(endereco: str) -> str:
        e = (endereco or "").strip().lower()
        # heurística: se começar por tipos comuns de logradouro, usar "na"
        if e.startswith(("rua ", "avenida ", "alameda ", "travessa ", "estrada ", "rodovia ")):
            return "na"
        # fallback seguro
        return "em"

    # ============================
    # Forma de pagamento (inalterado)
    # ============================
    texto_pagamento = pagamento_juridico()

    # ============================
    # Entrega de chaves (inalterado)
    # ============================
    texto_entrega = obter_clausula_entrega_chaves().strip()

    # ============================
    # OBJETO DO CONTRATO (um box único)
    # ============================
    linhas_objeto = []

    # ✅ PRIMEIRA LINHA: tipo + endereço (como você pediu)
    if endereco_imovel:
        tipo_txt = (tipo_imovel or "imóvel").strip()
        artigo_situado = sufixo_situado(tipo_txt)        # "o" ou "a"
        prep = preposicao_endereco(endereco_imovel)      # "na" ou "em"
        linhas_objeto.append(f"01 (um) {tipo_txt} situad{artigo_situado} {prep} {endereco_imovel}.")

    # ✅ Matrícula / Cartório / Comarca (dentro do bloco)
    linha_cartorio = []
    if matricula:
        linha_cartorio.append(f"MATRÍCULA: {matricula}")
    if cartorio:
        linha_cartorio.append(f"N.º DO CARTÓRIO: {cartorio}")
    if comarca:
        linha_cartorio.append(f"COMARCA DO CARTÓRIO: {comarca}")

    if linha_cartorio:
        linhas_objeto.append(" | ".join(linha_cartorio))

    # ✅ Descrição na matrícula (dentro do bloco)
    if descricao_matricula:
        linhas_objeto.append(descricao_matricula)

    # ✅ Nº do contribuinte (dentro do bloco)
    if contribuinte:
        linhas_objeto.append(f"Nº DO CONTRIBUINTE: {contribuinte}")

    texto_objeto = "<br><br>".join(linhas_objeto).strip()

    # ============================
    # SEÇÕES SEPARADAS (cada uma em um box)
    # ============================
    secoes = {}

    if preco_total:
        secoes["DO VALOR DO IMÓVEL"] = preco_total

    if texto_pagamento:
        secoes["DA FORMA DE PAGAMENTO DO PREÇO"] = texto_pagamento


    if texto_entrega:
        secoes["DO PRAZO DE ENTREGA DAS CHAVES DO IMÓVEL"] = texto_entrega

    return {
        "objeto": texto_objeto,
        "secoes": secoes
    }

def clausula_preambulo_clausulas_condicoes() -> str:
    """
    Texto imediatamente após o título 'DAS CLÁUSULAS E CONDIÇÕES'.
    Varia se houver financiamento ou não.
    """

    financiamento = get("preco_financiamento", "").strip()
    ha_financiamento = bool(financiamento)

    # ✅ variável conforme financiamento
    preambulo = "instituição financeira competente" if ha_financiamento else "tabelião de notas competente"

    return (
        "As partes qualificadas no quadro resumo pactuam entre si o presente compromisso de compra e venda "
        "do IMÓVEL, o qual será oportunamente aperfeiçoado mediante instrumento celebrado perante "
        f"{preambulo}, mediante as seguintes cláusulas e condições, a saber:"
    )

def nome_parte_assinatura(prefix: str) -> str:
    """
    Retorna o nome principal da parte para assinatura.
    - Se PF: retorna prefix__nome
    - Se PJ: retorna prefix__razao_social
    """
    tipo = get(f"{prefix}__tipo", "Pessoa Física").strip()

    if tipo == "Pessoa Jurídica":
        return get(f"{prefix}__razao_social", "").strip().upper()

    return get(f"{prefix}__nome", "").strip().upper()


def bloco_assinaturas_partes(titulo: str, lista_prefixos: list[str]) -> str:
    """
    Gera bloco de assinatura para N partes (PF ou PJ) com o formato:
    TITULO:
    ______________________
    NOME
    """
    if not lista_prefixos:
        return ""

    html = f"<b>{titulo}:</b><br><br>"

    for pfx in lista_prefixos:
        nome = nome_parte_assinatura(pfx)
        if not nome:
            continue

        html += (
            "<div style='border-bottom:1px solid #000; width:60%;'></div>"
            "<br>"
            f"<b>{nome}</b>"
            "<br><br><br>"
        )

    return html.strip()

# ============================================================
# CLÁUSULA (PLANILHA A FINAL!BH2 / BI2 / DW2)
# DECLARAÇÕES INICIAIS
# ============================================================

def titulo_clausula_01() -> str:
    return "DAS DECLARAÇÕES INICIAIS"

def clausula_bh2_abertura_matricula() -> str:
    """
    Replica exatamente a lógica da planilha A FINAL!BH2.

    Excel:
    =SE(OU(IMÓVEL!E7=IMÓVEL!M7; IMÓVEL!E7=IMÓVEL!P7; IMÓVEL!E7=IMÓVEL!S7); TEXTO; "")

    No seu sistema:
    - Aplica quando o tipo do imóvel contém "matrícula em área maior"
    """

    tipo_imovel = get("imovel__tipo", "").strip().lower()

    if "matrícula em área maior" in tipo_imovel:
        return (
            "A PARTE VENDEDORA declara que, na forma e sob as penas da lei, em relação à regularização da unidade "
            "perante o registro de imóveis competente, providenciará abertura da matrícula da unidade do empreendimento "
            "dentro de um prazo aproximado de até 90 (noventa) dias a partir da presente data, suportando o ônus de "
            "todas as despesas pertinentes para tanto."
        )

    return ""

def clausula_bi2_resilicao_por_forca_maior() -> str:
    """
    Replica exatamente a lógica da planilha A FINAL!BI2.

    Excel:
    =SE(OU(IMÓVEL!E7=IMÓVEL!M7; IMÓVEL!E7=IMÓVEL!P7; IMÓVEL!E7=IMÓVEL!S7); TEXTO; "")

    No seu sistema:
    - Aplica quando o tipo do imóvel contém "matrícula em área maior"
    """

    tipo_imovel = get("imovel__tipo", "").strip().lower()

    if "matrícula em área maior" in tipo_imovel:
        return (
            "Se, por caso fortuito e força maior, a PARTE VENDEDORA não conseguir providenciar a "
            "regularização da referida matrícula da unidade no prazo de até 90 (noventa) dias, este instrumento "
            "será extinto mediante resilição, ficando as partes contratantes isentas de multa contratual entre si, "
            "devendo assinar o instrumento extintivo do negócio ajustado num prazo máximo de até 5 (cinco) dias do "
            "vencimento do mencionado prazo de validade, comprometendo-se a PARTE VENDEDORA, ainda, se houver "
            "recebido ou se beneficiado de quaisquer valores e a qualquer título pagos ou desembolsados pela "
            "PARTE COMPRADORA, restituí-los no prazo de até 30 (trinta) dias a partir dos citados 90 (noventa) dias, "
            "sob pena de multa por infração contratual."
        )

    return ""

def clausula_dw2_alienacao_fiduciaria() -> str:
    """
    Replica exatamente a lógica da planilha (DW2).

    Excel:
    =SE(IMÓVEL!I31="NÃO";"";SE('PREÇO E ENTREGA DE CHAVES'!E17<>""; TEXTO_A; TEXTO_B))

    Aqui:
    - IMÓVEL!I31 -> get("imovel__alienado")  ("SIM"/"NÃO")
    - PREÇO E ENTREGA DE CHAVES!E17 -> get("preco_financiamento")
    """

    if get("imovel__alienado", "NÃO") != "SIM":
        return ""

    tem_financiamento = bool(get("preco_financiamento", "").strip())

    if tem_financiamento:
        return (
            "A PARTE COMPRADORA declara plena ciência de que o IMÓVEL ora se encontra alienado fiduciariamente a uma "
            "instituição financeira em razão de financiamento quando da aquisição da PARTE VENDEDORA, sendo que a quitação "
            "do financiamento contratado pela PARTE VENDEDORA será realizada por intermédio da instituição financeira "
            "competente ao financiamento a ser contratado pela PARTE COMPRADORA (interveniente quitante), conforme a forma "
            "de pagamento estipulada neste instrumento."
        )

    return (
        "A PARTE COMPRADORA declara plena ciência de que o IMÓVEL ora se encontra alienado fiduciariamente a uma instituição "
        "financeira em razão de financiamento quando da aquisição pela PARTE VENDEDORA, sendo que a quitação do financiamento "
        "contratado pela PARTE VENDEDORA será realizada por intermédio dos valores mencionados nas cláusulas seguintes."
    )

def clausula_bi2_propr_ou_posse() -> str:
    """
    Replica exatamente a lógica da planilha (BI2).

    Excel:
    =SE(DW2<>""; TEXTO_POSSE; TEXTO_PROPRIEDADE)

    Aqui:
    - DW2 equivale à cláusula de ALIENAÇÃO FIDUCIÁRIA (clausula_dw2_alienacao_fiduciaria()).
    - Se DW2 existir (não vazio) => retorna texto de POSSE.
    - Se DW2 não existir => retorna texto de PROPRIEDADE.
    """

    dw2 = clausula_dw2_alienacao_fiduciaria().strip()

    if dw2:
        return (
            "A PARTE VENDEDORA declara que é legítima possuidora do IMÓVEL com justo título, o qual está livre e "
            "desembaraçado de qualquer ônus ou gravame judicial, inclusive de natureza cível, trabalhista e/ou tributária; "
            "que não tem contra si qualquer protesto, ação e execução cível, criminal ou trabalhista cuja garantia pode vir "
            "a ser o IMÓVEL; que inexiste, a seu encargo, responsabilidade oriunda de tutela, curatela ou testamentária; "
            "que desconhece algo que possa impedir a presente transação, tanto ao IMÓVEL quanto à sua pessoa."
        )

    return (
        "A PARTE VENDEDORA declara que é proprietária e legítima possuidora do IMÓVEL com justo título, o qual está livre "
        "e desembaraçado de qualquer ônus ou gravame, judicial ou extrajudicial, inclusive de natureza cível, trabalhista "
        "e/ou tributária; que não tem contra si qualquer débito, protesto, ação e execução cível, criminal ou trabalhista "
        "cuja garantia pode vir a ser o IMÓVEL; que inexiste, a seu encargo, responsabilidade oriunda de tutela, curatela "
        "ou testamentária; que desconhece algo que possa impedir a presente transação, tanto ao IMÓVEL quanto à sua pessoa."
    )

def clausula_bi2_documentacao_processos() -> str:
    """
    Texto fixo (nível 3).
    Deve vir sempre imediatamente após clausula_bi2_propr_ou_posse().

    Conceito:
    - Se houver apontamento de ação, execução, protesto ou dívidas relativas ao imóvel,
      a PARTE VENDEDORA deve esclarecer e comprovar documentalmente até o prazo de validade,
      sob pena de multa por inadimplemento.
    """

    return (
        "Na hipótese de haver apontamento de distribuição de ação, execução judicial ou protesto "
        "contra a PARTE VENDEDORA, ou ainda débitos e/ou dívidas relativas ao IMÓVEL, a PARTE VENDEDORA "
        "compromete-se a prestar os esclarecimentos necessários à PARTE COMPRADORA ou à INTERMEDIADORA, "
        "mediante apresentação de cópias integrais dos processos, acesso aos autos digitais e/ou certidões "
        "negativas que comprovem inexistirem óbices à presente transação, tudo até o término do prazo de "
        "validade deste instrumento, sob pena de multa contratual por inadimplemento."
    )

def clausula_preço_forma_pagamento() -> str:
    
    return (
        "Pela presente transação, a PARTE VENDEDORA se compromete em transferir a propriedade do IMÓVEL à PARTE COMPRADORA mediante o recebimento de preço certo, líquido e exigível, conforme o preço do IMÓVEL e forma de pagamento do preço indicado no quadro resumo."
    )
    
def clausula_02_2_notas_pro() -> str:
    """
    Cláusula 2.2 («notas_pro»)

    Regra:
    - Se houver valor em preco_parcelamento_total, exibe o texto.
    - Se estiver vazio, retorna "".
    """

    parcelamento_total = get("preco_parcelamento_total", "").strip()

    if not parcelamento_total:
        return ""

    return (
        "As mencionadas parcelas no quadro resumo serão pagas mediante transferência bancária dos valores "
        "relativos a cada parcela em seu específico vencimento, na seguinte conta: "
        "Banco _________, Agência __________, conta ________________________, PIX: "
        "________________________., de titularidade de "
        "______________________________________________."
    )

def clausula_02_3_atraso() -> str:
    
    parcelamento_total = get("preco_parcelamento_total", "").strip()

    if not parcelamento_total:
        return ""

    return (
        "Em caso de mora nos pagamentos das parcelas no quadro resumo, as importâncias"
        "devidas serão acrescidas de multa moratória de 10% (dez por cento), mais juros de 0,033% (trinta e três milésimas por cento) ao dia"
        ", sendo tudo desde a data do vencimento até a data da liquidação da dívida."
        
        " Caso seja necessária a intervenção de advogado para eventuais cobranças extrajudiciais, a "
        "PARTE COMPRADORA será responsável pelos honorários advocatícios contratuais no importe de 10% (dez por cento) sobre os valores totais da dívida. "
        
        " Todavia, caso haja necessidade da PARTE VENDEDORA"
        " ou da INTERMEDIADORA ingressar com ação judicial para ver tutelado os seus direitos, serão devidos honorários advocatícios contratuais de 20% (vinte por cento), também, sobre os valores totais da dívida,"
        " a serem suportados integralmente pela PARTE COMPRADORA."
    )

def clausula_02_4_sinal() -> str:
    
    preco_sinal = get("preco_sinal", "").strip()

    if not preco_sinal:
        return ""

    return (
        "Com exceção das possibilidades do presente negócio ser desfeito por acordo entre as partes conforme previsão neste instrumento, "
        "as partes declaram plena ciência de que, em caso de descumprimento contratual pela PARTE VENDEDORA, ou qualquer outro ato impeditivo à conclusão do presente negócio por sua culpa exclusiva, "
        "esta parte ficará obrigada a pagar os valores ofertados a título de sinal e princípio de pagamento em dobro à PARTE COMPRADORA como indenização, "
        "nos termos dos artigos 417 até 419 do Código Civil, "
        "excluindo-se, neste caso, a eventual aplicação de multa por infração contratual."
    )

def clausula_02_5_sinal() -> str:
    
    preco_sinal = get("preco_sinal", "").strip()

    if not preco_sinal:
        return ""

    return (
        " Também, com exceção das possibilidades do presente negócio ser desfeito por acordo entre as partes conforme previsão neste instrumento,"
        " em caso de descumprimento contratual pela PARTE COMPRADORA, "
        "ou qualquer outro ato impeditivo à conclusão do presente negócio por sua culpa exclusiva,"
        " esta parte perderá o sinal em favor da PARTE VENDEDORA como indenização,"
        " nos termos dos artigos 417 até 419 do Código Civil, excluindo-se, neste caso, a eventual aplicação de multa por infração contratual."
    )

def clausula_03_1_financiamento_fgts() -> str:
    
    preco_financiamento = get("preco_financiamento", "").strip()

    if not preco_financiamento:
        return "As partes declaram pleno conhecimento de que o presente contrato será oportunamente aperfeiçoado, mediante novo instrumento celebrado perante o tabelião de notas, obrigando-se, desde já, a apresentarem todos os documentos exigidos às partes no momento oportuno à celebração de tal instrumento."

    return (
        "As partes declaram pleno conhecimento de que o presente contrato será oportunamente aperfeiçoado, "
        "mediante novo instrumento celebrado perante instituição financeira competente, obrigando-se, desde já, a apresentarem todos os documentos"
        " exigidos às partes no momento oportuno à celebração de tal instrumento."
    )

def clausula_03_2_financiamento_fgts() -> str:
    
    preco_financiamento = get("preco_financiamento", "").strip()

    if not preco_financiamento:
        return "As partes se obrigam a comparecer perante o tabelião de notas para a celebração e assinatura da respectiva escritura definitiva, em data e hora preestabelecida, sob a pena de multa de R$ 500,00 (quinhentos reais), os quais serão devidos à parte que cumpriu com a sua obrigação, salvo se o não comparecimento for dado em razão de casos fortuitos ou forças maiores, impossíveis de evitar ou impedir."

    return (
        "As partes se obrigam a comparecer perante instituição financeira competente para a celebração e assinatura da respectiva escritura definitiva, em data e hora preestabelecida, sob a pena de multa de R$ 500,00 (quinhentos reais) em face da parte que não comparecer, a qual será paga à parte que cumpriu com a sua obrigação, salvo se o não comparecimento for dado em razão de casos fortuitos ou forças maiores, impossíveis de evitar ou impedir."
    )

def clausula_03_3_inadimplencia() -> str:

    return (
        "A inadimplência da PARTE COMPRADORA em promover a lavratura da escritura definitiva de compra e venda no prazo pactuado isenta a PARTE VENDEDORA e eventualmente a INTERMEDIADORA da obrigação de apresentar novas certidões ou o seu teor."
    )

def clausula_03_4_1_financiamento_fgts() -> str:
    
    preco_financiamento = get("preco_financiamento", "").strip()

    if not preco_financiamento:
        return " pelo tabelião de notas competente "

    return (
        " pela instituição financeira competente "
    )

def clausula_03_4_2_financiamento_fgts() -> str:
    
    preco_financiamento = get("preco_financiamento", "").strip()

    if not preco_financiamento:
        return ""

    pela_pelo = clausula_03_4_1_financiamento_fgts()

    return (
        f"A PARTE COMPRADORA se obriga em protocolar o registro da escritura definitiva de venda e compra do IMÓVEL "
        f"lavrada{pela_pelo}em até 48 horas da sua respectiva posse deste documento, sob pena de multa diária no valor "
        f"de 0,5% (cinco décimas por cento) sobre o valor do IMÓVEL, salvo se tal protocolo de registro for intermediado ou procedido diretamente pela assessoria contratada pela PARTE COMPRADORA."
    )

def clausula_03_4_3_ITBI() -> str:
    
    preco_financiamento = get("preco_financiamento", "").strip()

    if not preco_financiamento:
        return "A PARTE COMPRADORA declara, neste ato, que lhe foram prestados amplos esclarecimentos acerca do presente contrato com relação a toda documentação, notadamente sobre as despesas com escrituração, como, também, Imposto de Transmissão de Bens Imóveis – ITBI, custas e emolumentos cartorários."

    pela_pelo = clausula_03_4_1_financiamento_fgts()

    return (
        f"A PARTE COMPRADORA se obriga em protocolar o registro da escritura definitiva de venda e compra do IMÓVEL "
        f"lavrada{pela_pelo}em até 48 horas da sua respectiva posse deste documento, sob pena de multa diária no valor "
        f"de 0,5% (cinco décimas por cento) sobre o valor do IMÓVEL, salvo se tal protocolo de registro for intermediado ou procedido diretamente pela assessoria contratada pela PARTE COMPRADORA."
    )

def titulo_04_financiamento_fgts() -> str:
    preco_financiamento = get("preco_financiamento", "").strip()
    preco_fgts = get("preco_fgts", "").strip()

    if preco_financiamento:
        return " DO FINANCIAMENTO" + (" E LIBERAÇÃO DO FGTS" if preco_fgts else "")

    return " DA LIBERAÇÃO DO FGTS" if preco_fgts else ""

def clausula_04_1_esclarecimentos_financiamento_fgts() -> str:
    preco_financiamento = get("preco_financiamento", "").strip()
    preco_fgts = get("preco_fgts", "").strip()

    # Parte fixa final (aparece em financiamento e fgts)
    final_comum = (
        " inclusive, sobre as despesas com assessoria, escrituração e/ou taxas da instituição financeira competente, "
        "como, também, Imposto de Transmissão de Bens Imóveis – ITBI, custas e emolumentos cartorários"
    )

    # Caso A: tem financiamento
    if preco_financiamento:
        meio = " as condições para o financiamento"
        if preco_fgts:
            meio += " e saque do FGTS, bem como, sobre as exigências do Sistema Financeiro de Habitação – SFH"
        return (
            " A PARTE COMPRADORA declara, neste ato, que lhe foi prestado amplos esclarecimentos acerca do presente contrato "
            "com relação a toda documentação, notadamente sobre" + meio + "," + final_comum
        )

    # Caso B: não tem financiamento, mas tem FGTS
    if preco_fgts:
        return (
            "A PARTE COMPRADORA declara, neste ato, que lhe foi prestado amplos esclarecimentos acerca do presente contrato "
            "com relação a toda documentação, notadamente sobre as condições para o saque do FGTS, "
            "bem como, sobre as exigências do Sistema Financeiro de Habitação – SFH," + final_comum
        )

    # Caso C: nenhum dos dois
    return (
        "A PARTE COMPRADORA declara, neste ato, que lhe foram prestados amplos esclarecimentos acerca do presente contrato "
        "com relação a toda documentação, notadamente sobre as despesas com escrituração, como, também, "
        "Imposto de Transmissão de Bens Imóveis – ITBI, custas e emolumentos cartorários"
    )

def clausula_04__2_qualidade_financiamento_fgts() -> str:
    preco_financiamento = get("preco_financiamento", "").strip()
    preco_fgts = get("preco_fgts", "").strip()

    if preco_financiamento:
        return (
            "A PARTE COMPRADORA declara que tem conhecimento da sistemática e exigências estabelecidas pela instituição financeira " "competente para a concessão do crédito pretendido, como, também, tem qualidade para cumprir integralmente todas as condições "
            "exigidas pela instituição financeira para a obtenção do financiamento"
            + (", bem como, para a obtenção dos valores vinculados à conta do Fundo de Garantia do Tempo de Serviço - FGTS."
               if preco_fgts else ".")
        )

    if preco_fgts:
        return (
            "A PARTE COMPRADORA declara que tem conhecimento da sistemática e exigências estabelecidas pela instituição financeira " "competente para a concessão do crédito pretendido, bem como, declara que tem qualidade para cumprir integralmente todas as condições "
            "exigidas para a obtenção dos valores vinculados à conta do Fundo de Garantia do Tempo de Serviço - FGTS."
        )
        
    return ""

def clausula_04__3_qualidade_financiamento_fgts() -> str:
    preco_financiamento = get("preco_financiamento", "").strip()
    preco_fgts = get("preco_fgts", "").strip()

    if preco_financiamento:
        return (
            "A PARTE COMPRADORA declara que tem conhecimento das atuais condições de resgate do financiamento a ser obtido, e reconhece e aceita o fato de que tais condições poderão sofrer"
            " modificações em razão de regulamentações supervenientes estabelecidas pelas autoridades governamentais ou pelo próprio órgão financiador que intervier na operação." + " A PARTE COMPRADORA se compromete, desde já, a suportar todos os ônus decorrentes de tais mudanças, em especial, no tocante à taxa nominal de juros ou outras condições econômico-financeiras, "
            "praticadas quando se der a assinatura do contrato perante órgão financiador, "
            "bem como, arcar com todo e qualquer tributo ou despesa que, por razões diversas, seja ou venha a ser cobrada, ou lançada, a qualquer título, em seu(s) nome(s)."
            )
        
    if preco_fgts:
        return (
            "A PARTE COMPRADORA declara que tem conhecimento das atuais condições de resgate do FGTS a ser obtido, e reconhece e aceita o fato de que tais condições poderão sofrer"
            " modificações em razão de regulamentações supervenientes estabelecidas pelas autoridades governamentais ou pela instituição financeira que intervier na operação."
        )
        
    return ""

def clausula_04__4_juizo_financiamento_fgts() -> str:
    preco_financiamento = get("preco_financiamento", "").strip()
    preco_fgts = get("preco_fgts", "").strip()

    if preco_financiamento:
        return (
            "As partes declaram ciência de que a instituição financeira competente, querendo, pode se reservar no direito de, ao seu juízo, não conceder os valores pretendidos caso a PARTE COMPRADORA"
            " não possua condições jurídicas ou socioeconômicas exigidas à época da análise à concessão do financiamento"
            + (",  e levantamento dos valores vinculados à conta do Fundo de Garantia do Tempo de Serviço - FGTS"
            ", ficando quaisquer diferença de valores sob ônus da PARTE COMPRADORA a serem pagos em moeda corrente nacional ou qualquer outro meio capaz de complementar os valores faltantes, a critério da PARTE VENDEDORA." +
            "<br>""<br>"
            "Caso não haja acordo entre as partes, o presente negócio será extinto sem quaisquer ônus aos envolvidos nesta transação, comprometendo-se a PARTE VENDEDORA,"
            " ainda, se houver recebido ou se beneficiado de quaisquer valores e a qualquer título pagos ou desembolsados pela PARTE COMPRADORA,"
            " restituí-los no prazo de até 30 (trinta) dias da não concessão dos valores pretendidos pela PARTE COMPRADORA nos termos acima, sob pena de multa por infração contratual."
               if preco_fgts else ".")
        )
           
    return ""

def clausula_05__1_juizo_entrega_chaves() -> str:
    
    return "A PARTE VENDEDORA se obriga a entregar a(s) chave(s) e o extrato das contas de consumo quitadas do IMÓVEL à PARTE COMPRADORA conforme o prazo indicado no quadro resumo, sob pena de multa diária no valor de R$ 100,00 (cem reais) à PARTE COMPRADORA, até a data efetiva da entrega da(s) referida(s) chave(s) e contas de consumo."

def clausula_05_2_livre_desocupado() -> str:
    
    return (
        "A PARTE VENDEDORA se compromete, ainda, a entregar o IMÓVEL livre e desocupado de pessoas e coisas, bem como, que arcará com as eventuais despesas de consumo de energia, água, gás, condomínio e IPTU até a entrega do IMÓVEL à PARTE COMPRADORA, "
        "sob pena de indenizá-la em caso de quaisquer prejuízos que venham a ocorrer em razão do não cumprimento ou satisfação de suas obrigações."
    )

def clausula_05_3_condominio() -> str:
    tipo_imovel = get("imovel__tipo", "").strip().lower()

    if tipo_imovel in ("casa", "terreno", "sobrado"):
        return ""

    return (
        "Caso seja IMÓVEL de condomínio, a PARTE VENDEDORA se compromete em apresentar a declaração de quitação de débito "
        "de taxas condominiais, com firma reconhecida do síndico (ou assinatura eletrônica pelo GOV.BR) e cópia autenticada da ata que elegeu o síndico ou "
        "administradora e, ainda, cópia da convenção e regulamento interno do condomínio, na assinatura do presente contrato, sob pena de multa por infração contratual."
    )

def clausula_06_1_transferencia_concessionaria() -> str:
    
    return (
        "A PARTE COMPRADORA se obriga a efetuar as transferências de titularidades das contas de consumo do IMÓVEL "
        "nas concessionárias de energia, água e gás, caso existam, no prazo máximo de 10 (dez) dias após receber a(s) chave(s) do IMÓVEL, sob pena de multa diária de R$ 50,00 (cinquenta reais), em favor da PARTE VENDEDORA."
    )

def clausula_06_1_transferencia_iptu() -> str:
    
    return (
        "A PARTE COMPRADORA se obriga, também, a providenciar a transferência do IPTU na prefeitura do município do IMÓVEL (caso esteja individualizado) no prazo máximo de 60 (sessenta) dias, "
        "a partir da data do registro da escritura, conforme a Lei n.º 10.819, de 28/12/1989 e Decreto n.º 28.494, de 09/01/1990, também, sob pena de multa diária de R$ 50,00 (cinquenta reais), "
        "em favor da PARTE VENDEDORA, até a data da apresentação dos protocolos de transferência perante prefeitura do município do IMÓVEL."
    )

def clausula_07_1_honorarios() -> str:
    quem_paga_comissao = get("quem_paga_comissao", "").strip()

    if quem_paga_comissao in ("PARTE VENDEDORA"):
        return (
            "Fica convencionado que a  PARTE VENDEDORA pagará a comissão pelos trabalhos ora praticados pela INTERMEDIADORA e seus corretores associados, nos termos do contrato de corretagem apresentado à PARTE VENDEDORA juntamente com este instrumento.")
    
    if quem_paga_comissao in ("PARTE COMPRADORA"):
        return (
            "Fica convencionado que a PARTE COMPRADORA pagará a comissão pelos trabalhos ora praticados pela INTERMEDIADORA e seus corretores associados, nos termos do contrato de corretagem apresentado à PARTE COMPRADORA juntamente com este instrumento.")              
    
    #if quem_paga_comissao in ("AMBAS AS PARTES"):
        #return (
            #"Fica convencionado que a comissão devida à INTERMEDIADORA pelos trabalhos oferecidos e praticados a ambas as partes do presente negócio, fixada nos valores de ";$AU$2;", será rateada entre a PARTE VENDEDORA e a PARTE COMPRADORA, na seguinte forma:")              
    
    return ""

def clausula_07_2_honorarios() -> str:
    quem_paga_comissao = get("quem_paga_comissao", "").strip()

    if quem_paga_comissao in ("PARTE VENDEDORA"):
        return (
            "A INTERMEDIADORA terá direito ao recebimento da comissão independentemente do referido contrato de corretagem. "
            "Caso a PARTE VENDEDORA não assine o referido contrato de corretagem com a INTERMEDIADORA, desde já, responsabilizar-se-á pelo pagamento da comissão com base na tabela mínima estabelecida pelo CRECI, "
            "sendo 6% (seis por cento) sobre o valor do IMÓVEL.")
    
    if quem_paga_comissao in ("PARTE COMPRADORA"):
        return (
            "A INTERMEDIADORA terá direito ao recebimento da comissão independentemente do referido contrato de corretagem. "
            "Caso a PARTE COMPRADORA não assine o referido contrato de corretagem com a INTERMEDIADORA, desde já, responsabilizar-se-á pelo pagamento da comissão com base na tabela mínima estabelecida pelo CRECI, "
            "sendo 6% (seis por cento) sobre o valor do IMÓVEL.")
    
    return ""

def clausula_07_3_honorarios() -> str:
    quem_paga_comissao = get("quem_paga_comissao", "").strip()

    if quem_paga_comissao in ("PARTE VENDEDORA"):
        return (
            "Caso seja necessária a intervenção de advogado para eventuais cobranças extrajudiciais, a PARTE VENDEDORA será responsável pelos honorários advocatícios contratuais no importe de 10% (dez por cento) sobre os valores totais da dívida. Todavia, caso haja necessidade de a INTERMEDIADORA ingressar com ação judicial para ver tutelado os seus direitos, serão devidos honorários advocatícios contratuais de 20% (vinte por cento), também, sobre os valores totais da dívida, a serem suportados integralmente pela PARTE VENDEDORA.")
    
    if quem_paga_comissao in ("PARTE COMPRADORA"):
        return (
            "Caso seja necessária a intervenção de advogado para eventuais cobranças extrajudiciais, a PARTE COMPRADORA será responsável pelos honorários advocatícios contratuais no importe de 10% (dez por cento) sobre os valores totais da dívida. Todavia, caso haja necessidade de a INTERMEDIADORA ingressar com ação judicial para ver tutelado os seus direitos, serão devidos honorários advocatícios contratuais de 20% (vinte por cento), também, sobre os valores totais da dívida, a serem suportados integralmente pela PARTE COMPRADORA.")
    
    return ""
    
def clausula_07_4_honorarios() -> str:
    
    return "A falta de qualquer pagamento por si só constituirá a PARTE responsável em mora, independentemente de qualquer aviso ou interpelação judicial ou extrajudicial."

def clausula_08_1_prazo_conclusao() -> str:
    
    parcelamento = get("preco_parcelamento_total", "").strip()  # V2
    financiamento = get("preco_financiamento", "").strip()      # L3
    fgts = get("preco_fgts", "").strip()                        # N3
    tipo_imovel = get("tipo_imovel", "").strip().lower()        # IMÓVEL!E7

    # Se existe parcelamento -> retorna vazio (como no Excel)
    if parcelamento:
        return ""

    # Tipos de imóvel que são "matrícula em área maior"
    tipos_matricula_area_maior = {
        "apartamento (matrícula em área maior)",
        "sobrado em condomínio (matrícula em área maior)",
        "casa em condomínio (matrícula em área maior)",
    }

    eh_matricula_area_maior = tipo_imovel in tipos_matricula_area_maior

    # Textos (equivalentes aos CONCATENAR do Excel)
    texto_60_area_maior = (
        " O presente instrumento tem o prazo de validade de 60 (sessenta) dias à sua conclusão e/ou integral "
        "cumprimento em seus termos dispostos a contar da data da efetiva regularização da referida matricula da "
        "unidade conforme estipulado na cláusula 1.1, podendo as partes, se vencido tal prazo sem o integral "
        "cumprimento deste instrumento e sem culpa de qualquer delas, manifestarem-se sobre a resilição do presente "
        "negócio em até 24 (vinte quatro) horas, sob a possibilidade deste instrumento se prorrogar automaticamente "
        "pelo período de mais 30 (trinta) dias."
    )

    texto_120_area_maior = (
        " O presente instrumento tem o prazo de validade de 120 (cento e vinte) dias à sua conclusão e/ou integral "
        "cumprimento em seus termos dispostos a contar da data da efetiva regularização da referida matricula da "
        "unidade conforme estipulado na cláusula 1.1, podendo as partes, se vencido tal prazo sem o integral "
        "cumprimento deste instrumento e sem culpa de qualquer delas, manifestarem-se sobre a resilição do presente "
        "negócio em até 24 (vinte quatro) horas, sob a possibilidade deste instrumento se prorrogar automaticamente "
        "pelo período de mais 60 (sessenta) dias."
    )

    texto_60_normal = (
        " O presente instrumento tem o prazo de validade de 60 (sessenta) dias à sua conclusão e/ou integral "
        "cumprimento, em seus termos dispostos, a contar da data indicada no final do presente contrato, com "
        "respectivas assinaturas das partes, podendo as partes, se vencido tal prazo sem o integral cumprimento deste "
        "instrumento e sem culpa de qualquer delas, manifestarem-se sobre a resilição do presente negócio em até 24 "
        "(vinte quatro) horas, sob a possibilidade deste instrumento se prorrogar automaticamente pelo período de "
        "mais 30 (trinta) dias."
    )

    texto_120_normal = (
        " O presente instrumento tem o prazo de validade de 120 (cento e vinte) dias à sua conclusão e/ou integral "
        "cumprimento em seus termos dispostos, a contar da data indicada no final do presente contrato, com "
        "respectivas assinaturas das partes, podendo as partes, se vencido tal prazo sem o integral cumprimento deste "
        "instrumento e sem culpa de qualquer delas, manifestarem-se sobre a resilição do presente negócio em até 24 "
        "(vinte quatro) horas, sob a possibilidade deste instrumento se prorrogar automaticamente pelo período de "
        "mais 60 (sessenta) dias."
    )

    # ✅ 1) Se L3 e N3 vazios e tipo é matrícula em área maior -> 60 dias (área maior)
    if (not financiamento) and (not fgts) and eh_matricula_area_maior:
        return texto_60_area_maior

    # ✅ 2) Se L3 e N3 preenchidos e tipo é matrícula em área maior -> 120 dias (área maior)
    if financiamento and fgts and eh_matricula_area_maior:
        return texto_120_area_maior

    # ✅ 3) Se L3 e N3 vazios -> 60 dias (normal)
    if (not financiamento) and (not fgts):
        return texto_60_normal

    # ✅ 4) Caso contrário -> 120 dias (normal)
    return texto_120_normal

def clausula_08_2_resilicao_por_prazo() -> str:
    parcelamento = get("preco_parcelamento_total", "").strip()  # V2

    if parcelamento:
        return ""

    return (
        " Vencendo este último prazo, também, sem a conclusão e/ou integral cumprimento "
        "do presente compromisso de compra e venda do IMÓVEL em seus termos e sem qualquer culpa das partes, "
        "este instrumento poderá ser extinto mediante resilição, ficando a PARTE VENDEDORA e PARTE COMPRADORA "
        "isentas de qualquer penalidade ou multa contratual entre si, devendo assinar o instrumento extintivo "
        "do negócio ajustado num prazo máximo de até 5 (cinco) dias do vencimento do mencionado prazo de validade."
    )

def clausula_08_3_resilicao_por_prazo() -> str:
    parcelamento = get("preco_parcelamento_total", "").strip()  # V2

    if parcelamento:
        return ""

    return (
        "Nesta hipótese, a PARTE VENDEDORA se comprometendo, ainda, se houver recebido ou se beneficiado de quaisquer "
        "valores e a qualquer título pagos ou desembolsados pela PARTE COMPRADORA, restituí-los no prazo de até "
        "30 (trinta) dias da assinatura do referido instrumento extintivo, sob pena de multa por infração contratual."
    )

def clausula_09_1_resolucao() -> str:
    preco_sinal = get("preco_sinal", "").strip()  # J3

    if preco_sinal:
        return (
            " Com exceção das possibilidades do presente negócio ser desfeito por acordo entre as partes conforme previsão "
            "neste instrumento, a parte que sofrer lesão por inadimplemento e culpa da outra parte poderá, além de ter os "
            "valores equivalentes de sinal como verbas indenizatórias, pedir a resolução do contrato, bem como, indenização "
            "suplementar."
        )

    return (
        "  Com exceção das possibilidades do presente negócio ser desfeito por acordo entre as partes conforme previsão "
        "neste instrumento, a parte que sofrer lesão por inadimplemento e culpa da outra parte poderá pedir a resolução do "
        "contrato, se não preferir lhe exigir o seu integral cumprimento, cabendo, ainda, em qualquer dos casos, multa de "
        "6% (seis por cento) sobre o valor total do IMÓVEL, além de indenização por perdas e danos se provar maior prejuízo."
    )

def clausula_09_2_desist_com_sinal() -> str:
    preco_sinal = get("preco_sinal", "").strip()  # J3

    if preco_sinal:
        return "Caso a desistência seja realizada pela PARTE VENDEDORA, deverá a PARTE COMPRADORA ser reembolsada na integralidade de valores pagos a este título, sem prejuízo do exposto acima e multa por infração contratual."

    return ""

def clausula_09_3_desist_com_sinal() -> str:
    preco_sinal = get("preco_sinal", "").strip()  # J3

    if preco_sinal:
        return "Caso a parte inocente preferir exigir o integral cumprimento do presente compromisso da parte o infringiu, poderá, ainda, requerer indenização por perdas e danos, valendo, também, as arras como o mínimo da indenização."

    return ""

def clausula_09_4_desist_com_sinal() -> str:
    preco_sinal = get("preco_sinal", "").strip()  # J3

    if preco_sinal:
        return "A parte que der causa à resolução do presente contrato, será, também, responsável pelo pagamento dos honorários à INTERMEDIADORA, do presente contrato, bem como, todas as suas despesas com documentações e honorários advocatícios contratuais, desde já, estabelecidos em 20% (vinte por cento) do valor do débito."

    return ""

def clausula_10_1_irretratabilidade() -> str:
    preco_financiamento = get("preco_financiamento", "").strip()
    preco_fgts = get("preco_fgts", "").strip()

    if preco_financiamento:
        return "O presente contrato é celebrado em caráter irretratável e irrevogável, obrigando não só as partes, mas, também, seus herdeiros e sucessores, não se admitindo o arrependimento de quaisquer das partes por quaisquer tipos de pretextos ou alegações, salvo o disposto na cláusula que trata sobre o prazo de validade deste compromisso, bem como, outras possibilidades do presente negócio ser desfeito por acordo entre as partes conforme previsão neste instrumento."

    if preco_fgts:
        return "O presente contrato é celebrado em caráter irretratável e irrevogável, obrigando não só as partes, mas, também, seus herdeiros e sucessores, não se admitindo o arrependimento de quaisquer das partes por quaisquer tipos de pretextos ou alegações, salvo o disposto na cláusula que trata sobre o prazo de validade deste compromisso, bem como, outras possibilidades do presente negócio ser desfeito por acordo entre as partes conforme previsão neste instrumento."
    
    return "O presente contrato é celebrado em caráter irretratável e irrevogável, obrigando não só as partes, mas, também, seus herdeiros e sucessores, não se admitindo o arrependimento de quaisquer das partes por quaisquer tipos de pretextos ou alegações, salvo eventuais possibilidades do presente negócio ser desfeito por acordo entre as partes conforme previsão neste instrumento."

def clausula_11_1_vicios() -> str:
        
    return "A PARTE VENDEDORA declara, na forma e sob as penas da lei, que responde pela evicção de direito, quando chamada à autoria em demandas judiciais e administrativas, e pelos vícios redibitórios em relação ao IMÓVEL ora transacionado, desde que seja constatado que tais vícios se originaram antes do presente negócio."

def clausula_11_2_vicios() -> str:
        
    return "Quaisquer dívidas da PARTE VENDEDORA que venham, eventualmente e a qualquer tempo, atingir o IMÓVEL, causando-lhe constrição judicial, bloqueio ou anulação do presente negócio, dá à PARTE COMPRADORA o direito quitar eventuais dívidas, de modo que não perca o IMÓVEL, podendo, ainda, pleitear judicialmente quaisquer perdas e danos sofridas em razão destes fatos."

def clausula_12_titulo_declaracoes() -> str:
    
    imovel__ficara_bens = get("imovel__ficara_bens", "").strip().upper()  # IMÓVEL!I33

    if imovel__ficara_bens == "SIM":
        return " DAS DECLARAÇÕES DAS PARTES EM RELAÇÃO AO IMÓVEL"

    if imovel__ficara_bens in ("NÃO", "NAO", ""):
        return " DA DECLARAÇÃO DA PARTE COMPRADORA EM RELAÇÃO AS CONDIÇÕES DO IMÓVEL"

    return ""

def clausula_12_1_ficara_bens() -> str:
    
    return "A PARTE COMPRADORA declara que visitou o IMÓVEL ora transacionado, aceitando-o no estado em que se encontra, estando ciente que após a assinatura deste compromisso não poderá reivindicar quaisquer reparos, com exceção à evicção de direito e vícios redibitórios."    

def clausula_12_2_ficara_bens() -> str:
    
    return "As partes convencionam que a presente venda do IMÓVEL é feita na forma “AD CORPUS”, ou seja, assim como está, independentemente das medidas."

def clausula_12_3_ficara_bens() -> str:
    imovel__ficara_bens = get("imovel__ficara_bens", "").strip().upper()   # I33
    imovel__bens = get("imovel__bens", "").strip()             # I35

    # Se for "NÃO" ou estiver vazio -> não exibe a cláusula
    if imovel__ficara_bens in ("NÃO", "NAO", ""):
        return ""

    # Se chegou aqui, presume-se que é "SIM" (ou equivalente)
    return (
        " A PARTE VENDEDORA declara que ficará integrado ao IMÓVEL e vinculado ao presente negócio: "
        f"{imovel__bens.lower()}."
    )

def clausula_13_1_termino_pretacao() -> str:
    preco_financiamento = get("preco_financiamento", "").strip()  # L3
    preco_fgts = get("preco_fgts", "").strip()                    # N3

    if preco_financiamento or preco_fgts:
        return (
            " Fica devidamente esclarecido às partes, ora contratantes, "
            "que a prestação de serviço da INTERMEDIADORA se aperfeiçoa com a assinatura do presente instrumento, contudo, "
            "acompanhará e auxiliará perante o competente cartório de registro de imóveis e o desbloqueio dos valores "
            "dos recursos na conta da PARTE VENDEDORA, "
            "não assumindo, neste segundo momento, qualquer responsabilidade ou encargo, tendo em vista que a sua prestação de serviço já fora "
            "totalmente concluída, em razão o fechamento da transação imobiliária."
        )

    return (
        " Fica devidamente esclarecido às partes, ora contratantes, que a prestação de serviço da "
        "INTERMEDIADORA se aperfeiçoa com a assinatura do presente instrumento, contudo, "
        "acompanhará e auxiliará perante o competente cartório de registro de imóveis, não assumindo, "
        "neste segundo momento, qualquer responsabilidade ou encargo, "
        "tendo em vista que a sua prestação de serviço já fora totalmente concluída, "
        "em razão o fechamento da transação imobiliária."
    )

def clausula_13_2_termino_pretacao() -> str:
    
    return (
        "As partes declaram que a INTERMEDIADORA lhes prestou todos os esclarecimentos necessários à presente transação, prestando-lhes, também, toda assistência necessária sob o devido zelo para que este negócio jurídico se realize com segurança, informando-lhes, ainda, sobre a necessidade de extrações das certidões necessárias por vias próprias e particulares, bem como, sobre eventuais riscos e toda situação documental apresentada das partes e do IMÓVEL."
    )
    
def Clausula_13_3_responsabilidade_intermediadora() -> str:
    preco_financiamento = get("preco_financiamento", "").strip()  # DA26
    preco_fgts = get("preco_fgts", "").strip()                    # DA28
    preco_carta_credito = get("preco_carta_credito", "").strip()        # DA30

    tem_fin = bool(preco_financiamento)
    tem_fgts = bool(preco_fgts)
    tem_carta = bool(preco_carta_credito)

    # Caso 0: nenhum meio especial => não exibe cláusula
    if not (tem_fin or tem_fgts or tem_carta):
        return ""

    # Textos por combinação (fiel à fórmula)
    if tem_fin and (not tem_fgts) and (not tem_carta):
        return (
            " A INTERMEDIADORA não será responsável por quaisquer resultados negativos quanto à obtenção do financiamento ou qualquer "
            "outra impossibilidade que venha a surgir em razão deste meio de pagamento que possa atrasar ou extinguir o presente negócio, "
            "sendo de total responsabilidade das partes o preenchimento e atendimento das condições impostas pela instituição financeira "
            "ou empresa competente."
        )

    if tem_fin and tem_fgts and (not tem_carta):
        return (
            " A INTERMEDIADORA não será responsável por quaisquer resultados negativos quanto à obtenção do financiamento e resgate do FGTS, "
            "bem como, por qualquer outra impossibilidade que venha a surgir em razão destes meios de pagamentos que possam atrasar ou extinguir "
            "o presente negócio, sendo de total responsabilidade das partes o preenchimento e atendimento das condições impostas pela instituição "
            "financeira ou empresa competente."
        )

    if tem_fin and tem_fgts and tem_carta:
        return (
            " A INTERMEDIADORA não será responsável por quaisquer resultados negativos quanto à obtenção do financiamento, resgate do FGTS "
            "ou utilização ou transferência dos valores ou direitos da carta crédito mencionada neste instrumento, bem como, por qualquer "
            "outra impossibilidade que venha a surgir em razão destes meios de pagamentos que possam atrasar ou extinguir o presente negócio, "
            "sendo de total responsabilidade das partes o preenchimento e atendimento das condições impostas pela instituição financeira ou "
            "empresa competente."
        )

    if (not tem_fin) and tem_fgts and tem_carta:
        return (
            " A INTERMEDIADORA não será responsável por quaisquer resultados negativos quanto ao resgate do FGTS ou utilização ou transferência "
            "dos valores ou direitos da carta crédito mencionada neste instrumento, bem como, por qualquer outra impossibilidade que venha a surgir "
            "em razão destes meios de pagamentos que possam atrasar ou extinguir o presente negócio, sendo de total responsabilidade das partes o "
            "preenchimento e atendimento das condições impostas pela instituição financeira ou empresa competente."
        )

    if (not tem_fin) and (not tem_fgts) and tem_carta:
        return (
            " A INTERMEDIADORA não será responsável por quaisquer resultados negativos quanto à utilização ou transferência dos valores ou direitos "
            "da carta crédito mencionada neste instrumento, bem como, por qualquer outra impossibilidade que venha a surgir em razão destes meios "
            "de pagamentos que possam atrasar ou extinguir o presente negócio, sendo de total responsabilidade das partes o preenchimento e "
            "atendimento das condições impostas pela instituição financeira ou empresa competente."
        )

    if tem_fin and (not tem_fgts) and tem_carta:
        return (
            " A INTERMEDIADORA não será responsável por quaisquer resultados negativos quanto à obtenção do financiamento ou à utilização ou "
            "transferência dos valores ou direitos da carta crédito mencionada neste instrumento, bem como, por qualquer outra impossibilidade "
            "que venha a surgir em razão destes meios de pagamentos que possam atrasar ou extinguir o presente negócio, sendo de total "
            "responsabilidade das partes o preenchimento e atendimento das condições impostas pela instituição financeira ou empresa competente."
        )

    # Cobertura extra: se cair em uma combinação não prevista, não mostra nada
    return ""

def Clausula_13_4_responsabilidade_intermediadora() -> str:
        
    return "Porventura houver quaisquer tipos de problemas posteriores a conclusão do presente negócio, as partes poderão providenciar nova tratativa de prestação de serviços perante a INTERMEDIADORA, seja no setor imobiliário ou no setor jurídico."

def Clausula_13_5_responsabilidade_intermediadora() -> str:
        
    return "Caso o presente negócio não se conclua por qualquer que seja o motivo ou por arrependimento de qualquer das partes e, posteriormente, as partes realizem a compra e venda diretamente entre si e sem a participação da INTERMEDIADORA, ser-lhe-ão devidos os honorários ajustados de 6% (seis por cento) sobre o valor do IMÓVEL, a qual será suportada solidariamente entre a PARTE VENDEDORA e a PARTE COMPRADORA, além de suportarem, também, solidariamente, as custas e despesas processuais e honorários advocatícios que, desde já, ficam estabelecidos em 20% sobre o valor total devido."

def clausula_14_1_disposicoes_gerais() -> str:
        
    return "Caso a PARTE COMPRADORA tenha interesse em registrar este compromisso junto ao competente Cartório de Registro de Imóveis, tais despesas correrão exclusivamente por sua conta."

def clausula_14_2_procuracao_vendedora() -> str:
    vendedores = get_list("vendedores")
    if len(vendedores) <= 1:
        return ""

    return (
        "Todos os integrantes da PARTE VENDEDORA se nomeiam e se constituem reciprocamente "
        "procuradores, bastante para receberem citações, intimações ou interpelações provenientes "
        "de eventual ação judicial ou extrajudicial, movida a qualquer um deles em razão do presente negócio."
    )

def clausula_14_3_procuracao_compradora() -> str:
    compradores = get_list("compradores")
    if len(compradores) <= 1:
        return ""

    return (
        "Todos os integrantes da PARTE COMPRADORA se nomeiam e se constituem reciprocamente "
        "procuradores, bastante para receberem citações, intimações ou interpelações provenientes "
        "de eventual ação judicial ou extrajudicial, movida a qualquer um deles em razão do presente negócio."
    )

def clausula_14_4_intimacoes() -> str:
    
    return (
        "Todos os integrantes da PARTE VENDEDORA se nomeiam e se constituem reciprocamente "
        "procuradores, bastante para receberem citações, intimações ou interpelações provenientes "
        "de eventual ação judicial ou extrajudicial, movida a qualquer um deles em razão do presente negócio."
    )

def clausula_14_5_comunicar_endereco() -> str:
    
    return (
        "A PARTE COMPRADORA e a PARTE VENDEDORA se obrigam mutuamente em comunicar eventuais mudanças de endereço, telefone celular, inclusive, correio eletrônico, presumindo-se válidas as citações, intimações ou notificações ao endereço constante neste instrumento ou ao endereço do IMÓVEL, ainda que não recebidas pessoalmente pelo interessado, se a modificação temporária ou definitiva não tiver sido devidamente comunicada nos termos expostos."
    )

def clausula_14_6_alterar_endereco() -> str:
    
    return (
        "Qualquer alteração de condição deste instrumento deverá ser formalizada via aditamento contratual devidamente assinado pelas partes em conjunto com duas testemunhas, sendo qualquer outro acordo realizado pelas partes de modo extracontratual considerados como mera tolerância e sem o efeito de novar o disposto neste instrumento."
    )

def clausula_15_1_foro() -> str:
    vendedores = get_list("vendedores")
    compradores = get_list("compradores")

    # ✅ Títulos automáticos conforme tipo do contrato
    titulo_vendedor = papel_parte_vendedora_ou_cedente()          # "PARTE VENDEDORA" ou "PARTE CEDENTE"
    titulo_comprador = papel_parte_compradora_ou_cessionaria()     # "PARTE COMPRADORA" ou "PARTE CESSIONÁRIA"

    return (
        "Fica eleito o foro da situação do IMÓVEL, com expressa renúncia a qualquer outro, por mais privilegiado que seja, "
        "para dirimir quaisquer questões oriundas do presente contrato."
        "<br><br>"
        "Por estarem assim justas e contratadas, sob declaração da expressão da verdade de todo o exposto acima, inclusive "
        "de seus dados e informações pessoais, as partes assinam o presente contrato em 03 (três) vias de igual teor e forma, "
        "na presença de duas testemunhas, para que produza seus normais efeitos de direito."
        "<br><br>"

        # ✅ DATA À DIREITA
        f"<div style='text-align:right;'>{linha_local_data()}</div>"
        "<br><br><br>"

        # ✅ ASSINATURAS: PARTE VENDEDORA/CEDENTE
        + bloco_assinaturas_partes(titulo_vendedor, vendedores)

        # ✅ ASSINATURAS: PARTE COMPRADORA/CESSIONÁRIA
        + bloco_assinaturas_partes(titulo_comprador, compradores)

        # ✅ TESTEMUNHAS
        + (
            "<b>TESTEMUNHAS:</b>"
            "<br><br>"
            "<div style='border-bottom:1px solid #000; width:60%;'></div>"
            "<br>"
            "Nome:"
            "<br>"
            "CPF:"
            "<br><br><br>"
            "<div style='border-bottom:1px solid #000; width:60%;'></div>"
            "<br>"
            "Nome:"
            "<br>"
            "CPF:"
        )
    )







# ============================================================
# SIDEBAR
# ============================================================

st.markdown("""
<style>
/* ====== MENU (radio) com aparência de botões ====== */
section[data-testid="stSidebar"] div[role="radiogroup"] label {
    background: transparent;
    border: 1px solid rgba(255,255,255,0.12);
    border-radius: 10px;
    padding: 10px 12px;
    margin-bottom: 8px;
    width: 100%;
    display: flex;
    align-items: center;
}

section[data-testid="stSidebar"] div[role="radiogroup"] label:hover {
    border: 1px solid rgba(255,255,255,0.25);
}

/* esconde o bolinho do radio */
section[data-testid="stSidebar"] div[role="radiogroup"] label input {
    display: none;
}

/* texto */
section[data-testid="stSidebar"] div[role="radiogroup"] label span {
    font-weight: 600;
    width: 100%;
}

/* ====== ITEM SELECIONADO = LARANJA ====== */
section[data-testid="stSidebar"] div[role="radiogroup"] label:has(input:checked) {
    background-color: #f57c00 !important;
    border: 1px solid rgba(0,0,0,0.12) !important;
}

section[data-testid="stSidebar"] div[role="radiogroup"] label:has(input:checked) span {
    color: white !important;
}
</style>
""", unsafe_allow_html=True)



st.sidebar.markdown("<hr style='opacity:0.2;'>", unsafe_allow_html=True)

st.sidebar.markdown("<h3 style='margin:0;'>📌 Etapas</h3>", unsafe_allow_html=True)

progress = (st.session_state.step_index + 1) / len(steps())
st.sidebar.progress(progress)

st.sidebar.markdown("<hr style='opacity:0.2;'>", unsafe_allow_html=True)

# ✅ Lista apenas das telas visíveis (não hidden)
steps_visiveis = [s for s in steps() if not s.get("hidden")]
labels = [f"{i+1}. {s['title']}" for i, s in enumerate(steps_visiveis)]

# ✅ Índice atual dentro da lista visível
idx_atual_visivel = 0
for i, s in enumerate(steps_visiveis):
    if steps().index(s) == st.session_state.step_index:
        idx_atual_visivel = i
        break

# ✅ Mantém o radio SEMPRE sincronizado com o step_index atual
label_atual = labels[idx_atual_visivel]
st.session_state["sidebar_nav_radio"] = label_atual

def _on_sidebar_nav_change():
    escolha = st.session_state.get("sidebar_nav_radio", label_atual)
    novo_idx_visivel = labels.index(escolha)
    novo_step_id = steps_visiveis[novo_idx_visivel]["id"]
    go_to_step(novo_step_id)
    st.rerun()

# ✅ Radio como menu (permite estilizar o selecionado)
st.sidebar.radio(
    " ",
    labels,
    key="sidebar_nav_radio",
    on_change=_on_sidebar_nav_change
)

st.sidebar.markdown("<hr style='opacity:0.2;'>", unsafe_allow_html=True)

st.sidebar.markdown("---")
st.sidebar.write(f"👤 Usuário: **{st.session_state.get('auth_user','')}**")
if st.sidebar.button("Sair", key="btn_logout"):
    do_logout()

# ============================================================
# ÍNDICE DE CLÁUSULAS (dinâmico) + NUMERAÇÃO AUTOMÁTICA
# ============================================================

def tem_financiamento():
    return bool(get("preco_financiamento", "").strip())

def tem_fgts():
    return bool(get("preco_fgts", "").strip())

def imovel_alienado():
    return get("imovel__alienado", "NÃO") == "SIM"

def render_subclausulas_dinamicas(numero_clausula_principal: int, textos: list[str], tamanho_px: int = 15):
    """
    Renderiza subcláusulas com numeração dinâmica:
      1.1, 1.2, 1.3...
    conforme os textos efetivamente presentes (não vazios).
    """
    contador = 1
    for t in textos:
        if not t or not t.strip():
            continue
        prefixo = f"{numero_clausula_principal}.{contador}. "
        texto_justificado(prefixo + t.strip(), tamanho_px=tamanho_px)
        st.markdown("<br>", unsafe_allow_html=True)
        contador += 1

# Cada cláusula: título + regra de visibilidade + render
CLAUSULAS = [
    {
        "id": "cl01",
        "titulo": "DAS DECLARAÇÕES INICIAIS",
        "visivel": lambda: True,
        "render": lambda numero=1: (
            render_subclausulas_dinamicas(
                numero_clausula_principal=numero,
                textos=[
                    clausula_bh2_abertura_matricula(),
                    clausula_bi2_resilicao_por_forca_maior(),
                    clausula_bi2_propr_ou_posse(),
                    clausula_bi2_documentacao_processos(),
                    clausula_dw2_alienacao_fiduciaria()
                ],
                tamanho_px=15
            )
        )
    },

    {
        "id": "cl02",
        "titulo": "DO PREÇO E FORMA DE PAGAMENTO",
        "visivel": lambda: True,
        "render": lambda numero=2: (
            render_subclausulas_dinamicas(
                numero_clausula_principal=numero,
                textos=[
                    clausula_preço_forma_pagamento(),
                    clausula_02_2_notas_pro(),
                    clausula_02_3_atraso(),
                    clausula_02_4_sinal(),
                    clausula_02_5_sinal(),
                ],
                tamanho_px=15)
        )
    },
    {
        "id": "cl03",
        "titulo": "DA ESCRITURA DEFINITIVA",
        "visivel": lambda: True,
        "render": lambda numero=3: (
            render_subclausulas_dinamicas(
                numero_clausula_principal=numero,
                textos=[
                    clausula_03_1_financiamento_fgts(),
                    clausula_03_2_financiamento_fgts(),
                    clausula_03_3_inadimplencia(),
                    clausula_03_4_2_financiamento_fgts(),
                    clausula_03_4_3_ITBI(),
                    
                ],
                tamanho_px=15)
            
        )
    },
    {
        "id": "cl04",
        "titulo": (titulo_04_financiamento_fgts()),
        "visivel": lambda: tem_financiamento() or tem_fgts(),
        "render": lambda numero=4: (
            render_subclausulas_dinamicas(
                numero_clausula_principal=numero,
                textos=[
                    clausula_04_1_esclarecimentos_financiamento_fgts(),
                    clausula_04__2_qualidade_financiamento_fgts(),
                    clausula_04__3_qualidade_financiamento_fgts(),
                    clausula_04__4_juizo_financiamento_fgts(),
                ],
                tamanho_px=15)
        )
    },
    {
        "id": "cl05",
        "titulo": "DA ENTREGA DAS CHAVES E DAS CONTAS DE CONSUMO",
        "visivel": lambda: True,
        "render": lambda numero=5: (
            render_subclausulas_dinamicas(
                numero_clausula_principal=numero,
                textos=[
                    clausula_05__1_juizo_entrega_chaves(),
                    clausula_05_2_livre_desocupado(),
                    clausula_05_3_condominio(),
                ],
                tamanho_px=15)
        )
    },
    {
        "id": "cl06",
        "titulo": "DAS TRANSFERÊNCIAS JUNTO À PREFEITURA E ÀS EVENTUAIS CONCESSIONÁRIAS DE ÁGUA, ENERGIA E GÁS",
        "visivel": lambda: True,
        "render": lambda numero=6: (
            render_subclausulas_dinamicas(
                numero_clausula_principal=numero,
                textos=[
                    clausula_06_1_transferencia_concessionaria(),
                    clausula_06_1_transferencia_iptu()
                ],
                tamanho_px=15)
        )
    },
    {
        "id": "cl07",
        "titulo": "DO PAGAMENTO DOS HONORÁRIOS DA INTERMEDIADORA",
        "visivel": lambda: True,
        "render": lambda numero=7: (
            render_subclausulas_dinamicas(
                numero_clausula_principal=numero,
                textos=[
                    clausula_07_1_honorarios(),
                    clausula_07_2_honorarios(),
                    clausula_07_3_honorarios(),
                    clausula_07_4_honorarios(),
                    
                ],
                tamanho_px=15)
        )
    },
    {
        "id": "cl08",
        "titulo": "DO PRAZO DE VALIDADE DO INSTRUMENTO À SUA CONCLUSÃO",
        "visivel": lambda: not get("preco_parcelamento_total", "").strip(),
        "render": lambda numero=8: (
            render_subclausulas_dinamicas(
                numero_clausula_principal=numero,
                textos=[
                    clausula_08_1_prazo_conclusao(),
                    clausula_08_2_resilicao_por_prazo(),
                    clausula_08_3_resilicao_por_prazo(),                    
                ],
                tamanho_px=15)
        )
    },
    {
        "id": "cl09",
        "titulo": "DA RESOLUÇÃO CONTRATUAL",
        "visivel": lambda: True,
        "render": lambda numero=9: (
            render_subclausulas_dinamicas(
                numero_clausula_principal=numero,
                textos=[
                    clausula_09_1_resolucao(),
                    clausula_09_2_desist_com_sinal(),
                    clausula_09_3_desist_com_sinal(),
                    clausula_09_4_desist_com_sinal(),
                ],
                tamanho_px=15)
        )
    },
    {
        "id": "cl10",
        "titulo": "DA IRRETRATABILIDADE",
        "visivel": lambda: True,
        "render": lambda numero=10: (
            render_subclausulas_dinamicas(
                numero_clausula_principal=numero,
                textos=[
                    clausula_10_1_irretratabilidade(),
                ],
                tamanho_px=15)
        )
    },
    {
        "id": "cl11",
        "titulo": "DA EVICÇÃO DE DIREITO E VÍCIOS REDIBITÓRIOS",
        "visivel": lambda: True,
        "render": lambda numero=11: (
            render_subclausulas_dinamicas(
                numero_clausula_principal=numero,
                textos=[
                    clausula_11_1_vicios(),
                    clausula_11_2_vicios(),
                ],
                tamanho_px=15)
        )
    },
    {
        "id": "cl12",
        "titulo": clausula_12_titulo_declaracoes(),
        "visivel": lambda: True,
        "render": lambda numero=12: (
            render_subclausulas_dinamicas(
                numero_clausula_principal=numero,
                textos=[
                    clausula_12_1_ficara_bens(),
                    clausula_12_2_ficara_bens(),
                    clausula_12_3_ficara_bens(),
                ],
                tamanho_px=15)
        )
    },
    {
        "id": "cl13",
        "titulo": "DO TÉRMINO DA PRESTAÇÃO DE SERVIÇO DA INTERMEDIADORA",
        "visivel": lambda: True,
        "render": lambda numero=13: (
            render_subclausulas_dinamicas(
                numero_clausula_principal=numero,
                textos=[
                    clausula_13_1_termino_pretacao(),
                    clausula_13_2_termino_pretacao(),
                    Clausula_13_3_responsabilidade_intermediadora(),
                    Clausula_13_4_responsabilidade_intermediadora(),
                    Clausula_13_5_responsabilidade_intermediadora(),
                ],
                tamanho_px=15)
        )
    },
    {
        "id": "cl14",
        "titulo": "DAS DISPOSIÇÕES GERAIS",
        "visivel": lambda: True,
        "render": lambda numero=14: (
            render_subclausulas_dinamicas(
                numero_clausula_principal=numero,
                textos=[
                    clausula_14_1_disposicoes_gerais(),
                    clausula_14_2_procuracao_vendedora(),
                    clausula_14_3_procuracao_compradora(),
                    clausula_14_4_intimacoes(),
                    clausula_14_5_comunicar_endereco(),
                    clausula_14_6_alterar_endereco(),
                ],
                tamanho_px=15)
        )
    },
    {
        "id": "cl15",
        "titulo": "ELEIÇÃO DO FORO",
        "visivel": lambda: True,
        "render": lambda numero=15: (
            render_subclausulas_dinamicas(
                numero_clausula_principal=numero,
                textos=[
                    clausula_15_1_foro(),
                ],
                tamanho_px=15)
        )
    },


]


# ============================================================
# MAIN
# ============================================================
st.title(f"📄 {step()['title']}")

# ============================================================
# TELA 1: INÍCIO
# ============================================================
if step()["id"] == "inicio":
    st.subheader("📝 Dados iniciais do contrato")

    c1, c2, c3 = st.columns([1, 1, 1])

    with c1:
        numero = st.text_input(
            "Número do contrato",
            value=get("contrato__numero", ""),
            key="contrato__numero_input",
            placeholder="Ex.: 1981"
        )
        set_("contrato__numero", numero)

    with c2:
        tipo = st.selectbox(
            "Tipo de contrato",
            ["Compromisso de Venda e Compra de Imóvel", "Cessão de Posse e Direitos sobre Imóvel"],
            index=0 if get("contrato__tipo", "Compra e Venda") == "Compra e Venda" else 1,
            key="contrato__tipo_select",
        )
        set_("contrato__tipo", tipo)

    with c3:
        email = st.text_input(
            "E-mail do solicitante do contrato",
            value=get("contrato__email_solicitante", ""),
            key="contrato__email_solicitante_input",
            placeholder="ex: cliente@cliente.com.br"
        )
        set_("contrato__email_solicitante", email)

# ============================================================
# TELA 2: IMÓVEL
# ============================================================
elif step()["id"] == "imovel":
    st.subheader("🏠 Dados do Imóvel")

    tipos_imovel = [
        "imóvel",
        "apartamento",
        "apartamento (matrícula em área maior)",
        "sobrado",
        "sobrado em condomínio",
        "sobrado em condomínio (matrícula em área maior)",
        "casa",
        "casa em condomínio",
        "casa em condomínio (matrícula em área maior)",
        "terreno",
        "outro",
    ]

    colA, colB = st.columns([1.1, 1.2])

    # ============================================================
    # COLUNA A — ENDEREÇO DO IMÓVEL
    # ============================================================
    with colA:
        render_endereco("imovel__end", "Endereço do imóvel")

    # ============================================================
    # COLUNA B — IDENTIFICAÇÃO + CONDIÇÕES
    # ============================================================
    with colB:
        st.markdown("### 📌 Identificação")

        tipo_imovel = st.selectbox(
            "Tipo do imóvel",
            tipos_imovel,
            index=tipos_imovel.index(get("imovel__tipo", "imóvel"))
            if get("imovel__tipo", "imóvel") in tipos_imovel
            else 0,
            key="imovel__tipo"
        )
        set_("imovel__tipo", tipo_imovel)

        matricula = st.text_input(
            "N.º matrícula",
            value=get("imovel__matricula", ""),
            key="imovel__matricula"
        )
        set_("imovel__matricula", matricula)

        # Cartório ordinal no campo (via callback)
        def cartorio_cb():
            st.session_state["imovel__cartorio"] = mask_ordinal_cartorio(
                st.session_state.get("imovel__cartorio", "")
            )
            set_("imovel__cartorio", st.session_state["imovel__cartorio"])

        if "imovel__cartorio" not in st.session_state:
            st.session_state["imovel__cartorio"] = get("imovel__cartorio", "")

        st.text_input(
            "N.º do cartório",
            key="imovel__cartorio",
            on_change=cartorio_cb,
            placeholder="Ex.: 2"
        )

        # ✅ Autopreenchimento da cidade do cartório com a cidade do imóvel (ViaCEP)
        cidade_auto = st.session_state.get("imovel__end__cidade", "").strip()
        uf_auto = st.session_state.get("imovel__end__uf", "").strip()

        if uf_auto == "SP" and cidade_auto:
            st.session_state["imovel__cidade_cartorio"] = cidade_auto
            set_("imovel__cidade_cartorio", cidade_auto)

        cidade_cartorio = st.text_input(
            "Cidade do cartório",
            value=st.session_state.get("imovel__cidade_cartorio", ""),
            key="imovel__cidade_cartorio"
        )
        set_("imovel__cidade_cartorio", cidade_cartorio)

        contribuinte = st.text_input(
            "Nº do contribuinte",
            value=get("imovel__contribuinte", ""),
            key="imovel__contribuinte"
        )
        set_("imovel__contribuinte", contribuinte)

        # ============================================================
        # ✅ Informações adicionais
        # ============================================================
        st.divider()
        st.markdown("### Informações adicionais")
        
        c1, c2, c3, c4 = st.columns(4)
                
        with c1:
            par_far = st.radio(
                "Imóvel do PAR ou FAR?",
                ["NÃO", "SIM"],
                horizontal=True,
                index=0,
                key="imovel__par_far"
            )
            set_("imovel__par_far", par_far)

        with c2:
            alienado = st.radio(
                "Alienado fiduciariamente?",
                ["NÃO", "SIM"],
                horizontal=True,
                index=0,
                key="imovel__alienado"
            )
            set_("imovel__alienado", alienado)
        
        with c3:
            alugado = st.radio(
                "O imóvel está locado a terceiros?",
                ["NÃO", "SIM"],
                horizontal=True,
                index=0,
                key="imovel__alugado"
            )
            set_("imovel__alugado", alugado)
        
        if alugado == "SIM":
            locacao = st.text_area(
                "O inquilino vai desocupar o imóvel ou a Parte Compradora vai assumir a locação?",
                value=get("imovel__locacao", ""),
                height=140,
                key="imovel__locacao"
            )
            set_("imovel__locacao", locacao)
        else:
            set_("imovel__locacao", "")
            
        with c4:
            ficara_bens = st.radio(
                "Ficará bens no imóvel?",
                ["NÃO", "SIM"],
                horizontal=True,
                index=0,
                key="imovel__ficara_bens"
            )
            set_("imovel__ficara_bens", ficara_bens)
            
        if ficara_bens == "SIM":
            bens = st.text_area(
                "O que ficará no imóvel? (indicar somente os bens - Exemplo.: armário, sofá, etc.)",
                value=get("imovel__bens", ""),
                height=140,
                key="imovel__bens"
            )
            set_("imovel__bens", bens)
        else:
            set_("imovel__bens", "")

    # ============================================================
    # DESCRIÇÃO DO IMÓVEL NA MATRÍCULA
    # ============================================================
    st.divider()

    nao_lancar_descricao = "matrícula em área maior" in (tipo_imovel or "").lower()

    if nao_lancar_descricao:
        st.warning("🟡 Regra aplicada: NÃO lançar descrição do imóvel (matrícula em área maior).")
        set_("imovel__descricao_matricula", "")
    else:
        descricao = st.text_area(
            "📝 Descrição do imóvel na matrícula",
            value=get("imovel__descricao_matricula", ""),
            height=180,
            key="imovel__descricao_matricula"
        )
        set_("imovel__descricao_matricula", descricao)

# ============================================================
# TELA 3: VENDEDORES
# ============================================================
elif step()["id"] == "vendedores":
    st.subheader("👥 Parte Vendedora")

    ensure_min_one_party("vendedores", "vend")
    vendedores = get_list("vendedores")

    c1, c2 = st.columns(2)
    with c1:
        if st.button("➕ Adicionar vendedor"):
            add_party("vendedores", "vend")
            st.rerun()
    with c2:
        if st.button("🗑️ Remover último vendedor", disabled=(len(vendedores) <= 1)):
            remove_last_party("vendedores")
            st.rerun()

    st.divider()
    for i, pfx in enumerate(vendedores, start=1):
        with st.expander(f"Parte Vendedora {i}", expanded=(i == 1)):
            render_parte(pfx, f"PARTE VENDEDORA {i}")

# ============================================================
# TELA 4: COMPRADORES
# ============================================================
elif step()["id"] == "compradores":
    st.subheader("👥 Parte Compradora")

    ensure_min_one_party("compradores", "comp")
    compradores = get_list("compradores")

    c1, c2 = st.columns(2)
    with c1:
        if st.button("➕ Adicionar comprador"):
            add_party("compradores", "comp")
            st.rerun()
    with c2:
        if st.button("🗑️ Remover último comprador", disabled=(len(compradores) <= 1)):
            remove_last_party("compradores")
            st.rerun()

    st.divider()
    for i, pfx in enumerate(compradores, start=1):
        with st.expander(f"Parte Compradora {i}", expanded=(i == 1)):
            render_parte(pfx, f"PARTE COMPRADORA {i}")

# ============================================================
# TELA 5: PREÇO E CHAVES
# ============================================================
elif step()["id"] == "preco_chaves":
    st.subheader("💰 Preço / Chaves / Comissão")
    st.caption("Preencha a composição do preço. Os valores serão formatados automaticamente.")

    # ==========================================================
    # FUNÇÃO AUXILIAR PARA INPUT DE DINHEIRO COM MÁSCARA
    # ==========================================================
    def money_input(label: str, key: str, placeholder="R$ 0,00"):
        if key not in st.session_state:
            st.session_state[key] = get(key, "")

        def _cb():
            st.session_state[key] = mask_money_br(st.session_state.get(key, ""))
            set_(key, st.session_state[key])

        st.text_input(label, key=key, on_change=_cb, placeholder=placeholder)
        set_(key, st.session_state.get(key, ""))

        return st.session_state.get(key, "")

    colL, colR = st.columns([1.1, 1.0])

    # ==========================================================
    # COLUNA ESQUERDA — COMPOSIÇÃO DO PREÇO
    # ==========================================================
    with colL:
        st.markdown("### 🧾 Composição do Preço")

        preco_total = money_input("PREÇO TOTAL", "preco_total")

        financiamento = money_input("🏦 FINANCIAMENTO", "preco_financiamento")
        fgts = money_input("📌 FGTS", "preco_fgts")
        entrada = money_input("💵 ENTRADA", "preco_entrada")
        sinal = money_input("✍️ SINAL", "preco_sinal")
        recurso_proprio = money_input("👤 RECURSO PRÓPRIO", "preco_recurso_proprio")
        carta_credito = money_input("📄 CARTA DE CRÉDITO", "preco_carta_credito")
        subsidio = money_input("🎯 SUBSÍDIO", "preco_subsidio")

        # Parcelamento Total (valor total parcelado)
        parc_total = money_input("🧾 PARCELAMENTO (VALOR TOTAL PARCELADO)", "preco_parcelamento_total")

        outros = money_input("➕ OUTROS (valor total)", "preco_outros")
        outros_desc = st.text_area("Descreva OUTROS (se houver)", value=get("preco_outros_descricao", ""), height=100, key="preco_outros_descricao")
        set_("preco_outros_descricao", outros_desc)

        st.divider()

        # ==========================================================
        # ATIVA TELAS DETALHADAS
        # ==========================================================
        ativar_parc = st.checkbox(
            "Ativar tela de Parcelamento detalhado",
            value=bool(get("parcelamento_ativado", False) or parc_total.strip()),
            key="parcelamento_ativado_chk"
        )
        set_("parcelamento_ativado", ativar_parc)

        ativar_dacao = st.checkbox(
            "Ativar tela de Permutas / Dação em pagamento",
            value=get("permutas_dacao_ativado", False),
            key="permutas_dacao_chk"
        )
        set_("permutas_dacao_ativado", ativar_dacao)

    # ==========================================================
    # COLUNA DIREITA — CHAVES / COMISSÃO + CORRETORES
    # ==========================================================
    with colR:
        st.markdown("### 🔑 Chaves / Comissão")

        entrega = st.selectbox(
            "Entrega de chaves",
            [
                "30 dias após crédito em conta",
                "30 dias após assinatura no Banco",
                "30 dias após assinatura do CCV",
                "No ato da assinatura no Banco",
                "No ato da assinatura do CCV",
                "24 horas do crédito em conta",
                "Escrever no contrato",
            ],
            key="entrega_chaves"
        )
        set_("entrega_chaves", entrega)

        if entrega == "Escrever no contrato":
            txt = st.text_area(
                "Texto exato para o CCV final",
                value=get("entrega_chaves_texto", ""),
                key="entrega_chaves_texto",
                height=110
            )
            set_("entrega_chaves_texto", txt)
        else:
            set_("entrega_chaves_texto", "")

        quem = st.selectbox(
            "Quem paga a comissão?",
            ["PARTE VENDEDORA", "PARTE COMPRADORA", "AMBAS", "TERCEIRO", "NÃO SE APLICA"],
            key="quem_paga_comissao"
        )
        set_("quem_paga_comissao", quem)

        valor_comissao = money_input("Valor da comissão", "valor_comissao")

        momento = st.selectbox(
            "Momento do pagamento",
            ["NA ESCRITURA", "NA ASSINATURA DO CONTRATO", "NA LIBERAÇÃO DE VALORES NA CONTA DO VENDEDOR"],
            key="momento_pagto"
        )
        set_("momento_pagto", momento)

        st.divider()
        
        # 🔐 botão para abrir ADMIN com senha
        if st.button("🔐 Gerenciar Corretores (senha)", key="btn_admin_corretores"):
            abrir_admin_corretores_com_senha(step_voltar=st.session_state.step_index)
        
        st.markdown("### 👔 Corretores")

        ensure_agents()

        # ----------------------------
        # Corretores de venda
        # ----------------------------
        st.markdown("#### Corretores(as) de Venda")
        corv = get_list("corretores_venda")

        # ✅ Primeiro mostra os corretores
        for i, pfx in enumerate(corv, start=1):
            render_agente(pfx, f"Corretor de venda {i}", "30")

        # ✅ Agora os botões ficam embaixo
        colA, colB = st.columns(2)

        with colA:
            if st.button("➕ Adicionar mais um(a) corretor(a) de venda", key="add_corv"):
                corv.append(f"corv{len(corv)+1:02d}")
                set_list("corretores_venda", corv)
                st.rerun()

        with colB:
            if st.button("🗑️ Remover último corretor de venda", disabled=(len(corv) <= 1), key="rem_corv"):
                corv.pop()
                set_list("corretores_venda", corv)
                st.rerun()


        # ----------------------------
        # Corretores de captação
        # ----------------------------
        st.markdown("#### Corretores(as) de Captação")
        corc = get_list("corretores_captacao")

        # ✅ Primeiro mostra os corretores
        for i, pfx in enumerate(corc, start=1):
            render_agente(pfx, f"Corretor de captação {i}", "15")

        # ✅ Botões embaixo
        colA, colB = st.columns(2)

        with colA:
            if st.button("➕ Adicionar mais um(a) corretor(a) de captação", key="add_corc"):
                corc.append(f"corc{len(corc)+1:02d}")
                set_list("corretores_captacao", corc)
                st.rerun()

        with colB:
            if st.button("🗑️ Remover último corretor de captação", disabled=(len(corc) <= 1), key="rem_corc"):
                corc.pop()
                set_list("corretores_captacao", corc)
                st.rerun()


# ============================================================
# TELA EXTRA: CADASTRO DE CORRETOR (oculta)
# ============================================================
elif step()["id"] == "cadastro_corretor":
    st.subheader("🧑‍💼 Cadastro de Corretor")

    nome = st.text_input("Nome completo", value=get("novo_corretor_nome", ""), key="novo_corretor_nome")
    set_("novo_corretor_nome", nome)

    if "novo_corretor_cpf" not in st.session_state:
        st.session_state["novo_corretor_cpf"] = get("novo_corretor_cpf", "")

    st.text_input(
        "CPF",
        key="novo_corretor_cpf",
        on_change=lambda: cpf_callback_key("novo_corretor_cpf"),
        placeholder="000.000.000-00"
    )
    set_("novo_corretor_cpf", st.session_state["novo_corretor_cpf"])

    st.divider()
    st.markdown("### 💳 Dados bancários")

    banco = st.text_input("Banco", value=get("novo_corretor_banco", ""), key="novo_corretor_banco")
    agencia = st.text_input("Agência", value=get("novo_corretor_agencia", ""), key="novo_corretor_agencia")
    conta = st.text_input("Conta", value=get("novo_corretor_conta", ""), key="novo_corretor_conta")
    pix = st.text_input("Chave PIX", value=get("novo_corretor_pix", ""), key="novo_corretor_pix")

    set_("novo_corretor_banco", banco)
    set_("novo_corretor_agencia", agencia)
    set_("novo_corretor_conta", conta)
    set_("novo_corretor_pix", pix)

    st.divider()

    col1, col2 = st.columns(2)

    with col1:
        if st.button("✅ Concluir cadastro"):
            if nome.strip():

                novo_id = adicionar_corretor_completo(
                    nome=nome.strip(),
                    cpf=get("novo_corretor_cpf", ""),
                    banco=banco.strip(),
                    agencia=agencia.strip(),
                    conta=conta.strip(),
                    pix=pix.strip()
                )

                # define automaticamente no agente que chamou
                prefix = get("cadastro_corretor_prefix", "")
                if prefix:
                    set_(f"{prefix}__nome", nome)
                    st.session_state[f"{prefix}__nome"] = nome
                    st.session_state[f"{prefix}__select"] = nome

                    # salva os dados completos no agente
                    set_(f"{prefix}__cpf", get("novo_corretor_cpf", ""))
                    set_(f"{prefix}__banco", banco)
                    set_(f"{prefix}__agencia", agencia)
                    set_(f"{prefix}__conta", conta)
                    set_(f"{prefix}__pix", pix)

                voltar_para_preco_chaves()
            else:
                st.error("⚠️ Informe o nome completo do corretor.")

    with col2:
        if st.button("⬅️ Voltar sem cadastrar"):
            voltar_para_preco_chaves()


# ============================================================
# TELA 6: PARCELAMENTO (detalhado)
# ============================================================
elif step()["id"] == "parcelamento":
    st.subheader("📆 Parcelamento (Detalhado)")
    desc = st.text_area("Descreva o parcelamento (parcelas, datas, forma)", value=get("parcelamento_descricao", ""), height=220, key="parcelamento_descricao")
    set_("parcelamento_descricao", desc)

# ============================================================
# TELA 7: PERMUTAS / DAÇÃO (detalhado)
# ============================================================
elif step()["id"] == "permutas_dacao":
    st.subheader("🔁 Permutas / Dação em Pagamento (Detalhado)")

    d_veic = st.selectbox("Há dação em VEÍCULO?", ["NÃO", "SIM"], key="dacao_veiculo")
    set_("dacao_veiculo", d_veic)

    d_imov = st.selectbox("Há dação em IMÓVEL?", ["NÃO", "SIM"], key="dacao_imovel")
    set_("dacao_imovel", d_imov)

    if d_imov == "SIM":
        render_endereco("dacao_imovel__end", "Imóvel dado em pagamento")

    if d_imov == "SIM" or d_veic == "SIM":
        desc = st.text_area("Descreva a dação/permutas (bem, valor, condições)", value=get("dacao_descricao", ""), height=220, key="dacao_descricao")
        set_("dacao_descricao", desc)
    else:
        set_("dacao_descricao", "")

# ============================================================
# TELA OCULTA: LOGIN (Admin / Imobiliárias)
# ============================================================
elif step()["id"] == "senha_admin":
    st.subheader("🔐 Acesso restrito")
    st.info("Informe usuário e senha para acessar as áreas restritas.")

    usuario = st.text_input("Usuário", key="auth_usuario")
    senha = st.text_input("Senha", type="password", key="auth_senha")

    col1, col2 = st.columns(2)

    with col1:
        if st.button("✅ Entrar", key="btn_auth_entrar"):
            if validar_login(usuario, senha):
                # salva o usuário logado (imobiliária)
                st.session_state["auth_user"] = usuario.strip()

                # libera admin (se for admin) e também libera as telas restritas
                st.session_state.admin_liberado = (usuario.strip() == "admin")
                st.session_state.admin_corretores_liberado = True

                destino = get("destino_admin", "admin_corretores")
                if destino == "admin_clausulas":
                    abrir_admin_clausulas()
                else:
                    abrir_admin_corretores()
            else:
                st.error("❌ Usuário ou senha incorretos.")

    with col2:
        if st.button("⬅️ Voltar", key="btn_auth_voltar"):
            go_to_step("preco_chaves")  # mantém seu fluxo atual
            st.rerun()

# ============================================================
# TELA OCULTA: ADMIN CORRETORES (LISTA / EDITAR / EXCLUIR)
# ============================================================
elif step()["id"] == "admin_corretores":

    if not st.session_state.get("admin_corretores_liberado", False):
        st.error("⛔ Acesso negado.")
        if st.button("⬅️ Voltar"):
            voltar_da_admin_para_origem()
        st.stop()

    st.subheader("🧑‍💼 Corretores Cadastrados (Admin)")

    base = st.session_state.dados.get("corretores_cadastrados", [])

    if len(base) == 0:
        st.warning("Nenhum corretor cadastrado ainda.")

    st.divider()

    for idx, cor in enumerate(base):
        with st.expander(f"👤 {cor.get('nome','(sem nome)')}", expanded=False):

            nome = st.text_input("Nome", value=cor.get("nome",""), key=f"adm_nome_{idx}")
            cpf = st.text_input("CPF", value=cor.get("cpf",""), key=f"adm_cpf_{idx}")
            banco = st.text_input("Banco", value=cor.get("banco",""), key=f"adm_banco_{idx}")
            agencia = st.text_input("Agência", value=cor.get("agencia",""), key=f"adm_agencia_{idx}")
            conta = st.text_input("Conta", value=cor.get("conta",""), key=f"adm_conta_{idx}")
            pix = st.text_input("PIX", value=cor.get("pix",""), key=f"adm_pix_{idx}")

            colA, colB = st.columns(2)

            with colA:
                if st.button("💾 Salvar alterações", key=f"adm_save_{idx}"):
                    corretor_id = base[idx].get("id","")
                
                    # grava no Supabase
                    salvar_corretor_supabase(
                        nome=nome, cpf=cpf, banco=banco, agencia=agencia, conta=conta, pix=pix, corretor_id=corretor_id
                    )
                
                    # recarrega lista
                    _carregar_corretores_supabase()
                    st.success("✅ Alterações salvas.")
                    st.rerun()


            with colB:
                if st.button("🗑️ Excluir corretor", key=f"adm_del_{idx}"):
                    corretor_id = base[idx].get("id","")
                    ok = excluir_corretor_supabase(corretor_id)
                
                    _carregar_corretores_supabase()
                    if ok:
                        st.warning("Corretor excluído.")
                    else:
                        st.error("Não foi possível excluir no Supabase (verifique se existe coluna id e permissões).")
                    st.rerun()
    
    col1 = st.columns(1)

    if st.button("⬅️ Voltar", key="btn_admin_voltar"):
        go_to_step("preco_chaves")
        st.rerun()

# ============================================================
# TELA: CLÁUSULAS (VISUALIZAÇÃO - ENTREGA DE CHAVES)
# ============================================================
elif step()["id"] == "clausulas":

    tipo_contrato = get("contrato__tipo", "").strip()

    if tipo_contrato:
        st.markdown(
            f"<h3 style='text-align:center; text-transform:uppercase;'>{tipo_contrato}</h3>",
            unsafe_allow_html=True
        )

    st.subheader("📄 Pré-visualização do contrato final")


    colS1, colS2 = st.columns([1, 3])
    
    with colS2:
         
        # Filtra só cláusulas visíveis
        clausulas_visiveis = [c for c in CLAUSULAS if c["visivel"]()]

    st.divider()

    # ============================================================
    # QUADRO RESUMO / DAS PARTES
    # ============================================================

    st.markdown("<br>", unsafe_allow_html=True)

    texto_centralizado("QUADRO RESUMO", tamanho_px=18, negrito=True)

    st.markdown("<br>", unsafe_allow_html=True)

    # ✅ DAS PARTES (título)
    st.markdown("### DAS PARTES")

    # ✅ frase variável: PARTE VENDEDORA ou PARTE CEDENTE
    st.markdown(f"<div style='text-align:justify; font-size:15px; line-height:1.6;'>{frase_adiante_designado()}</div>", unsafe_allow_html=True)

        # ✅ QUALIFICAÇÃO PARTE VENDEDORA/CEDENTE (com borda externa)
    qualificacao_v = bloco_qualificacao_vendedores()

    if qualificacao_v:
        box_texto_justificado(qualificacao_v, tamanho_px=15)
    else:
        st.warning("Nenhuma PARTE VENDEDORA/CEDENTE cadastrada na etapa 'Parte Vendedora'.")

    st.markdown("<br>", unsafe_allow_html=True)

    # ✅ FRASE VARIÁVEL PARTE COMPRADORA/CESSIONÁRIA
    st.markdown(
        f"<div style='text-align:justify; font-size:15px; line-height:1.6;'>{frase_adiante_designado_compradora()}</div>",
        unsafe_allow_html=True
    )

    # ✅ QUALIFICAÇÃO PARTE COMPRADORA/CESSIONÁRIA (com borda externa)
    qualificacao_c = bloco_qualificacao_compradores()

    if qualificacao_c:
        box_texto_justificado(qualificacao_c, tamanho_px=15)
    else:
        st.warning("Nenhuma PARTE COMPRADORA/CESSIONÁRIA cadastrada na etapa 'Parte Compradora'.")
        st.markdown("<br>", unsafe_allow_html=True)

    # ============================================================
    # DA INTERMEDIADORA (FIXO)
    # ============================================================

    st.markdown("### DA INTERMEDIADORA")

    st.markdown(
        "<div style='text-align:justify; font-size:15px; line-height:1.6;'>"
        "Adiante simplesmente designado como <b>INTERMEDIADORA</b>:"
        "</div>",
        unsafe_allow_html=True
    )

    texto_intermediadora = bloco_intermediadora()

    if texto_intermediadora:
        box_texto_justificado(texto_intermediadora, tamanho_px=15)
    else:
        st.warning("Texto da intermediadora não definido.")

    # ============================
    # DO OBJETO DO CONTRATO (FIXO + DADOS DO IMÓVEL)
    # ============================

    st.markdown("### DO OBJETO DO CONTRATO")

    st.markdown(
        "<div style='text-align:justify; font-size:15px; line-height:1.6;'>"
        "Adiante simplesmente designado como <b>IMÓVEL</b>:"
        "</div>",
        unsafe_allow_html=True
    )

    dados_objeto = bloco_objeto()

    texto_objeto_do_contrato = dados_objeto.get("objeto", "")
    secoes_separadas = dados_objeto.get("secoes", {})

    # ============================
    # DO OBJETO DO CONTRATO (um box)
    # ============================
    if texto_objeto_do_contrato:
        box_texto_justificado(texto_objeto_do_contrato, tamanho_px=15)
    else:
        st.warning("Texto do OBJETO DO CONTRATO não definido.")

    st.markdown("<br>", unsafe_allow_html=True)

    # ============================
    # OUTRAS SEÇÕES (boxes separados)
    # ============================
    for titulo, conteudo in secoes_separadas.items():
        st.markdown(f"### {titulo}")
        box_texto_justificado(conteudo, tamanho_px=15)
        st.markdown("<br>", unsafe_allow_html=True)

    # ============================
    # 7) TÍTULO DAS CLÁUSULAS E CONDIÇÕES (FIXO)
    # ============================
    st.markdown("<br><br>", unsafe_allow_html=True)
    texto_centralizado("DAS CLÁUSULAS E CONDIÇÕES", tamanho_px=15, negrito=True)
    st.markdown("<br>", unsafe_allow_html=True)

    # ✅ PREÂMBULO VARIÁVEL (COM OU SEM FINANCIAMENTO)
    texto_preambulo = clausula_preambulo_clausulas_condicoes()
    texto_justificado(texto_preambulo, tamanho_px=15)
    st.markdown("<br>", unsafe_allow_html=True)

    # ============================================================
    # ✅ CLÁUSULAS DO CONTRATO (CORPO FINAL)
    # ============================================================

    clausulas_visiveis = [c for c in CLAUSULAS if c["visivel"]()]

    for i, c in enumerate(clausulas_visiveis, start=1):

        # ✅ título numerado, aparece sempre
        st.markdown(f"### {i}. {c['titulo']}")
        c["render"](i)


# ============================================================
# NAV BUTTONS (não exibir em telas ocultas)
# ============================================================
def existe_bloqueio_conjuge_na_tela_atual() -> bool:
    """
    Verifica se algum PF desta tela marcou CASADO/UNIÃO ESTÁVEL
    e não preencheu o cônjuge/companheiro(a).
    """
    step_id = step()["id"]

    if step_id == "vendedores":
        lista = get_list("vendedores")
    elif step_id == "compradores":
        lista = get_list("compradores")
    else:
        return False

    for pfx in lista:
        if get(f"{pfx}__tipo", "Pessoa Física") == "Pessoa Física":
            if get(f"{pfx}__bloqueio_avancar", False):
                return True

    return False


bloquear = existe_bloqueio_conjuge_na_tela_atual()

# ============================================================
# FOOTER: BOTÕES DE NAVEGAÇÃO
# ============================================================

col_prev, col_next = st.columns([1, 1])

with col_prev:
    if st.button("⬅️ Voltar", key="btn_footer_voltar", disabled=(st.session_state.step_index == 0)):
        go_prev()
        st.rerun()

with col_next:
    if st.button("Avançar ➡️", key="btn_footer_avancar", disabled=bloquear):
        go_next()
        st.rerun()



def abrir_admin_clausulas():
    st.session_state.step_index = steps().index(next(s for s in steps() if s["id"] == "admin_clausulas"))
    st.rerun()
