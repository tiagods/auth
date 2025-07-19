API Key: identificação do cliente
O que é?
- Um token estático gerado para identificar quem está consumindo sua API, geralmente usado por aplicações.
  Quando usar?
- Para apps que não representam usuários diretamente (ex: cron jobs, integrações).
- Autenticação básica de aplicações no seu sistema.
  Características:
- Pode ser armazenada no banco com metadados: appName, dono, escopo, limites.
- Pode ser enviada via header: X-API-Key
- Permite rate limiting por chave, e revogação específica.

🪪 Access Token: autenticação de usuário autenticado
O que é?
- Token temporário (geralmente JWT ou opaque token) que representa um usuário autenticado com permissões específicas.
  Como funciona?
- Após o login, o usuário recebe um access_token + refresh_token.
- Esse access_token é usado para autenticar nas rotas protegidas da API.
  Características:
- Pode incluir claims como sub, exp, scope, role, etc.
- Enviado via Authorization: Bearer {token}
- Vida curta (ex: 15min), ideal para segurança.

🏗️ Fluxo combinado (exemplo)
- Cliente externo (app) usa API Key para se identificar.
- Usuário loga, recebe access_token.
- Requisição inclui ambos:
- X-API-Key: abc123
- Authorization: Bearer ey...
  Isso permite verificar:
- Se o app é permitido.
- Se o usuário está autenticado.
- Se ambos possuem permissões compatíveis.

🚨 Boas práticas
- Revogação de API Keys: dashboard para criador excluir ou rotacionar.
- Scopes/Permissões: limite de acesso por chave (read:user, auth:login).
- Rate limiting por chave: evitar abuso por app.
- Token binding: vincular access_token à chave do app (avançado)


#### Sugestao de implementação
Estrutura sugerida: Tabelas principais
📁 api_keys
| Campo | Tipo | Descrição |
| id | UUID | Identificador único da chave |
| key | string | API Key (gerada, segura, hashada) |
| token | string | API Key (gerada, segura, hashada) |
| client_name | string | Nome da app/serviço cliente |
| owner_user_id | UUID | Dono da chave (opcional) |
| scopes | string[] / JSON | Lista de escopos permitidos (read:user) |
| status | enum | active, revoked, expired |
| rate_limit | int | Requisições permitidas por minuto |
| created_at | timestamp | Data de geração da chave |
| expires_at | timestamp | Validade (opcional) | 

📁 access_tokens
| Campo | Tipo | Descrição |
| id | UUID | Identificador do token |
| token | string | JWT ou opaque token |
| user_id | UUID | Usuário autenticado |
| client_id | UUID | Relacionado à api_keys.id |
| scopes | string[] / JSON | Escopos usados na emissão do token |
| issued_at | timestamp | Quando foi emitido |
| expires_at | timestamp | Validade |
| revoked | boolean | Token foi revogado ou não |


🧪 Validação por escopo (exemplo de lógica)
No seu microserviço Go, ao receber uma requisição com API Key + Access Token, você pode:
- Buscar os escopos permitidos pela api_key
- Buscar os escopos contidos no access_token
- Validar se os escopos do token estão dentro dos escopos da chave
  // Pseudocódigo
  if !apiKeyScopes.includesAll(accessTokenScopes) {
  return 403 // Escopo não permitido
  }


Assim, você consegue garantir que:
- O token foi gerado por uma app autorizada
- O usuário está com permissões compatíveis com a chave da aplicação

Como implementar no backend (ex: em Go + Echo)
1. Salvar IPs autorizados junto com a API Key ou token

{
"api_key": "abc123",
"allowed_ips": ["192.0.2.10", "203.0.113.5"],
"environment": "prod"
}

````go
func ValidateIP(next echo.HandlerFunc) echo.HandlerFunc {
return func(c echo.Context) error {
ip := c.RealIP() // ou use X-Forwarded-For se estiver atrás de proxy
apiKey := c.Request().Header.Get("X-API-Key")
// Buscar a lista de IPs autorizados para essa chave
allowed := GetAllowedIPs(apiKey)
if !contains(allowed, ip) {
return c.JSON(http.StatusForbidden, map[string]string{"error": "IP não autorizado"})
}
return next(c)
}
}
````

3. Restrição por ambiente (ex: header ou variável)
   Exemplo: exigir que requisições venham com X-Env: production
   env := c.Request().Header.Get("X-Env")
   if env != "production" {
   return c.JSON(http.StatusForbidden, map[string]string{"error": "Ambiente inválido"})
   }



Estratégias avançadas
- CIDR blocks: usar intervalos como 192.168.0.0/24
- Fingerprint da requisição: validar IP + User-Agent + assinatura
- Token-bound IP: vincular o token técnico a um IP na hora da geração e validar em tempo real.



Usar CIDR blocks (Classless Inter-Domain Routing) como estratégia de restrição por IP te dá muito mais controle e flexibilidade do que validar IPs isoladamente. É perfeito para aceitar faixas de IPs, como redes corporativas, datacenters, ou instâncias em cloud.

