import type { Messages } from '../../types'

const messages: Messages['oidc'] = {
  provider: 'Fornecedor',
  backToAccount: 'Voltar às definições da conta',
  link: {
    loadingTitle: 'A associar fornecedor de início de sessão',
    successTitle: 'Fornecedor associado',
    errorTitle: 'A associação falhou',
    successSubtitle: 'O seu fornecedor externo de início de sessão está agora associado a esta conta.',
    errorSubtitle: 'Não foi possível associar o fornecedor à sua conta.',
    missingProvider: 'O retorno do fornecedor não indicava um fornecedor válido. Volte a iniciar a associação nas Definições da conta.',
    incomplete: 'O retorno do fornecedor estava incompleto. Volte a iniciar a associação nas Definições da conta.',
    linked: 'Identidade OIDC associada.',
    cancelled: 'A associação foi cancelada ou recusada no fornecedor. Pode tentar novamente nas Definições da conta.',
    providerError: 'O fornecedor devolveu um erro antes de a associação terminar. Tente novamente ou peça a um administrador que reveja a configuração do fornecedor.',
    failed: 'Não foi possível associar o fornecedor. Volte a iniciar a associação nas Definições da conta.',
  },
  handshake: 'A concluir a ligação segura com o fornecedor...',
  misconfiguredDetail: 'O fornecedor de início de sessão não está configurado corretamente: {detail}. Peça a um administrador que reveja as definições do fornecedor.',
  misconfigured: 'O fornecedor de início de sessão não está configurado corretamente. Peça a um administrador que reveja as definições do fornecedor.',
  login: {
    loadingTitle: 'A iniciar sessão',
    successTitle: 'Sessão iniciada',
    errorTitle: 'Não foi possível iniciar sessão',
    successSubtitle: 'A sua sessão segura está pronta.',
    errorSubtitle: 'Não foi possível concluir o início de sessão com o fornecedor.',
    continue: 'Continuar para a coleção',
    back: 'Voltar ao início de sessão',
    missingProvider: 'O retorno do fornecedor não indicava um fornecedor válido. Inicie sessão novamente.',
    incomplete: 'O retorno do fornecedor estava incompleto. Inicie sessão novamente.',
    signedIn: 'Sessão iniciada. Continue para a sua coleção.',
    cancelled: 'O início de sessão foi cancelado ou recusado no fornecedor. Pode tentar novamente ou usar a sua palavra-passe local.',
    providerError: 'O fornecedor devolveu um erro antes de o início de sessão terminar. Tente novamente ou peça a um administrador que reveja a configuração do fornecedor.',
    existingAccount: 'Esta conta do fornecedor corresponde a uma conta local existente. Inicie sessão localmente e associe o fornecedor nas Definições da conta.',
    validation: 'Não foi possível validar a resposta do fornecedor. Inicie sessão novamente.',
    failed: 'O início de sessão OIDC falhou. Tente novamente ou use a sua palavra-passe local.',
  },
}

export default messages
