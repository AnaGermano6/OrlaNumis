import type { Messages } from '../../types'

const messages: Messages['quickAccessPage'] = {
  refreshing: 'A atualizar o Acesso rápido',
  refresh: 'Atualizar o Acesso rápido',
  loadingLabel: 'A carregar o Acesso rápido',
  emptyTitle: 'Ainda não há nada fixado',
  emptyMessage: 'Fixe moedas e conjuntos para os ter sempre à mão.',
  pinnedItems: 'Itens fixados',
  pinned: 'Fixado em {date}',
  removing: 'A remover {title} do Acesso rápido',
  unpin: 'Desafixar {title} do Acesso rápido',
  types: {
    coin: 'Moeda',
    coinSet: 'Conjunto',
  },
  wishlistCoin: 'Moeda da lista de desejos',
  collectionCoin: 'Moeda da coleção',
  removed: '{title} removido do Acesso rápido',
  unpinFailed: 'Não foi possível desafixar este item.',
}

export default messages
