RBAC: Role-Based Access Control
Baseado em papéis atribuídos ao usuário.
Cada papel (role) tem permissões definidas, e usuários herdam essas permissões conforme seu papel.
Exemplo:
- Papel admin → acesso completo a recursos
- Papel editor → pode editar, mas não deletar
- Papel viewer → só leitura
  Como implementar:
- Estrutura de dados
  {
  "userId": "123",
  "roles": ["editor"]
  }
- Tabela de permissões
  {
  "editor": ["update_post", "create_post"],
  "viewer": ["read_post"]
  }
- Checagem no middleware ou endpoint
- Verifique se user.roles inclui role com a permissão requerida.
- Vantagens
- Simples, escalável para papéis estáticos.
- Ideal para sistemas com hierarquia clara.

🧬 ABAC: Attribute-Based Access Control
Baseado em atributos do usuário, recurso ou contexto.
Mais dinâmico e flexível que RBAC.
Exemplo:
- Usuário só pode acessar recurso se for do mesmo departamento.
- Permissão depende do horário ou localização.
  Como implementar:
- Atributos do usuário
  {
  "userId": "123",
  "department": "finance",
  "location": "BR"
  }
- Políticas de acesso
  allow if user.department == resource.department
  allow if user.location == "BR" && resource.sensitive == false
- Engine de avaliação
- Pode usar libs como Casbin, OPA (Open Policy Agent) para definir regras declarativas.
- Vantagens
- Alta flexibilidade.
- Permissões contextuais e dinâmicas.

🧠 Qual escolher?
| Critério | RBAC | ABAC |
| Complexidade | Simples | Alta |
| Flexibilidade | Média | Alta |
| Facilidade de manutenção | Boa com poucos papéis | Exige boas práticas de modelagem |
| Escopo de uso | Usuários com funções fixas | Acessos baseados em contexto |



Você pode até combinar os dois modelos! Por exemplo: usar RBAC para papéis fixos e ABAC para regras contextuais extras (como horário de acesso, local, nível de sensibilidade).

````json
{
  "user": {
    "id": "u123",
    "roles": ["editor"],
    "attributes": {
      "department": "marketing",
      "location": "BR",
      "access_level": "standard"
    }
  },
  "resource": {
    "type": "document",
    "owner_department": "marketing",
    "sensitivity": "normal"
  },
  "policies": [
    {
      "id": "policy_1",
      "type": "RBAC",
      "role": "admin",
      "permissions": ["read", "write", "delete"]
    },
    {
      "id": "policy_2",
      "type": "RBAC",
      "role": "editor",
      "permissions": ["read", "write"]
    },
    {
      "id": "policy_3",
      "type": "ABAC",
      "conditions": {
        "user.attributes.department": "resource.owner_department",
        "resource.sensitivity": "normal"
      },
      "permissions": ["read"]
    },
    {
      "id": "policy_4",
      "type": "ABAC",
      "conditions": {
        "user.attributes.location": "BR",
        "user.attributes.access_level": ["standard", "premium"]
      },
      "permissions": ["write"]
    }
  ]
}
````

Como isso funcionaria
- O sistema avalia se o usuário tem role com permissões suficientes (editor → read, write).
- Em paralelo, verifica se os atributos do usuário satisfazem as condições ABAC (ex: só pode escrever se for do Brasil e tiver nível padrão ou premium).
- Esse modelo permite uma avaliação flexível e escalável, ideal para sistemas com regras mais complexas.
  Se quiser, posso te ajudar a transformar esse esquema em algo mais performático com cache em Redis, ou integrar com libs como Casbin ou OPA. Avançamos pro tópico 8 depois? 📡
