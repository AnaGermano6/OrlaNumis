import type { Messages } from '../../types'

const messages: Messages['timeline'] = {
  title: 'Cronologia da coleção',
  allCoins: 'Todas as moedas',
  allBanknotes: 'Todas as notas',
  loading: 'A carregar a cronologia...',
  empty: {
    title: 'Sem dados de cronologia',
    body: 'Adicione datas de compra às suas moedas para as ver na cronologia.',
    bodyBanknotes: 'Adicione datas de compra às suas notas para as ver na cronologia.',
  },
  coins: 'Moedas',
  banknotes: 'Notas',
  banknoteCount: { one: '{count} nota', other: '{count} notas' },
  yearSpan: 'Intervalo de anos',
}

export default messages
