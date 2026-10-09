import type { Messages } from '../../types'

const messages: Messages['agent'] = {
  errors: {
    noBody: 'Resposta sem conteúdo',
    stream: 'Erro no fluxo do agente',
    ended: 'O fluxo terminou inesperadamente',
    failed: 'O fluxo falhou',
    copilotStream: 'Não foi possível abrir o fluxo do Coin Copilot (HTTP {status}).',
  },
}

export default messages