🧠 O que é CIDR?
CIDR define um intervalo de IPs com base em máscara. Por exemplo:
- 192.168.0.0/24 inclui todos os IPs de 192.168.0.0 a 192.168.0.255
- 10.0.0.0/16 inclui 10.0.0.0 até 10.0.255.255
  Isso te permite dizer: "Aceito qualquer requisição que venha da rede X".

Como implementar em Go
Você pode usar o pacote net padrão do Go para isso:
1. Definindo os CIDRs permitidos
   var allowedCIDRs = []string{
   "192.168.0.0/24",
   "10.10.0.0/16",
   }


2. Função para verificar se o IP está contido
   func isIPAllowed(ipStr string, cidrs []string) bool {
   ip := net.ParseIP(ipStr)
   for _, cidr := range cidrs {
   _, subnet, err := net.ParseCIDR(cidr)
   if err != nil {
   continue
   }
   if subnet.Contains(ip) {
   return true
   }
   }
   return false
   }


3. Usando no middleware Echo
   func CIDRMiddleware(cidrs []string) echo.MiddlewareFunc {
   return func(next echo.HandlerFunc) echo.HandlerFunc {
   return func(c echo.Context) error {
   ip := c.RealIP()
   if !isIPAllowed(ip, cidrs) {
   return c.JSON(http.StatusForbidden, map[string]string{"error": "IP fora do intervalo permitido"})
   }
   return next(c)
   }
   }
   }



🔐 Dicas de segurança
- Combine com validação da API Key ou token técnico, para uma camada dupla.
- Evite permitir ranges amplos sem necessidade (0.0.0.0/0 é basicamente tudo).
- Logue as requisições bloqueadas, para detectar tentativas de invasão ou uso indevido.


O que é Token-bound IP?
É o conceito de emitir um access token que só é aceito se usado a partir de um IP previamente vinculado. Se o token for interceptado ou reutilizado de outro IP, ele é invalidado na hora.
Isso reduz riscos como:
- Vazamento de token técnico por erro de configuração.
- Reuso malicioso do token fora do ambiente autorizado.
- Ataques vindos de redes externas simulando requisições válidas.

🧷 Como implementar?
1. Associar IP no momento da emissão
   Ao criar o access token técnico, salve no banco o IP público de onde ele será usado:
   {
   "token": "eyJhbGciOi...",
   "bound_ip": "203.0.113.5",
   "client_id": "key-abc123",
   "scope": ["report:run"],
   "expires_at": null
   }


2. Validar IP em cada requisição
   No middleware, extraia o IP do request:
   ip := c.RealIP()
   token := ExtractAccessToken(c)

// Buscar token no banco ou cache
stored := FindTokenInDB(token)
if stored.BoundIP != ip {
return c.JSON(http.StatusForbidden, map[string]string{"error": "IP não autorizado para este token"})
}


3. Alternativa via claim no JWT
   Você também pode colocar o IP como claim no token:
   {
   "sub": "service-client",
   "scope": ["job:run"],
   "ip": "203.0.113.5",
   "exp": 9999999999
   }


E validar o claim ip contra c.RealIP() na middleware.
⚠️ Lembre-se: tokens JWT são assinados, mas os claims podem ser lidos por qualquer um. Então não confie só neles se o token for armazenado em lugares inseguros.

🎯 Dica prática
Você pode combinar token-bound IP com restrição por CIDR, criando tokens que só funcionam dentro de uma faixa autorizada. Exemplo:
- Gera token válido apenas para 10.0.0.0/24
- Salva essa restrição como campo ip_range
- No middleware, verifica se ip ∈ range

Se quiser, posso te ajudar a montar o payload JWT ideal com ip ou audience, criar uma tabela technical_tokens com IP vinculado, ou escrever o middleware completo em Go para validar isso em tempo real. Bora blindar essa camada com eficiência e estilo? 😄


Sobre JWT Claims: Públicos vs Sensíveis
Os claims em JWT são partes do payload codificado em base64, o que significa que qualquer um com acesso ao token pode ler seu conteúdo, mesmo que não possa modificá-lo sem invalidar a assinatura.
Portanto, é importante diferenciar:
| Tipo de Claim | Exemplo | Seguro para o JWT? | Observação |
| Identificadores | sub, userId, sessionId | ✅ Sim | Devem ser irreversíveis ou genéricos |
| Permissões | roles, scope | ✅ Sim | Não deve conter lógicas críticas |
| Dados sensíveis | email, IP, CPF | ⚠️ Evitar | Preferível manter fora do token |



🔒 Onde colocar cada dado?
✅ Pode estar nos Claims (JWT Payload)
- sub: identificador principal (ex: UUID do usuário)
- sessionId: ID de controle de sessão (opcional e gerado internamente)
- roles: lista de papéis (ex: ["admin", "editor"])
- scope: escopos permitidos (ex: ["user:read"])
- userId: pode ser redundante com sub, mas ok se bem modelado
  🧠 Dica: use nomes curtos e comuns — ajuda na interoperabilidade.

