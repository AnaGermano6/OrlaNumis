import type { Messages } from '../../types'

const messages: Messages['trays'] = {
  backToCollection: 'Voltar à coleção',
  loading: 'A carregar os tabuleiros...',
  none: 'Ainda não há tabuleiros configurados.',
  openSettings: 'Abrir definições',
  loadFailed: 'Não foi possível carregar os tabuleiros.',
  noCoins: 'Sem moedas no tabuleiro',
  loadingCoins: 'A carregar as moedas...',
  noDiameters: 'Nenhuma moeda ativa da coleção tem diâmetro conhecido. Adicione o tamanho das moedas para as mostrar no tabuleiro.',
  emptyCollection: 'A sua coleção está vazia ou nenhuma moeda corresponde aos filtros atuais.',
  coinsLoadFailed: 'Não foi possível carregar as moedas do tabuleiro. Verifique a ligação e tente novamente.',
  trayOf: 'Tabuleiro {current} de {total}',
  untitled: 'Sem título',
  columnsLabel: '{name}, {columns} colunas',
  gridLabel: '{name}, {rows} linhas por {columns} colunas, {occupied} de {capacity} ocupadas',
  empty: 'Vazia, linha {row}, coluna {column}',
  coinAt: '{name}, linha {row}, coluna {column}',
}

export default messages
