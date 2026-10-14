import type { Messages } from '../../types'

const messages: Messages['auctions'] = {
  statuses: {
    watching: 'Em observação',
    bidding: 'Em licitação',
    won: 'Ganho',
    lost: 'Perdido',
    passed: 'Desistido',
  },
  source: {
    autoDetected: 'Detetado automaticamente',
    autoDetectedTitle: 'Este resultado foi detetado automaticamente a partir do fornecedor — não é necessária nenhuma ação.',
    manual: 'Definido manualmente',
    manualTitle: 'Este resultado foi definido por uma alteração manual do estado.',
  },
}

export default messages