⚠️ Evite em Claims, preferir servidor/cache:
- ip: o IP do cliente deve ser validado diretamente na requisição (ex: via middleware).
- email, CPF, dados pessoais: mantenha fora, busque via userId se necessário.
- permissions muito detalhadas: melhor armazenar no backend ou cache.

🛡️ Estratégia segura recomendada:
- JWT leve: com sub, roles, scope, sessionId, exp.
- Validação complementar via cache:
- Redis: session:{sessionId} → dados da sessão, status, IP, permissões detalhadas
- Middleware robusto: valida assinatura, expiração, escopo, e crosscheck com cache se necessário.

Excelente! Bora destrinchar como essas três abordagens — RBAC + scopes, ABAC, e Permission Matrix — podem ser aplicadas de forma prática, eficiente e segura na tua arquitetura. 🛡️

🎭 RBAC + Scopes
RBAC (Role-Based Access Control) atribui papéis fixos aos usuários (admin, editor, viewer), enquanto os scopes refinam as permissões para ações específicas (user:read, post:delete, etc.).
🔧 Como implementar
- Defina roles
  "roles": ["editor"]
- Mapeie permissões por role
  {
  "editor": ["post:create", "post:update", "post:read"],
  "admin": ["*"]
  }
- Inclua scopes no JWT
  {
  "sub": "user123",
  "roles": ["editor"],
  "scope": ["post:create", "post:read"]
  }
- Valide no middleware
  if !userHasScope("post:delete") {
  return 403
  }


✅ Ideal para:
- Sistemas com papéis estáticos e lógica de acesso modular
- APIs com endpoints bem definidos

🧬 ABAC (Attribute-Based Access Control)
ABAC permite regras baseadas em atributos do usuário, recurso ou ambiente. Ele é mais dinâmico e contextual.
🔧 Como implementar
- Colete atributos relevantes
  "user": {
  "department": "finance",
  "location": "BR"
  },
  "resource": {
  "owner_department": "finance"
  }
- Defina política ABAC
  allow if user.department == resource.owner_department
- Use uma engine como OPA ou Casbin
- Carregue atributos no contexto da requisição
- Avalie política e permita ou negue acesso
  ✅ Ideal para:
- Multi-tenant, regras complexas, segurança baseada em contexto
- Ambientes dinâmicos com dados que mudam com frequência

ABAC usando Casbin em Go — com modelo, políticas, atributos dinâmicos e validação em runtime. Esse exemplo vai avaliar se um usuário pode acessar um relatório, baseado no departamento e nível de acesso. Bora lá! 💼🔐

📦 1. Instale Casbin
go get github.com/casbin/casbin/v2



🧪 2. Modelo ABAC (model.conf)
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub_dept, obj_dept, min_level, act

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub.department == p.sub_dept &&
r.obj.department == p.obj_dept &&
r.sub.level >= p.min_level &&
r.act == p.act


Esse matcher exige:
- Mesmo departamento do usuário e do recurso
- Nível de acesso do usuário ≥ exigido
- Ação compatível (ex: view, edit)

📋 3. Política (policy.csv)
p, finance, finance, 3, view
p, finance, finance, 5, edit
p, marketing, marketing, 2, view
p, marketing, marketing, 4, edit



👤 4. Atributos de entrada em Go
type User struct {
Department string
Level      int
}

type Report struct {
Department string
}



🧪 5. Código Go com validação ABAC
package main

import (
"github.com/casbin/casbin/v2"
"github.com/casbin/casbin/v2/model"
"github.com/casbin/casbin/v2/persist"
"fmt"
)

func main() {
m, _ := model.NewModelFromFile("model.conf")
a := persist.NewFileAdapter("policy.csv")
e, _ := casbin.NewEnforcer(m, a)

    // Usuário atual
    user := User{
        Department: "finance",
        Level:      4,
    }

    // Recurso acessado
    report := Report{
        Department: "finance",
    }

    action := "edit"

    ok, err := e.Enforce(user, report, action)
    if err != nil {
        panic(err)
    }

    if ok {
        fmt.Println("✅ Acesso permitido")
    } else {
        fmt.Println("🚫 Acesso negado")
    }
}



🧩 Resultado
- Se o usuário for do mesmo departamento do relatório (finance) e tiver nível ≥ 5, pode editar.
- Se tiver nível 3, pode visualizar, mas não editar.

🔐 E mais...
Com Casbin ABAC, você pode usar:
- Comparações entre múltiplos atributos (tenant, região, cargo)
- Operações com valores booleans, arrays e maps
- Adaptadores de banco (Postgres, Redis, etc.)
- Middleware com Echo ou Gin para proteger rotas de API
