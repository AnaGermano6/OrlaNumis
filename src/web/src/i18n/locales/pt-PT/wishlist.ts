import type { Messages } from '../../types'

const messages: Messages['wishlist'] = {
  checkAvailability: 'Verificar disponibilidade',
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
}

export default messages
