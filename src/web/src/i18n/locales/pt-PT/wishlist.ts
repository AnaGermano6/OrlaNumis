import type { Messages } from '../../types'

const messages: Messages['wishlist'] = {
  checkAvailability: 'Verificar disponibilidade',
  share: 'Partilhar lista',
  searchAlerts: 'Alertas de pesquisa',
  runHistory: 'Histórico de execuções',
  addTo: 'Adicionar à lista de desejos',
  finderAgent: 'Adicionar agente de pesquisa da lista de desejos',
  check: {
    available: '{count} disponíveis',
    unavailable: '{count} indisponíveis',
    unknown: '{count} desconhecidas',
    checked: '{count} verificadas',
  },
  empty: {
    title: 'A sua lista de desejos está vazia',
    body: 'Adicione moedas à sua lista de desejos para acompanhar o que procura',
    searchAi: 'Pesquisar moedas com IA',
  },
  add: {
    fromUrl: 'Adicionar a partir de URL',
    fromUrlHint: 'Importe e reveja um anúncio público de um comerciante.',
    fromImage: 'Adicionar a partir da análise de imagem',
    fromImageHint: 'Fotografe ou carregue uma moeda para identificação.',
    another: 'Adicionar de outra forma',
  },
  leave: {
    title: 'Na lista de desejos',
    message: {
      one: 'Esta moeda estava na sua lista de desejos: {names}. Retirá-la da lista de desejos, agora que já a tem?',
      other: 'Estas {count} moedas da lista de desejos são esta: {names}. Retirá-las da lista de desejos, agora que já a tem?',
    },
    confirm: { one: 'Retirar', other: 'Retirar todas' },
    keep: 'Manter na lista',
    done: { one: 'Retirada da lista de desejos.', other: '{count} moedas retiradas da lista de desejos.' },
    failed: 'Não foi possível retirar a moeda da lista de desejos. Apague-a na página dela, na lista de desejos.',
  },
}

export default messages
