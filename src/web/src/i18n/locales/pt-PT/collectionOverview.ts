import type { Messages } from '../../types'

const messages: Messages['collectionOverview'] = {
  title: 'Coleção',
  failed: 'Não foi possível carregar alguns números da coleção. Tente de novo daqui a pouco.',
  all: {
    title: 'Toda a coleção',
    items: 'Peças',
    countries: 'Países',
    value: 'Valor total',
  },
  coins: {
    title: 'Moedas',
    open: 'Ver moedas',
    highlights: 'Destaques das moedas',
  },
  banknotes: {
    title: 'Notas de banco',
    open: 'Ver notas',
    highlights: 'Destaques das notas',
  },
  banknoteHighlights: {
    noRarity: 'Ainda nenhuma nota tem raridade.',
    noYear: 'Ainda nenhuma nota tem ano.',
    noValue: 'Ainda nenhuma nota tem valor atual.',
  },
}

export default messages
