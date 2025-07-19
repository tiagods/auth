Bloqueio de Conta: Protegendo contra força bruta
Objetivo: impedir que atacantes façam inúmeras tentativas de senha para invadir contas.
Como implementar:
- Monitoramento de tentativas falhas
- Armazene no Redis ou banco o número de falhas por userId, IP, ou e-mail.
- Limites configuráveis
- Ex: após 5 tentativas falhas em 15 minutos, bloqueie o login por 30 minutos ou exija verificação extra.
- Feedback seguro
- Evite mensagens que revelam se o e-mail existe ou não no sistema — responda de forma genérica.
- Desbloqueio programado ou por ação
- Após período de tempo, ou via link de desbloqueio enviado por e-mail.
- Administração de bloqueios
- Interface (ou endpoint seguro) para administradores revogarem bloqueios manualmente.


Captcha Adaptativo: Balanceando segurança e experiência
Objetivo: aplicar Captcha apenas quando há indícios de uso abusivo, sem atrapalhar usuários legítimos.
Como implementar:
- Gatilhos inteligentes
- Ative Captcha em situações como:
- Múltiplas falhas consecutivas
- IP suspeito ou desconhecido
- Geolocalização incomum
- Dispositivo ou navegador novo
- Integrações comuns
- Google reCAPTCHA, hCaptcha, Friendly Captcha (via frontend)
- Backend valida o token do Captcha antes de prosseguir com autenticação
- Experiência fluida
- Usuários legítimos passam direto sem ver Captcha.
- Usuários suspeitos têm que resolver desafios antes de continuar.
- Logging de eventos de captcha
- Útil para análise de segurança e para melhorar heurísticas.
