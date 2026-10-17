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
    noAccount: 'Não tem conta?',
    createOne: 'Crie uma',
    invalidCredentials: 'Nome de utilizador ou palavra-passe inválidos',
    tooManyAttempts: 'Demasiadas tentativas. Tente novamente mais tarde.',
    retryIn: { one: 'Demasiadas tentativas. Tente novamente mais tarde. Nova tentativa dentro de {count} segundo.', other: 'Demasiadas tentativas. Tente novamente mais tarde. Nova tentativa dentro de {count} segundos.' },
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
