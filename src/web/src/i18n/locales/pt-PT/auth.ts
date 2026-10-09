import type { Messages } from '../../types'

const messages: Messages['auth'] = {
  username: 'Nome de utilizador',
  password: 'Palavra-passe',
  email: 'Email',
  confirmPassword: 'Confirmar palavra-passe',
  login: {
    subtitle: 'Inicie sessão na sua coleção',
    signingIn: 'A iniciar sessão...',
    signIn: 'Iniciar sessão',
    signInBiometrics: 'Iniciar sessão com biometria',
    alternateSignIn: 'Outras formas de iniciar sessão',
    loadingProviders: 'A carregar fornecedores de início de sessão...',
    noAccount: 'Não tem conta?',
    createOne: 'Crie uma',
    invalidCredentials: 'Nome de utilizador ou palavra-passe inválidos',
    noAuthorizationUrl: 'O fornecedor de início de sessão não devolveu um URL de autorização. Peça a um administrador que verifique a configuração do fornecedor.',
    signInWith: 'Iniciar sessão com {provider}',
    providerDisabled: 'Este fornecedor de início de sessão está desativado. Peça a um administrador que verifique as definições do fornecedor.',
    oidcCancelled: 'O início de sessão foi cancelado ou recusado no fornecedor. Pode tentar novamente ou usar a sua palavra-passe local.',
    oidcConflict: 'Esta conta do fornecedor corresponde a uma conta local existente. Inicie sessão localmente e depois associe o fornecedor nas Definições da conta.',
    oidcMisconfigured: 'O fornecedor de início de sessão não está configurado corretamente. Peça a um administrador que teste as definições do fornecedor.',
    oidcValidation: 'Não foi possível validar a resposta do fornecedor. Tente novamente ou peça a um administrador que reveja a configuração do fornecedor.',
    oidcFailed: 'O início de sessão OIDC falhou. Tente novamente ou use a sua palavra-passe local.',
    tooManyAttempts: 'Demasiadas tentativas. Tente novamente mais tarde.',
    retryIn: { one: 'Demasiadas tentativas. Tente novamente mais tarde. Nova tentativa dentro de {count} segundo.', other: 'Demasiadas tentativas. Tente novamente mais tarde. Nova tentativa dentro de {count} segundos.' },
    biometricFailed: 'A autenticação biométrica falhou',
  },
  register: {
    createAccount: 'Criar conta',
    subtitle: 'Comece a acompanhar a sua coleção',
    creating: 'A criar...',
    haveAccount: 'Já tem conta?',
    signIn: 'Iniciar sessão',
    passwordMismatch: 'As palavras-passe não coincidem',
    failed: 'O registo falhou — o nome de utilizador pode já existir',
  },
}

export default messages
