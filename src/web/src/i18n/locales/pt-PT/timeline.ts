import type { Messages } from '../../types'

const messages: Messages['timeline'] = {
  title: 'Cronologia da coleção',
  allCoins: 'Todas as moedas',
  collectionOnly: 'Só a coleção',
  soldOnly: 'Só vendidas',
  loading: 'A carregar a cronologia...',
  empty: {
    title: 'Sem dados de cronologia',
    body: 'Adicione datas de compra às suas moedas para as ver na cronologia.',
  },
  coins: 'Moedas',
  yearSpan: 'Intervalo de anos',
}

export default messages
