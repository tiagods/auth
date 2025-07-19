O que é logout remoto?
Logout remoto é a capacidade de invalidar uma sessão ativa de um usuário de forma programática, mesmo que ele esteja logado em outro dispositivo, navegador ou aplicação. Muito útil em cenários como:
- O usuário percebe um login suspeito e quer encerrar todas as sessões.
- Políticas corporativas exigem logout após período de inatividade.
- O backend precisa revogar acessos por mudança de permissão.

🧩 Como implementar isso?
1. Controle de sessões no backend
- Ao autenticar, armazene a sessão ativa no banco (com ID da sessão, userId, token, dispositivo, timestamps).
- Salve tokens associados a cada sessão, principalmente o refresh token, já que ele é quem prolonga o acesso.
2. Logout local
- Quando o usuário faz logout pelo app, exclua seus tokens da sessão correspondente no backend e também limpe localmente.
3. Logout remoto
- Crie uma rota segura (POST /users/{id}/logout-all) que:
- Exclui todos os refresh tokens ativos para aquele usuário.
- Invalida os access tokens (dependendo do formato, como JWT, isso pode ser feito com uma blacklist ou mudando iat/version).
- Opcional: envie notificações para dispositivos conectados, se quiser uma UX mais dinâmica.
4. Revogação eficaz
- Acesse o Redis, se você estiver cacheando tokens/sessões, e exclua as entradas.
- Atualize o estado do usuário no banco, como forceLogout: true para forçar revalidação em próximas requisições.
