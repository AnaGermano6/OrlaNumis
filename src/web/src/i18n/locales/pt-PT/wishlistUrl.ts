import type { Messages } from '../../types'

const messages: Messages['wishlistUrl'] = {
  listingUrl: 'URL do anúncio público da moeda',
  placeholder: 'https://comerciante.exemplo/moeda',
  analyzing: 'A analisar...',
  analyze: 'Analisar',
  viewExisting: 'Ver a moeda existente na lista de desejos',
  reviewProposal: 'Rever proposta',
  reviewHint: 'Verifique e edite todos os campos antes de adicionar à sua lista de desejos.',
  listedValue: 'Valor anunciado',
  listingStatus: 'Estado do anúncio',
  statuses: {
    available: 'Disponível',
    sold: 'Vendida',
    reserved: 'Reservada',
    withdrawn: 'Retirada',
  },
  obverseLegend: 'Legenda do anverso',
  obverseDescription: 'Descrição do anverso',
  reverseLegend: 'Legenda do reverso',
  reverseDescription: 'Descrição do reverso',
  imagePreviewAlt: 'Pré-visualização da imagem do anúncio',
  attachImage: 'Anexar esta imagem como anverso',
  confirmAdd: 'Confirmar e adicionar',
  fields: {
    ruler: 'Governante ou emissor',
    dateRange: 'Intervalo de datas',
    rarity: 'Raridade',
    references: 'Referências',
  },
  messages: {
    previewFailed: 'A proposta está pronta, mas não foi possível pré-visualizar a imagem do anúncio.',
    duplicate: 'Este anúncio já está na sua lista de desejos.',
    noProposal: 'O anúncio não devolveu nenhuma proposta.',
    ready: 'Proposta pronta para revisão.',
    needsReview: 'O anúncio precisa de mais revisão antes de poder ser adicionado.',
    cancelled: 'Análise cancelada.',
    analyzeFailed: 'Não foi possível analisar este anúncio.',
    addedNoImage: 'Moeda adicionada à lista de desejos, mas não foi possível anexar a imagem opcional.',
    added: 'Moeda adicionada à lista de desejos.',
    addFailed: 'Não foi possível adicionar esta moeda à lista de desejos.',
  },
}

export default messages
