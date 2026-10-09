import type { Messages } from '../../types'

const messages: Messages['quickAccessPage'] = {
  refreshing: 'A atualizar o Acesso rápido',
  refresh: 'Atualizar o Acesso rápido',
  loadingLabel: 'A carregar o Acesso rápido',
  emptyTitle: 'Ainda não há nada fixado',
  emptyMessage: 'Fixe moedas, conjuntos, lotes de leilão ativos ou eventos manuais do calendário para os ter sempre à mão.',
  pinnedItems: 'Itens fixados',
  pinned: 'Fixado em {date}',
  removing: 'A remover {title} do Acesso rápido',
  unpin: 'Desafixar {title} do Acesso rápido',
  types: {
    coin: 'Moeda',
    coinSet: 'Conjunto',
    auctionLot: 'Lote de leilão',
    calendarEvent: 'Evento do calendário',
  },
  dateNotSet: 'Data não definida',
  wishlistCoin: 'Moeda da lista de desejos',
  collectionCoin: 'Moeda da coleção',
  removed: '{title} removido do Acesso rápido',
  unpinFailed: 'Não foi possível desafixar este item.',
}

export default messages
