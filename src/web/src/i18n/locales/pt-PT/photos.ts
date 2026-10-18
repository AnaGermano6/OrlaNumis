import type { Messages } from '../../types'

const messages: Messages['photos'] = {
  title: 'Fotos reais da moeda',
  chooseHint: 'Escolha uma foto para o anverso e outra para o reverso. Ficam guardadas na moeda, com o crédito do autor nas notas.',
  query: 'Pesquisa de fotos',
  queryPlaceholder: 'ex.: 50 escudos 1968',
  searching: 'A procurar fotos...',
  searchFailed: 'Não foi possível procurar fotos agora.',
  noneFound: 'Nenhuma foto encontrada. Experimente menos palavras ou outras.',
  results: 'Fotos encontradas',
  photo: 'Foto da moeda',
  viewSource: 'Ver origem',
  licenseHint: 'As fotos do Wikimedia Commons e do Numista pertencem aos seus autores e mantêm a respetiva licença.',
  creditObverse: 'Foto do anverso: {credit}',
  creditReverse: 'Foto do reverso: {credit}',
  chosenForCoin: 'Fotos reais a adicionar: {sides}.',
  importFailed: 'A origem recusou a foto ({sides}). Escolha outra foto ou adicione a sua.',
  save: 'Guardar fotos',
  findReal: 'Procurar fotos reais',
  sources: {
    ecb: 'Banco Central Europeu',
    incm: 'Imprensa Nacional-Casa da Moeda',
  },
  swapSides: 'Trocar anverso e reverso',
  swapSidesFailed: 'Não foi possível trocar as imagens. Tente novamente.',
  banknote: {
    title: 'Fotos reais da nota',
    chooseHint: 'Escolha uma foto para a frente e outra para o verso. Ficam guardadas na nota, com o crédito do autor no campo Notas.',
    photo: 'Foto da nota',
  },
}

export default messages
