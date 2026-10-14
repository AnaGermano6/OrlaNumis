import type { Messages } from '../../types'

const messages: Messages['notifications'] = {
  errors: {
    refreshCount: 'Não foi possível atualizar o número de notificações. Tente novamente ou aguarde pela próxima atualização automática.',
  },
  markAllRead: 'Marcar todas como lidas',
  loading: 'A carregar as notificações...',
  empty: 'Ainda não há notificações',
  emptyHint: 'Será notificado sobre pedidos de seguidores, alterações na lista de desejos e novas moedas de utilizadores que segue.',
  runNumber: 'Execução n.º {id}.',
  hints: {
    proposalFailed: 'Experimente um pedido mais restrito por intervalo de datas, casa da moeda, governante ou série e submeta novamente.',
    proposalReady: 'Abra a revisão para analisar as posições propostas, a interpretação do âmbito e as estimativas de correspondência com a coleção antes de aprovar.',
    creationFailed: 'Abra Conjuntos e tente novamente a partir de uma proposta revista, ou gere novamente com instruções mais claras.',
    setNewCoins: 'Abra o conjunto para as ver.',
  },
  setNewCoins: 'Novas moedas para «{set}»',
}

export default messages
